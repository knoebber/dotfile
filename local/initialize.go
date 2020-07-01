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

	if !exists(storageDir) {
		err = os.Mkdir(storageDir, 0755)
	}

	return
}

func newStorage(home, dir, alias string) (*Storage, error) {
	s := new(Storage)

	if home == "" {
		return nil, errors.New("home cannot be empty")
	}
	if dir == "" {
		return nil, errors.New("dir cannot be empty")
	}
	if alias == "" {
		return nil, errors.New("alias cannot be empty")
	}

	s.Home = home
	s.dir = dir
	s.Alias = alias
	s.jsonPath = filepath.Join(s.dir, s.Alias+".json")

	return s, nil
}

// LoadFile initializes the storage directory and loads.Alias's data.
// Returns error when alias is not tracked.
func LoadFile(home, dir, alias string) (*Storage, error) {
	s, err := newStorage(home, dir, alias)
	if err != nil {
		return nil, err
	}

	if !exists(s.jsonPath) {
		return nil, file.ErrNotTracked(alias)
	}

	if err := s.get(); err != nil {
		return nil, err
	}

	return s, nil
}

// InitFile sets up a new file to be tracked.
// It will setup the storage directory if its the first file.
func InitFile(home, dir, path, alias string) (string, error) {
	convertedPath, err := convertPath(path, home)
	if err != nil {
		return "", err
	}

	alias, err = file.GetAlias(alias, convertedPath)

	s, err := newStorage(home, dir, alias)
	if err != nil {
		return "", nil
	}

	if exists(s.jsonPath) {
		return "", fmt.Errorf("%#v is already tracked", alias)
	}

	// Example: ~/.config/dotfile
	if err := createDir(dir); err != nil {
		return "", nil
	}

	s.Tracking = &trackedFile{
		Path:    convertedPath,
		Commits: []commit{},
	}
	if err := file.Init(s, convertedPath, alias); err != nil {
		return "", err
	}

	return alias, nil
}
