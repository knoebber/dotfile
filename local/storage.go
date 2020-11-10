package local

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/knoebber/dotfile/dotfile"
	"github.com/knoebber/dotfile/dotfileclient"
	"github.com/knoebber/dotfile/usererror"
	"github.com/pkg/errors"
)

var (
	// ErrNotTracked is returned when the current alias in storage is not tracked.
	ErrNotTracked = errors.New("file not tracked")
	// ErrNoData is returned when a method expects non nil file data.
	ErrNoData = errors.New("tracking data not loaded")
)

// Storage provides methods for manipulating tracked files on the file system.
type Storage struct {
	Alias    string                // The name of the file that is being tracked.
	Dir      string                // The path to the folder where data will be stored.
	FileData *dotfile.TrackingData // The current file that storage is tracking.
}

func (s *Storage) jsonPath() string {
	return filepath.Join(s.Dir, s.Alias+".json")
}

func (s *Storage) hasSavedData() bool {
	return exists(s.jsonPath())
}

// JSON returns the tracked files json.
func (s *Storage) JSON() ([]byte, error) {
	jsonContent, err := ioutil.ReadFile(s.jsonPath())
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotTracked
	} else if err != nil {
		return nil, errors.Wrap(err, "reading tracking data")
	}

	return jsonContent, nil
}

// SetTrackingData reads the tracking data from the filesystem into FileData.
func (s *Storage) SetTrackingData() error {
	if s.Alias == "" {
		return errors.New("cannot set tracking data: alias is empty")
	}
	if s.Dir == "" {
		return errors.New("cannot set tracking data: dir is empty")
	}

	s.FileData = new(dotfile.TrackingData)

	jsonContent, err := s.JSON()
	if err != nil {
		return err
	}

	if err = json.Unmarshal(jsonContent, s.FileData); err != nil {
		return errors.Wrapf(err, "unmarshaling tracking data")
	}

	return nil
}

func (s *Storage) save() error {
	content, err := json.MarshalIndent(s.FileData, "", jsonIndent)
	if err != nil {
		return errors.Wrap(err, "marshalling tracking data to json")
	}

	// Create the storage directory if it does not yet exist.
	// Example: ~/.local/share/dotfile
	if err := createDir(s.Dir); err != nil {
		return err
	}

	// Example: ~/.local/share/dotfile/bash_profile.json
	if err := ioutil.WriteFile(s.jsonPath(), content, 0644); err != nil {
		return errors.Wrap(err, "saving tracking data")
	}

	return nil
}

// HasCommit return whether the file has a commit with hash.
func (s *Storage) HasCommit(hash string) (exists bool, err error) {
	if s.FileData == nil {
		return false, ErrNoData
	}

	for _, c := range s.FileData.Commits {
		if c.Hash == hash {
			return true, nil
		}
	}
	return
}

// Revision returns the files state at hash.
func (s *Storage) Revision(hash string) ([]byte, error) {
	revisionPath := filepath.Join(s.Dir, s.Alias, hash)

	content, err := ioutil.ReadFile(revisionPath)
	if err != nil {
		return nil, errors.Wrapf(err, "reading revision %q", hash)
	}

	return content, nil
}

// DirtyContent reads the content of the tracked file.
func (s *Storage) DirtyContent() ([]byte, error) {
	path, err := s.Path()
	if err != nil {
		return nil, err
	}

	result, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "reading %q", s.Alias)
	}

	return result, nil
}

// SaveCommit saves a commit to the file system.
// Creates a new directory when its the first commit.
// Updates the file's revision field to point to the new hash.
func (s *Storage) SaveCommit(buff *bytes.Buffer, c *dotfile.Commit) error {
	if s.FileData == nil {
		return ErrNoData
	}

	s.FileData.Commits = append(s.FileData.Commits, *c)
	if err := writeCommit(buff.Bytes(), s.Dir, s.Alias, c.Hash); err != nil {
		return err
	}

	s.FileData.Revision = c.Hash
	return s.save()
}

// Revert writes files with buff and sets it current revision to hash.
func (s *Storage) Revert(buff *bytes.Buffer, hash string) error {
	path, err := s.Path()
	if err != nil {
		return err
	}

	if err := createDirectories(path); err != nil {
		return err
	}

	err = ioutil.WriteFile(path, buff.Bytes(), 0644)
	if err != nil {
		return errors.Wrapf(err, "reverting file %q", s.Alias)
	}

	s.FileData.Revision = hash
	return s.save()
}

