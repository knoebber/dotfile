package local

import (
	"fmt"
	"path"

	"github.com/knoebber/dotfile/file"
)

// Push pushes a file's commits to a remote dotfile server.
// Updates the remote file with the new content from local.
func Push(s *Storage, cfg *UserConfig, alias string) error {
	var hashesToPush []string

	// TODO similar code to pull;
	// TODO confusing that file load is happening in this package whereas other times its loaded in the cli package.
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

	s.FileData, hashesToPush, err = file.MergeTrackingData(remoteTrackedFile, s.FileData)
	if err != nil {
		return err
	}

	fmt.Println("TODO push", hashesToPush, "to", fileURL)

	return nil
}
