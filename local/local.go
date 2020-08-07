// Package local tracks files by writing to JSON files in the dotfile directory.
//
// For every new file that is tracked a new .json file is created.
// For each commit on a tracked file, a new file is created with the same name as the hash.
//
// Example: ~/.emacs.d/init.el is added with alias "emacs".
// Supposing Storage.dir is ~/.config/dotfile, then the following files are created:
//
// ~/.config/dotfile/emacs.json
// ~/.config/dotfile/emacs/8f94c7720a648af9cf9dab33e7f297d28b8bf7cd
//
// The emacs.json file would look something like this:
// {
//   "path": "~/.emacs.d/init.el",
//   "revision": "8f94c7720a648af9cf9dab33e7f297d28b8bf7cd"
//   "commits": [{
//     "hash": "8f94c7720a648af9cf9dab33e7f297d28b8bf7cd",
//     "timestamp": 1558896290,
//     "message": "Initial commit"
//   }]
// }
package local

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

const jsonIndent = "  "

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

// GetDefaultStorageDir returns the default location for storing dotfile information.
// Creates the location when it does not exist.
func GetDefaultStorageDir(home string) (storageDir string, err error) {
	localSharePath := filepath.Join(home, ".local/share/")
	if exists(localSharePath) {
		// Priority one : ~/.local/share/dotfile
		storageDir = filepath.Join(localSharePath, "dotfile/")
	} else {
		// Priority two: ~/.dotfile/
		storageDir = filepath.Join(home, ".dotfile/")
	}

	if err = createDir(storageDir); err != nil {
		return
	}

	return
}

// NewStorage returns a new storage.
// Dir is the directory for storing dotfile tracking information.
// Creates dir if it does not exist.
func NewStorage(home, storageDir string) (*Storage, error) {
	if home == "" {
		return nil, errors.New("home cannot be empty")
	}
	if storageDir == "" {
		return nil, errors.New("dir cannot be empty")
	}

	s := new(Storage)

	s.Home = home
	s.dir = storageDir

	// Example: ~/.local/share/dotfile
	if err := createDir(storageDir); err != nil {
		return nil, err
	}

	return s, nil
}
