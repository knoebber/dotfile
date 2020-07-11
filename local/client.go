package local

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path"
	"time"

	"github.com/knoebber/dotfile/file"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const timeoutSeconds = 30

type remoteRevision struct {
	hash     string
	revision []byte
}

func getClient() *http.Client {
	return &http.Client{
		Timeout: time.Second * timeoutSeconds,
	}
}

func getRemoteTrackingData(client *http.Client, fileURL string) (*file.TrackingData, error) {
	resp, err := client.Get(fileURL)
	if err != nil {
		return nil, errors.Wrap(err, "sending request for remote tracked file")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching remote tracked file: %s", resp.Status)
	}

	result := new(file.TrackingData)
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, errors.Wrap(err, "decoding remote tracked file")
	}

	return result, nil
}

func getRemoteData(s *Storage) (*http.Client, *file.TrackingData, string, error) {
	client := getClient()

	fileURL := s.User.Remote + path.Join("/api", s.User.Username, s.Alias)

	remoteData, err := getRemoteTrackingData(client, fileURL)
	if err != nil {
		return nil, nil, fileURL, err
	}

	return client, remoteData, fileURL, nil
}

// Fetches revisions at hash from remote server concurrently.
// Returns an error if any fetches fail or are non 200.
// fileURL is the files api end point, E.G https://dotfilehub.com/api/knoebber/bashrc
//
// The errgroup code is based on the following example:
// https://pkg.go.dev/golang.org/x/sync/errgroup?tab=doc#example-Group-Pipeline
func getRemoteRevisions(client *http.Client, fileURL string, hashes []string) ([]*remoteRevision, error) {
	g, ctx := errgroup.WithContext(context.Background())
	resultChan := make(chan *remoteRevision)

	for _, hash := range hashes {
		hash := hash // Otherwise closure will always use the first hash.
		url := fileURL + "/" + hash
		g.Go(func() error {
			resp, err := client.Get(url)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			fmt.Println("GET", url)
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("fetching file content: %s", resp.Status)
			}

			revision, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			rr := &remoteRevision{
				hash:     hash,
				revision: revision,
			}

			select {
			case resultChan <- rr:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
	}

	// Start waiting for the results and close them after it finishes.
	go func() {
		g.Wait()
		close(resultChan)
	}()

	// Process the results.
	result := make([]*remoteRevision, len(hashes))
	index := 0
	for r := range resultChan {
		result[index] = r
		index++
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}

// Push new local revisions to remote using a multipart POST request.
// The first part is fileData JSON the rest are form files with compressed revisions.
func postData(s *Storage, client *http.Client, newHashes []string, url string) error {
	body := new(bytes.Buffer)

	w := multipart.NewWriter(body)
	defer w.Close()

	jsonPart, err := w.CreatePart(textproto.MIMEHeader{"Content-Type": {"application/json"}})
	if err != nil {
		return errors.Wrap(err, "creating JSON part")
	}

	if err := json.NewEncoder(jsonPart).Encode(s.FileData); err != nil {
		return errors.Wrap(err, "encoding json part: %v")
	}

	for _, hash := range newHashes {
		fmt.Println("adding", hash)
		revisionPart, err := w.CreateFormFile("revision", hash)
		if err != nil {
			return errors.Wrap(err, "creating revision part")
		}

		revision, err := s.GetRevision(hash)
		if err != nil {
			return err
		}

		revisionPart.Write(revision)
	}

	contentType := fmt.Sprintf("multipart/mixed; boundary=%s", w.Boundary())
	resp, err := client.Post(url, contentType, body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pushing file content: %s", resp.Status)
	}

	return nil

}
