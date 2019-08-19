package file

// Dotfile tracks files by writing to a json file in the users home direcory.
// This file provides functions for reading and writing data to the file system.
//
// Example generated json:
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

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

// Storage provides methods for reading and writing json data and compressed commit bytes.
// All exported fields should be set.
type Storage struct {
	home string // The path to the users home directory.
	dir  string // The path to the folder where data will be stored.
	name string // The name of the json file.

	path  string
	files map[string]*trackedFile
}

// Sets up the storage directory if it hasn't been done yet and pulls its data.
func (s *Storage) Setup(home, dir, name string) error {
	if home == "" {
		return errors.New("home cannot be empty  ")
	}
	if dir == "" {
		return errors.New("dir cannot be empty")
	}
	if name == "" {
		return errors.New("name cannot be empty")
	}

	s.home = home
	s.dir = dir
	s.name = name

	if s.dir[len(s.dir)-1] != '/' {
		s.dir += "/"
	}

	s.path = fmt.Sprintf("%s%s", s.dir, s.name)

	if err := createIfNotExist(s.dir, s.path); err != nil {
		return err
	}

	return s.get()
}

func (s *Storage) GetHome() string {
	return s.home
}

func (s *Storage) saveCommit(c *commit, alias string, t *trackedFile, bytes []byte) error {
	// Create the directory for the files commits if it doesn't exist
	commitdir := fmt.Sprintf("%s%s", s.dir, alias)

	// The name of the file will be the hash
	commitPath := fmt.Sprintf("%s/%s", commitdir, c.Hash)

	if err := createIfNotExist(commitdir, commitPath); err != nil {
		return err
	}

	if err := ioutil.WriteFile(commitPath, bytes, 0644); err != nil {
		return err
	}

	return s.save(alias, t)
}

// Reads the json and makes the tracked file map.
func (s *Storage) get() error {
	s.files = make(map[string]*trackedFile)

	f, err := os.Open(s.path)
	if err != nil {
		return errors.Wrapf(err, "failed to open %s", s.path)
	}
	defer f.Close()

	// Read the entire file into bytes so it can be unmarshalled.
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	if len(bytes) == 0 {
		return nil
	}

	if err := json.Unmarshal(bytes, &s.files); err != nil {
		return errors.Wrapf(err, "failed to unmarshal %s", s.path)
	}
	return nil
}

// Gets a tracked file by its alias.
func (s *Storage) getTrackedFile(alias string) (*trackedFile, error) {
	if err := s.get(); err != nil {
		return nil, err
	}

	t, ok := s.files[alias]
	if !ok {
		return nil, fmt.Errorf("%#v not tracked, use 'dot init <file>' first", alias)
	}
	return t, nil
}

// Saves the trackedFile map to json.
func (s *Storage) save(alias string, t *trackedFile) error {
	s.files[alias] = t

	json, err := json.MarshalIndent(s.files, "", " ")
	if err != nil {
		return errors.Wrapf(err, "failed to marshal %s", s.path)
	}

	return ioutil.WriteFile(s.path, json, 0644)
}

// Creates a directory and a file.
// If the file already exists nothing happens.
func createIfNotExist(dir, fileName string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
		fmt.Printf("Created %s\n", dir)
	} else if err != nil {
		return errors.Wrapf(err, "failed to create directory %s", dir)
	}

	_, err = os.Stat(fileName)
	if os.IsNotExist(err) {
		f, createErr := os.Create(fileName)
		if createErr != nil {
			return errors.Wrapf(err, "failed to create file %s", dir)
		}
		defer f.Close()

		fmt.Printf("Created %s\n", fileName)
	} else if err != nil {
		return err
	}
	return nil
}
