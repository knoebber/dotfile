package local

import (
	"fmt"
	"github.com/knoebber/dotfile/file"
	"github.com/pkg/errors"
	"path/filepath"
)

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

	if !Exists(s.jsonPath) {
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

	if alias == "" {
		generatedAlias, err := file.PathToAlias(convertedPath)
		if err != nil {
			return "", err
		}
		alias = generatedAlias
	}

	s, err := newStorage(home, dir, alias)
	if err != nil {
		return "", nil
	}

	if Exists(s.jsonPath) {
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
	if err := file.Init(s); err != nil {
		return "", err
	}

	return alias, nil
}
