package local

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/knoebber/dotfile/file"
	"github.com/knoebber/dotfile/usererr"
)

type remoteRevision struct {
	hash     string
	revision []byte
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

// Pull retrieves a file's commits from a dotfile server.
// Updates the local file with the new content from remote.
func Pull(s *Storage, cfg *UserConfig, alias string) error {
	var hashesToPull []string

	if err := s.LoadFile(alias); err != nil {
		return err
	}

	if err := AssertClean(s); err != nil {
		return err
	}

	client := getClient()
	fileURL := cfg.Remote + path.Join("/api", cfg.Username, alias)

	remoteTrackedFile, err := getRemoteTrackedFile(client, fileURL)
	if err != nil {
		return err
	}

	s.FileData, hashesToPull, err = file.MergeTrackingData(s.FileData, remoteTrackedFile)
	if err != nil {
		return err
	}

	// If the pulled file is new and a file with the remotes path already exists.
	if !s.HasFile && exists(s.GetPath()) {
		return usererr.Invalid(remoteTrackedFile.Path +
			" already exists and is not tracked by dotfile. Remove the file or initialize it before pulling")
	}

	remoteRevisions, err := getRemoteRevisions(client, fileURL, hashesToPull)
	if err != nil {
		return err
	}

	for _, rr := range remoteRevisions {
		if err := writeCommit(rr.revision, s.dir, s.Alias, rr.hash); err != nil {
			return err
		}
	}

	return file.Checkout(s, s.FileData.Revision)
}
