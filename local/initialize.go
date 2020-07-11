package local

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/knoebber/dotfile/file"
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

	if exists(storageDir) {
		return
	}

	if err = os.Mkdir(storageDir, 0755); err != nil {
		err = errors.Wrap(err, "creating storage dir")
	}

	return
}

// NewStorage returns a new storage.
// Dir is the directory for storing dotfile tracking information.
func NewStorage(home, dir string) (*Storage, error) {
	s := new(Storage)

	user, err := GetUserConfig(home)
	if err != nil {
		return nil, err
	}
	s.User = user

	if home == "" {
		return nil, errors.New("home cannot be empty")
	}
	if dir == "" {
		return nil, errors.New("dir cannot be empty")
	}

	s.User = user
	s.Home = home
	s.dir = dir

	return s, nil
}

// InitFile sets up a new file to be tracked.
// It will setup the storage directory if its the first file.
// Closes storage.
func InitFile(home, dir, path, alias string) (string, error) {
	convertedPath, err := convertPath(path, home)
	if err != nil {
		return "", err
	}

	alias, err = file.GetAlias(alias, convertedPath)
	if err != nil {
		return "", err
	}

	s, err := NewStorage(home, dir)
	if err != nil {
		return "", err
	}

	if err := s.LoadFile(alias); err != nil {
		return "", err
	}

	if s.HasFile {
		return "", fmt.Errorf("%#v is already tracked", alias)
	}

	// Example: ~/.local/share/dotfile
	if err := createDir(dir); err != nil {
		return "", nil
	}
	s.FileData.Path = convertedPath

	if err := file.Init(s, convertedPath, alias); err != nil {
		return "", err
	}

	return alias, nil
}
