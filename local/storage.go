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
	"bytes"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Storage implements the file.Storer interface.
// It represents the local data storage for a file that dot is Tracking.
type Storage struct {
	Home     string       // The path to the users home directory.
	Alias    string       // A friendly name for the file that is being tracked.
	Tracking *TrackedFile // The current file that storage is tracking.
	HasFile  bool         // Whether the storage has a TrackedFile loaded.

	dir      string // The path to the folder where data will be stored.
	jsonPath string
}

// GetJSON returns the tracked files json.
func (s *Storage) GetJSON() ([]byte, error) {
	jsonContent, err := ioutil.ReadFile(s.jsonPath)
	if err != nil {
		return nil, errors.Wrap(err, "reading tracking data")
	}

	return jsonContent, nil
}

// LoadFile sets storage to track alias.
// Loads the tracking data when it exists, otherwise sets an empty TrackedFile.
func (s *Storage) LoadFile(alias string) error {
	if alias == "" {
		return errors.New("alias cannot be empty")
	}
	s.Alias = alias
	s.jsonPath = filepath.Join(s.dir, s.Alias+".json")

	if !exists(s.jsonPath) {
		s.Tracking = new(TrackedFile)
		s.Tracking.Commits = []Commit{}
		s.HasFile = false
		return nil
	}

	s.Tracking = new(TrackedFile)

	jsonContent, err := s.GetJSON()
	if err != nil {
		return nil
	}

	if err = json.Unmarshal(jsonContent, &s.Tracking); err != nil {
		return errors.Wrapf(err, "unmarshaling tracking data")
	}

	s.HasFile = true
	return nil
}

// Close updates the files JSON with s.Tracking.
func (s *Storage) Close() error {
	bytes, err := json.MarshalIndent(s.Tracking, "", jsonIndent)
	if err != nil {
		return errors.Wrap(err, "marshalling tracking data to json")
	}

	// Example: ~/.local/share/dotfile/bash_profile.json
	if err := ioutil.WriteFile(s.jsonPath, bytes, 0644); err != nil {
		return errors.Wrap(err, "saving tracking data")
	}

	return nil
}

// HasCommit return whether the file has a commit with hash.
// This never returns an error; it's present to satisfy a file.Storer requirement.
func (s *Storage) HasCommit(hash string) (exists bool, err error) {
	for _, c := range s.Tracking.Commits {
		if c.Hash == hash {
			return true, nil
		}
	}
	return
}

// GetRevision returns the files state at hash.
// The bytes are zlib compressed - see file/commit.go.
func (s *Storage) GetRevision(hash string) ([]byte, error) {
	revisionPath := filepath.Join(s.dir, s.Alias, hash)

	bytes, err := ioutil.ReadFile(revisionPath)
	if err != nil {
		return nil, errors.Wrapf(err, "reading revision %#v", hash)
	}

	return bytes, nil
}

// GetContents reads the contents of the file that is being tracked.
func (s *Storage) GetContents() ([]byte, error) {
	contents, err := ioutil.ReadFile(s.GetPath())
	if err != nil {
		return nil, errors.Wrap(err, "reading file contents")
	}

	return contents, nil
}

// SaveCommit saves a commit to the file system.
// Creates a new directory when its the first commit.
// Updates the file's revision field to point to the new hash.
func (s *Storage) SaveCommit(buff *bytes.Buffer, hash, message string, timestamp time.Time) error {
	s.Tracking.Commits = append(s.Tracking.Commits, Commit{
		Hash:      hash,
		Message:   message,
		Timestamp: timestamp.Unix(),
	})

	if err := writeCommit(buff.Bytes(), s.dir, s.Alias, hash); err != nil {
		return err
	}

	s.Tracking.Revision = hash
	return nil
}

// Revert overwrites a file at path with contents.
func (s *Storage) Revert(buff *bytes.Buffer, hash string) error {
	err := ioutil.WriteFile(s.GetPath(), buff.Bytes(), 0644)
	if err != nil {
		return errors.Wrap(err, "reverting file")
	}

	s.Tracking.Revision = hash
	return nil
}

// GetPath gets the full path to the file.
// Returns an empty string when path is not set.
func (s *Storage) GetPath() string {
	if s.Tracking.Path == "" {
		return ""
	}

	if s.Tracking.Path[0] == '/' {
		return s.Tracking.Path
	}

	return strings.Replace(s.Tracking.Path, "~", s.Home, 1)
}
