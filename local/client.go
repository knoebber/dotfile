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
	"github.com/knoebber/dotfile/usererror"
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

type dotfilehubClient struct {
	client   *http.Client
	remote   string
	username string
	token    string
}

func newDotfilehubClient(u *UserConfig) *dotfilehubClient {
	return &dotfilehubClient{
		username: u.Username,
		remote:   u.Remote,
		token:    u.Token,
		client: &http.Client{
			Timeout: time.Second * timeoutSeconds,
		},
	}
}

func (dc *dotfilehubClient) fileURL(alias string) string {
	return dc.remote + path.Join("/api", dc.username, alias)
}

func (dc *dotfilehubClient) getRemoteFileList() ([]string, error) {
	var result []string

	resp, err := dc.client.Get(dc.remote + "/api/" + dc.username)
	if err != nil {
		return nil, errors.Wrap(err, "sending request for file list")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getting remote file list: %v", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(err, "decoding  file list")
	}

	return result, nil
}

func (dc *dotfilehubClient) getRemoteTrackingData(alias string) (*file.TrackingData, error) {
	resp, err := dc.client.Get(dc.fileURL(alias))
	if err != nil {
		return nil, errors.Wrap(err, "sending request for remote tracked file")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching remote tracked file %q: %s", alias, resp.Status)
	}

	result := new(file.TrackingData)
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, errors.Wrap(err, "decoding remote tracked file")
	}

	return result, nil
}

func (dc *dotfilehubClient) getRemoteRevision(revisionURL string) ([]byte, error) {
	resp, err := dc.client.Get(revisionURL)
	if err != nil {
		return nil, err
	}

	fmt.Println("GET", revisionURL)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching file content: %s", resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}

// Fetches revisions at hash from remote server concurrently.
// Returns an error if any fetches fail or are non 200.
// fileURL is the files api end point, E.G https://dotfilehub.com/api/knoebber/bashrc
func (dc *dotfilehubClient) getRemoteRevisions(alias string, hashes []string) ([]*remoteRevision, error) {
	fileURL := dc.fileURL(alias)
	g := new(errgroup.Group)
	results := make([]*remoteRevision, len(hashes))

	for i, hash := range hashes {
		i, hash := i, hash // https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			revision, err := dc.getRemoteRevision(fileURL + "/" + hash)
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

// Pushes new local revisions to remote using a multipart POST request.
// The first part is fileData JSON the rest are form files with compressed revisions.
func (dc *dotfilehubClient) postRevisions(s *Storage, newHashes []string) error {
	var body bytes.Buffer
	url := dc.fileURL(s.Alias)

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

	resp, err := dc.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "creating upload request for push")
	}

	if resp.StatusCode == http.StatusBadRequest {
		return usererror.Invalid(resp.Status)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("uploading file revisions: %s", resp.Status)
	}

	return resp.Body.Close()

}
