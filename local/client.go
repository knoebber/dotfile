package local

import (
	"encoding/json"
	"net/http"
	"path"
	"time"

	"github.com/pkg/errors"
)

const timeoutSeconds = 10

func getClient() *http.Client {
	return &http.Client{
		Timeout: time.Second * timeoutSeconds,
	}
}

func getRemoteTrackedFile(client *http.Client, apiPath string) (*TrackedFile, error) {
	req, err := http.NewRequest("GET", apiPath+"/json", nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating request for remote tracked file")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "sending request for remote tracked file")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	result := new(TrackedFile)
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, errors.Wrap(err, "decoding remote tracked file")
	}

	return result, nil
}

// Pull retrieves a file and commits from a remote dotfile server
// and installs the contents to the local filesystem.
func Pull(remote, username, alias string) error {
	client := getClient()
	apiPath := remote + path.Join("/api", username, alias)

	tf, err := getRemoteTrackedFile(client, apiPath)
	if err != nil {
		return err
	}

	println(tf)

	return nil
}
