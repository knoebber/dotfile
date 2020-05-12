package local

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type trackedFile struct {
	Path     string   `json:"path"`
	Revision string   `json:"revision"`
	Commits  []commit `json:"commits"`
}

type commit struct {
	Hash      string `json:"hash"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"` // Unix timestamp in nanoseconds.
}

// Creates a path that is reusable between machines.
// Returns an error when path does not exist.
func convertPath(path, home string) (string, error) {
	var err error

	if !Exists(path) {
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

func fullPath(path, home string) string {
	if path[0] == '/' {
		return path
	}
	return strings.Replace(path, "~", home, 1)
}

// Exists returns whether the file or directory exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return true
}

// Creates a director if it does not exist.
func createDir(dir string) error {
	if Exists(dir) {
		return nil
	}

	return os.Mkdir(dir, 0755)
}
