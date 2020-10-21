package dotfileclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"time"

	"github.com/knoebber/dotfile/dotfile"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const timeoutSeconds = 30

// Revision is the bytes of a revision and its hash.
// Note that the bytes would not hash to Hash - its from the original uncompressed content.
type Revision struct {
	Hash  string
	Bytes []byte
}

// Client contains a http client and the information needed for interacting with the dotfilehub api.
type Client struct {
	Client   *http.Client
	Remote   string
	Username string
	Token    string
}

// New returns a client that is ready to communicate with a dotfilehub server.
func New(remote, username, token string) *Client {
	return &Client{
		Client: &http.Client{
			Timeout: time.Second * timeoutSeconds,
		},
		Remote:   remote,
		Username: username,
		Token:    token,
	}
}

func (c *Client) userURL() string {
	return c.Remote + "/api/v1/user/" + c.Username
}

func (c *Client) fileURL(alias string) string {
	return c.userURL() + "/" + alias
}

func (c *Client) rawFileURL(alias string) string {
	return c.fileURL(alias) + "/raw"
}

func (c *Client) revisionURL(alias, hash string) string {
	return c.fileURL(alias) + "/" + hash
}

// List lists the files that the remote user has saved.
func (c *Client) List(path bool) ([]string, error) {
	var result []string

	url := c.userURL()
	if path {
		url += "?path=true"
	}
	resp, err := c.Client.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "getting file list from %q for %q", c.Remote, c.Username)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getting file list: %v", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(err, "decoding file list")
	}

	return result, nil
}

// TrackingDataBytes returns the tracking data for alias in bytes.
func (c *Client) TrackingDataBytes(alias string) ([]byte, error) {
	resp, err := c.Client.Get(c.fileURL(alias))
	if err != nil {
		return nil, errors.Wrapf(err, "getting remote tracked file from %q for %q %q", c.Remote, c.Username, alias)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching remote tracked file %q: %s", alias, resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}

// TrackingData returns the file tracking data for alias on remote.
func (c *Client) TrackingData(alias string) (*dotfile.TrackingData, error) {
	data, err := c.TrackingDataBytes(alias)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}

	result := new(dotfile.TrackingData)
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, errors.Wrap(err, "unmarshalling remote tracked file")
	}

	return result, nil
}

func (c *Client) revision(alias, hash string) ([]byte, error) {
	url := c.revisionURL(alias, hash)
	resp, err := c.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Println("GET", url)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching file content: %s", resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}

// Content fetches the current content of alias.
func (c *Client) Content(alias string) ([]byte, error) {
	resp, err := c.Client.Get(c.rawFileURL(alias))
	if err != nil {
		return nil, errors.Wrapf(err, "getting raw content from %q for %q %q", c.Remote, c.Username, alias)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching file content: %s", resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}

// Revisions fetches all of the revisions for alias in the hashes argument.
// Returns an error if any fetches fail or are non 200.
func (c *Client) Revisions(alias string, hashes []string) ([]*Revision, error) {
	g := new(errgroup.Group)
	results := make([]*Revision, len(hashes))

	for i, hash := range hashes {
		i, hash := i, hash // https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			r, err := c.revision(alias, hash)
			if err != nil {
				return err
			}

			results[i] = &Revision{
				Hash:  hash,
				Bytes: r,
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

// UploadRevisions uploads revisions to remote using a multipart POST request.
// The first part is the fileData JSON the rest are form files with the revision bytes.
func (c *Client) UploadRevisions(alias string, data *dotfile.TrackingData, revisions []*Revision) error {
	var body bytes.Buffer

	if len(revisions) == 0 {
		fmt.Println("Upto date.")
		return nil
	}

	url := c.fileURL(alias)

	writer := multipart.NewWriter(&body)

	jsonPart, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type": {"application/json"},
	})

	if err != nil {
		return errors.Wrap(err, "creating JSON part")
	}

	if err = json.NewEncoder(jsonPart).Encode(data); err != nil {
		return errors.Wrap(err, "encoding json part: %v")
	}

	for _, r := range revisions {
		revisionPart, err := writer.CreateFormFile("revision", r.Hash)
		if err != nil {
			return errors.Wrap(err, "creating revision part")
		}

		n, err := revisionPart.Write(r.Bytes)
		if err != nil {
			return err
		}

		fmt.Printf("pushing %s (%d bytes)\n", r.Hash, n)
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
	req.SetBasicAuth(c.Username, c.Token)

	resp, err := c.Client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "uploading revisions to %q for %q %q", c.Remote, c.Username, alias)
	}
	if err := resp.Body.Close(); err != nil {
		return fmt.Errorf("closing response body after push: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("uploading file revisions: %s", resp.Status)
	}

	return nil
}
