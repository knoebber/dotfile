package local

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/knoebber/dotfile/file"
	"github.com/knoebber/dotfile/usererr"
	"github.com/pkg/errors"
)

const jsonIndent = "  "

// TrackedFile represents a locally tracked file.
type TrackedFile struct {
	Path     string   `json:"path"`
	Revision string   `json:"revision"`
	Commits  []Commit `json:"commits"`
}

// Commit represents a tracked files revision.
type Commit struct {
	Hash      string `json:"hash"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"` // Unix timestamp in nanoseconds.
}

// Creates a path that is reusable between machines.
// Returns an error when path does not exist.
func convertPath(path, home string) (string, error) {
	var err error

	if !exists(path) {
		return "", fmt.Errorf("%#v not found", path)
	}

	// Get the full path.
	if path[0] != '/' {
		path, err = filepath.Abs(path)
		if err != nil {
			return "", err
		}
	}

	// If the path is not in $HOME then use as is.
	if !strings.Contains(path, home) {
		return path, nil
	}

	return strings.Replace(path, home, "~", 1), nil
}

// Returns whether the file or directory exists.
func exists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return true
}

// Creates a directory if it does not exist.
func createDir(dir string) error {
	if exists(dir) {
		return nil
	}

	return os.Mkdir(dir, 0755)
}

func writeCommit(contents []byte, storageDir string, alias, hash string) error {
	// The directory for the files commits.
	commitDir := filepath.Join(storageDir, alias)

	// Example: ~/.local/share/dotfile/bash_profile
	if err := createDir(commitDir); err != nil {
		return errors.Wrap(err, "creating directory for revision")
	}

	// Example: ~/.local/share/dotfile/bash_profile/8f94c7720a648af9cf9dab33e7f297d28b8bf7cd
	commitPath := filepath.Join(commitDir, hash)

	if err := ioutil.WriteFile(commitPath, contents, 0644); err != nil {
		return errors.Wrap(err, "writing revision")
	}

	return nil
}

// AssertClean returns an error when the tracked file has uncommitted changes.
func AssertClean(s *Storage) error {
	if !s.HasFile {
		return nil
	}

	_, err := file.Diff(s, s.Tracking.Revision, "")
	if errors.Is(err, file.ErrNoChanges) {
		return nil
	} else if err != nil {
		return err
	}

	return usererr.Invalid("File has uncommited changes")
}
