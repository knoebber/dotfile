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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

// Storage implements the file.Storer interface.
// It represents the local data storage for a file that dot is Tracking.
type Storage struct {
	Home     string // The path to the users home directory.
	Alias    string // A friendly name for the file that is being tracked.
	Tracking *trackedFile

	dir      string // The path to the folder where data will be stored.
	jsonPath string
}

// Load the tracked file.
func (s *Storage) get() error {
	bytes, err := ioutil.ReadFile(s.jsonPath)
	if err != nil {
		return errors.Wrap(err, "reading tracking data")
	}

	if len(bytes) == 0 {
		return fmt.Errorf("%s is empty", s.jsonPath)
	}

	s.Tracking = new(trackedFile)

	if err := json.Unmarshal(bytes, &s.Tracking); err != nil {
		return errors.Wrapf(err, "unmarshaling tracking data")
	}
	return nil
}

// Updates the json file with the updated data from the tracked file.
func (s *Storage) save() error {
	bytes, err := json.MarshalIndent(s.Tracking, "", " ")
	if err != nil {
		return errors.Wrap(err, "marshalling tracking data to json")
	}

	// Example: ~/.config/dotfile/bash_profile.json
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
		return nil, errors.Wrap(err, "reading revision")
	}

	return bytes, nil
}

// GetContents reads the contents of the file that is being tracked.
func (s *Storage) GetContents() ([]byte, error) {
	contents, err := ioutil.ReadFile(fullPath(s.Tracking.Path, s.Home))
	if err != nil {
		return nil, errors.Wrap(err, "reading file contents")
	}

	return contents, nil
}

// SaveCommit saves a commit to the file system.
// Creates a new directory when its the first commit.
// Updates the file's revision field to point to the new hash.
func (s *Storage) SaveCommit(buff *bytes.Buffer, hash, message string, timestamp time.Time) error {
	s.Tracking.Commits = append(s.Tracking.Commits, commit{
		Hash:      hash,
		Message:   message,
		Timestamp: timestamp.Unix(),
	})

	// The directory for the files commits.
	commitDir := filepath.Join(s.dir, s.Alias)

	// Example: ~/.config/dotfile/bash_profile
	if err := createDir(commitDir); err != nil {
		return errors.Wrap(err, "creating directory for revision")
	}

	// Example: ~/.config/dotfile/bash_profile/8f94c7720a648af9cf9dab33e7f297d28b8bf7cd
	commitPath := filepath.Join(commitDir, hash)

	if err := ioutil.WriteFile(commitPath, buff.Bytes(), 0644); err != nil {
		return errors.Wrap(err, "writing revision")
	}

	s.Tracking.Revision = hash
	return s.save()
}

// Revert overwrites a file at path with contents.
func (s *Storage) Revert(buff *bytes.Buffer, hash string) error {
	err := ioutil.WriteFile(fullPath(s.Tracking.Path, s.Home), buff.Bytes(), 0644)
	if err != nil {
		return errors.Wrap(err, "reverting file")
	}

	s.Tracking.Revision = hash
	return s.save()
}

// GetPath gets the full path to the file.
func (s *Storage) GetPath() (string, error) {
	return fullPath(s.Tracking.Path, s.Home), nil
}
