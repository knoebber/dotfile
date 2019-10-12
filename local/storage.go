// Package local tracks files by writing to a JSON file in the users home directory.
//
// Example ~/.dotfile/file.json
// {
//   "bashrc": {
//     "path": "~/.bashrc"
//     "current": "451de414632e08c0ca3adf7a473b15f37c1b2e60"
//     "commits": [{
//       "hash":"451de414632e08c0ca3adf7a473b15f37c1b2e60",
//       "timestamp":"1558896245",
// 	 "message": "add aliases for dotfile"
//    }],
//  },
//   "emacs": {
//     "path": "~/.emacs.d/init.el",
//     "commits": [{
//       "hash":"8f94c7720a648af9cf9dab33e7f297d28b8bf7cd",
//       "timestamp":"1558896290",
//       "message": "add bindings for dotfile"
//  }]
// }
package local

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/knoebber/dotfile/file"
	"github.com/pkg/errors"
)

// Storage implements the file.Storer interface.
type Storage struct {
	Home string // The path to the users home directory.
	dir  string // The path to the folder where data will be stored.
	name string // The name of the json file.

	path  string
	files map[string]*file.Tracked
}

// NewStorage initializes the storage directory and loads its data.
func NewStorage(home, dir, name string) (*Storage, error) {
	s := new(Storage)

	if home == "" {
		return nil, errors.New("home cannot be empty")
	}
	if dir == "" {
		return nil, errors.New("dir cannot be empty")
	}
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	s.Home = home
	s.dir = dir
	s.name = name

	if s.dir[len(s.dir)-1] != '/' {
		s.dir += "/"
	}

	s.path = fmt.Sprintf("%s%s", s.dir, s.name)

	if _, err := createIfNotExist(s.dir, s.path); err != nil {
		return nil, err
	}

	if err := s.get(); err != nil {
		return nil, err
	}

	return s, nil
}

// Populate the files map.
func (s *Storage) get() error {
	s.files = make(map[string]*file.Tracked)

	bytes, err := ioutil.ReadFile(s.path)
	if err != nil {
		return errors.Wrap(err, "reading local storage")
	}

	if len(bytes) == 0 {
		return nil
	}

	if err := json.Unmarshal(bytes, &s.files); err != nil {
		return errors.Wrapf(err, "unmarshalling %#v", s.path)
	}
	return nil
}

// Saves storage to local JSON file.
func (s *Storage) save() error {
	bytes, err := json.MarshalIndent(s.files, "", " ")
	if err != nil {
		return errors.Wrap(err, "marshalling files json")
	}

	if err := ioutil.WriteFile(s.path, bytes, 0644); err != nil {
		return errors.Wrap(err, "writing files json")
	}

	return nil
}

// GetRevision implements file.Storer
func (s *Storage) GetRevision(alias, hash string) ([]byte, error) {
	revisionPath := fmt.Sprintf("%s%s/%s", s.dir, alias, hash)

	bytes, err := ioutil.ReadFile(revisionPath)
	if err != nil {
		return nil, errors.Wrap(err, "reading revision")
	}

	return bytes, nil
}

// GetContents implements file.Storer.
func (s *Storage) GetContents(relativePath string) ([]byte, error) {
	return ioutil.ReadFile(fullPath(relativePath, s.Home))
}

// SaveRevision implements file.Storer
func (s *Storage) SaveRevision(tf *file.Tracked, c *file.Commit) (err error) {
	var created bool

	// Create the directory for the files commits if it doesn't exist
	commitDir := fmt.Sprintf("%s%s", s.dir, tf.Alias)

	// The name of the file will be the hash
	commitPath := fmt.Sprintf("%s/%s", commitDir, c.Hash)

	if created, err = createIfNotExist(commitDir, commitPath); err != nil {
		return errors.Wrap(err, "creating directory for commits")
	}

	if !created {
		return errors.New("revision already exists")
	}

	if err = ioutil.WriteFile(commitPath, c.Compressed.Bytes(), 0644); err != nil {
		return errors.Wrap(err, "writing commit bytes")
	}

	s.files[tf.Alias] = tf
	return s.save()
}

// GetTracked implements file.Storer
func (s *Storage) GetTracked(alias string) (*file.Tracked, error) {
	t, ok := s.files[alias]
	if !ok {
		return nil, nil
	}

	t.Alias = alias
	return t, nil
}

// SaveTracked implements file.Storer
func (s *Storage) SaveTracked(tf *file.Tracked) error {
	s.files[tf.Alias] = tf
	return s.save()
}

// Revert implements file.Storer
func (s *Storage) Revert(contents []byte, relativePath string) error {
	path := fullPath(relativePath, s.Home)

	if err := ioutil.WriteFile(path, contents, 0644); err != nil {
		return errors.Wrap(err, "reverting file")
	}

	return nil
}

// GetPath gets the full path to a file from its alias.
func (s *Storage) GetPath(alias string) (string, error) {
	tf, err := s.GetTracked(alias)
	if err != nil {
		return "", err
	}

	return fullPath(tf.RelativePath, s.Home), nil
}
