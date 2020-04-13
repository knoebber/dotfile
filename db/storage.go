package db

import (
	"github.com/knoebber/dotfile/file"
)

// Storage implements the file.Storage interface using a sqlite database.
type Storage struct {
	userID int
}

// NewStorage returns a storage specific to a user.
func NewStorage(userID int) *Storage {
	return &Storage{
		userID: userID,
	}
}

// GetContents reads the bytes from the users temp file.
// Returns an error if the temp file is not set.
func (s *Storage) GetContents() (contents []byte, err error) {
	return nil, nil
}

// GetTracked returns a users tracked file.
func (s *Storage) GetTracked(alias string) (*file.Tracked, error) {
	return nil, nil
}

// GetRevision pulls a users revision from the database.
// Result is zlib compressed.
func (s *Storage) GetRevision(string, string) (compressed []byte, err error) {
	return nil, nil
}

// SaveRevision saves a commit to the database.
func (s *Storage) SaveRevision(*file.Tracked, *file.Commit) error {
	return nil
}

// Revert overwrites the files current contents with bytes.
func (s *Storage) Revert([]byte, string) error {
	return nil
}
