package local

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const timeoutSeconds = 30

func getClient() *http.Client {
	return &http.Client{
		Timeout: time.Second * timeoutSeconds,
	}
}

func getRemoteTrackedFile(client *http.Client, fileURL string) (*TrackedFile, error) {
	resp, err := client.Get(fileURL)
	if err != nil {
		return nil, errors.Wrap(err, "sending request for remote tracked file")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching remote tracked file: %s", resp.Status)
	}

	result := new(TrackedFile)
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, errors.Wrap(err, "decoding remote tracked file")
	}

	return result, nil
}
