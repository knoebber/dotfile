package local

import (
	"path/filepath"

	"github.com/pkg/errors"
)

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
func NewStorage(home, storageDir, configDir string) (*Storage, error) {
	if home == "" {
		return nil, errors.New("home cannot be empty")
	}
	if storageDir == "" {
		return nil, errors.New("dir cannot be empty")
	}
	if configDir == "" {
		return nil, errors.New("config dir cannot be empty")
	}
	s := new(Storage)

	user, err := GetUserConfig(configDir)
	if err != nil {
		return nil, err
	}

	s.User = user
	s.Home = home
	s.dir = storageDir

	// Example: ~/.local/share/dotfile
	if err := createDir(storageDir); err != nil {
		return nil, err
	}

	return s, nil
}
