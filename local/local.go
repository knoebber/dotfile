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

	"github.com/knoebber/dotfile/file"
	"github.com/pkg/errors"
)

const jsonIndent = "  "

// Creates a path that is reusable between machines.
// Returns an error when path does not exist.
func convertPath(path string) (string, error) {
	var err error

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if !exists(path) {
		return "", fmt.Errorf("%#v not found", path)
	}

	// Get the full path.
	if !filepath.IsAbs(path) {
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
func GetDefaultStorageDir() (storageDir string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

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

// List returns a slice of aliases for all locally tracked files.
// When the file has uncommited changes an asterisks is added to the end.
func List(storageDir string) ([]string, error) {
	// TODO move to local.go => local.List()
	var alias string

	files, err := filepath.Glob(filepath.Join(storageDir, "*.json"))
	if err != nil {
		return nil, err
	}
	result := make([]string, len(files))

	s := &Storage{Dir: storageDir}
	s.FileData = new(file.TrackingData)
	for i, filename := range files {
		parts := strings.Split(filename, "/")
		if len(parts) != 0 {
			alias = parts[len(parts)-1]
		}

		alias = strings.TrimSuffix(alias, ".json")
		s.Alias = alias

		if err := s.SetTrackingData(); err != nil {
			return nil, err
		}

		path, err := s.GetPath()
		if err != nil {
			return nil, err
		}

		if !exists(path) {
			alias += " - removed"
		} else {
			clean, err := file.IsClean(s, s.FileData.Revision)
			if err != nil {
				return nil, err
			}

			if !clean {
				alias += "*"
			}
		}

		result[i] = alias
	}

	return result, nil
}
