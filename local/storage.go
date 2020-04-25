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
	"bytes"
	"encoding/json"
	"io/ioutil"
	"path/filepath"

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

	s.path = filepath.Join(s.dir, s.name)

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
		return errors.Wrapf(err, "unmarshaling %#v to json", s.path)
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

// GetRevision returns the bytes from a stored revision.
func (s *Storage) GetRevision(alias, hash string) ([]byte, error) {
	revisionPath := filepath.Join(s.dir, alias, hash)

	bytes, err := ioutil.ReadFile(revisionPath)
	if err != nil {
		return nil, errors.Wrap(err, "reading revision")
	}

	return bytes, nil
}

// GetContents wraps ioutil.Readfile to implement a file.Storer requirement.
func (s *Storage) GetContents(relativePath string) ([]byte, error) {
	return ioutil.ReadFile(fullPath(relativePath, s.Home))
}

// SaveRevision saves a commit to the file system.
// Creates a new directory when its the first commit.
func (s *Storage) SaveRevision(tf *file.Tracked, buf *bytes.Buffer, hash string) (err error) {
	var created bool

	// Create the directory for the files commits if it doesn't exist
	commitDir := filepath.Join(s.dir, tf.Alias)

	// The name of the file will be the hash
	commitPath := filepath.Join(commitDir, hash)

	if created, err = createIfNotExist(commitDir, commitPath); err != nil {
		return errors.Wrap(err, "creating directory for commits")
	}

	if !created {
		return errors.New("revision already exists")
	}

	if err = ioutil.WriteFile(commitPath, buf.Bytes(), 0644); err != nil {
		return errors.Wrap(err, "writing commit bytes")
	}

	s.files[tf.Alias] = tf
	return s.save()
}

// GetTracked returns a tracked file from an alias.
// Returns nil when alias isn't tracked.
// This never returns an error - it is present to satisfy file.Storer.
func (s *Storage) GetTracked(alias string) (*file.Tracked, error) {
	tf, ok := s.files[alias]
	if !ok {
		return nil, nil
	}

	tf.Alias = alias
	return tf, nil
}

// SaveTracked saves a tracked file to JSON.
// Overwrites the old entry.
func (s *Storage) SaveTracked(tf *file.Tracked) error {
	s.files[tf.Alias] = tf
	return s.save()
}

// Revert overwrites a file at path with contents.
func (s *Storage) Revert(contents []byte, relativePath string) error {
	path := fullPath(relativePath, s.Home)

	if err := ioutil.WriteFile(path, contents, 0644); err != nil {
		return errors.Wrap(err, "reverting file")
	}

	return nil
}

// GetPath gets the full path to a file from its alias.
func (s *Storage) GetPath(alias string) (string, error) {
	tf, err := file.MustGetTracked(s, alias)
	if err != nil {
		return "", err
	}

	return fullPath(tf.RelativePath, s.Home), nil
}
