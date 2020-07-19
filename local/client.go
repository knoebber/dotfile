package local

import (
	"bytes"
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

func getRemoteData(s *Storage, client *http.Client) (*file.TrackingData, string, error) {
	fileURL := s.User.Remote + path.Join("/api", s.User.Username, s.Alias)

	remoteData, err := getRemoteTrackingData(client, fileURL)
	if err != nil {
		return nil, fileURL, err
	}

	return remoteData, fileURL, nil
}

func getRemoteRevision(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	fmt.Println("GET", url)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching file content: %s", resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}

// Fetches revisions at hash from remote server concurrently.
// Returns an error if any fetches fail or are non 200.
// fileURL is the files api end point, E.G https://dotfilehub.com/api/knoebber/bashrc
func getRemoteRevisions(client *http.Client, fileURL string, hashes []string) ([]*remoteRevision, error) {
	g := new(errgroup.Group)
	results := make([]*remoteRevision, len(hashes))

	for i, hash := range hashes {
		i, hash := i, hash // https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			revision, err := getRemoteRevision(client, fileURL+"/"+hash)
			if err != nil {
				return err
			}

			results[i] = &remoteRevision{
				hash:     hash,
				revision: revision,
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

// Push new local revisions to remote using a multipart POST request.
// The first part is fileData JSON the rest are form files with compressed revisions.
func postData(s *Storage, client *http.Client, newHashes []string, url string) error {
	var body bytes.Buffer

	writer := multipart.NewWriter(&body)

	jsonPart, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type": {"application/json"},
	})

	if err != nil {
		return errors.Wrap(err, "creating JSON part")
	}

	if err = json.NewEncoder(jsonPart).Encode(s.FileData); err != nil {
		return errors.Wrap(err, "encoding json part: %v")
	}

	for _, hash := range newHashes {
		revision, err := s.GetRevision(hash)
		if err != nil {
			return err
		}

		revisionPart, err := writer.CreateFormFile("revision", hash)
		if err != nil {
			return errors.Wrap(err, "creating revision part")
		}

		n, err := revisionPart.Write(revision)
		if err != nil {
			return err
		}

		fmt.Printf("pushing %s (%d bytes)\n", hash, n)
	}

	contentType := fmt.Sprintf("multipart/mixed; boundary=%s", writer.Boundary())
	if err = writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", contentType)
	req.SetBasicAuth(s.User.Username, s.User.Token)

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "creating upload request for push")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("uploading file revisions: %s", resp.Status)
	}

	return resp.Body.Close()

}