// Path gets the full path to the file.
// Utilizes $HOME to convert paths with ~ to absolute.
func (s *Storage) Path() (string, error) {
	if s.FileData == nil {
		return "", ErrNoData
	}
	if s.FileData.Path == "" {
		return "", errors.New("file data is missing path")
	}

	// If the saved path is absolute return it.
	if filepath.IsAbs(s.FileData.Path) {
		return s.FileData.Path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return strings.Replace(s.FileData.Path, "~", home, 1), nil
}

// Push pushes a file's commits to a remote dotfile server.
// Updates the remote file with the new content from local.
func (s *Storage) Push(client *dotfileclient.Client) error {
	var newHashes []string

	if s.FileData == nil {
		return ErrNoData
	}

	remoteData, err := client.TrackingData(s.Alias)
	if err != nil {
		return err
	}

	if remoteData == nil {
		// File isn't yet tracked on remote, push all local revisions.
		for _, c := range s.FileData.Commits {
			newHashes = append(newHashes, c.Hash)
		}
	} else {
		s.FileData, newHashes, err = dotfile.MergeTrackingData(remoteData, s.FileData)
		if err != nil {
			return err
		}
	}
	revisions := make([]*dotfileclient.Revision, len(newHashes))

	for i, hash := range newHashes {
		revision, err := s.Revision(hash)
		if err != nil {
			return err
		}

		revisions[i] = &dotfileclient.Revision{
			Bytes: revision,
			Hash:  hash,
		}
	}

	if err := client.UploadRevisions(s.Alias, s.FileData, revisions); err != nil {
		return err
	}

	return nil
}

// Pull retrieves a file's commits from a dotfile server.
// Updates the local file with the new content from remote.
// FileData does not need to be set; its possible to pull a file that does not yet exist.
func (s *Storage) Pull(client *dotfileclient.Client) error {
	var newHashes []string

	hasSavedData := s.hasSavedData()

	if hasSavedData {
		if err := s.SetTrackingData(); err != nil {
			return err
		}

		clean, err := dotfile.IsClean(s, s.FileData.Revision)
		if err != nil {
			return err
		}

		if !clean {
			return usererror.Invalid("file has uncommitted changes")
		}
	}

	remoteData, err := client.TrackingData(s.Alias)
	if err != nil {
		return err
	}
	if remoteData == nil {
		return fmt.Errorf("%q not found on remote %q", s.Alias, client.Remote)
	}

	s.FileData, newHashes, err = dotfile.MergeTrackingData(s.FileData, remoteData)
	if err != nil {
		return err
	}

	path, err := s.Path()
	if err != nil {
		return err
	}

	// If the pulled file is new and a file with the remotes path already exists.
	if exists(path) && !hasSavedData {
		return usererror.Invalid(remoteData.Path +
			" already exists and is not tracked by dotfile (remove the file or initialize it before pulling)")
	}

	fmt.Printf("pulling %d new revisions for %s\n", len(newHashes), s.FileData.Path)

	revisions, err := client.Revisions(s.Alias, newHashes)
	if err != nil {
		return err
	}

	for _, revision := range revisions {
		if err = writeCommit(revision.Bytes, s.Dir, s.Alias, revision.Hash); err != nil {
			return err
		}
	}

	return dotfile.Checkout(s, s.FileData.Revision)
}

// Move moves the file currently tracked by storage.
func (s *Storage) Move(newPath string, parentDirs bool) error {
	currentPath, err := s.Path()
	if err != nil {
		return err
	}

	if parentDirs {
		if err := createDirectories(newPath); err != nil {
			return err
		}
	}

	if err := os.Rename(currentPath, newPath); err != nil {
		return err
	}

	s.FileData.Path, err = convertPath(newPath)
	if err != nil {
		return err
	}

	return s.save()
}

// Rename changes a files alias.
func (s *Storage) Rename(newAlias string) error {
	if err := dotfile.CheckAlias(newAlias); err != nil {
		return err
	}

	newDir := filepath.Join(s.Dir, newAlias)
	if exists(newDir) {
		return usererror.Invalid(fmt.Sprintf("%q already exists", newAlias))
	}

	err := os.Rename(filepath.Join(s.Dir, s.Alias), newDir)
	if err != nil {
		return err
	}

	jsonPath := s.jsonPath()
	s.Alias = newAlias

	err = os.Rename(jsonPath, s.jsonPath())
	if err != nil {
		return err
	}

	return nil
}

// Forget removes all tracking information for alias.
func (s *Storage) Forget() error {
	if err := os.Remove(s.jsonPath()); err != nil {
		return err
	}

	return os.RemoveAll(filepath.Join(s.Dir, s.Alias))
}

// RemoveCommits removes all commits except for the current.
func (s *Storage) RemoveCommits() error {
	var current dotfile.Commit

	if s.FileData == nil {
		return ErrNoData
	}

	for _, c := range s.FileData.Commits {
		if c.Hash == s.FileData.Revision {
			current = c
			continue
		}
		if err := os.Remove(filepath.Join(s.Dir, s.Alias, c.Hash)); err != nil {
			return err
		}
	}

	if current.Hash != "" {
		s.FileData.Commits = []dotfile.Commit{current}
		return s.save()
	}

	return nil
}

// Remove deletes the file that is tracked and all its data.
func (s *Storage) Remove() error {
	path, err := s.Path()
	if err != nil {
		return err
	}

	if err = os.Remove(path); err != nil {
		return err
	}

	return s.Forget()
}
