package db

import (
	"bytes"
	"time"

	"github.com/pkg/errors"
)

// Storage implements the file.Storage interface using a sqlite database.
type Storage struct {
	file *File
}

// NewStorage returns a storage loaded with a users file.
func NewStorage(userID int64, alias string) (*Storage, error) {
	f, err := getFile(userID, alias)
	if err != nil {
		return nil, err
	}

	return &Storage{file: f}, nil
}

// GetContents reads the bytes from the users temp file.
// Returns an error if the temp file is not set.
func (s *Storage) GetContents(path string) ([]byte, error) {
	temp, err := getTempFileByPath(s.file.UserID, path)
	if err != nil {
		return nil, err
	}

	return temp.Content, nil
}

// GetRevision pulls a user's revision at hash from the database.
func (s *Storage) GetRevision(alias, hash string) ([]byte, error) {
	return getRevision(s.file.UserID, alias, hash)
}

// SaveCommit saves a commit to the database.
func (s *Storage) SaveCommit(buff *bytes.Buffer, alias, hash, message string, timestamp time.Time) error {
	tx, err := connection.Begin()
	if err != nil {
		return errors.Wrap(err, "starting transaction for save revision")
	}

	commit := &Commit{
		FileID:    s.file.ID,
		Hash:      hash,
		Message:   message,
		Revision:  buff.Bytes(),
		Timestamp: timestamp,
	}

	if _, err := insert(commit, tx); err != nil {
		return errors.Wrapf(err, "inserting commit for file %d", s.file.ID)
	}

	if err := updateRevision(tx, s.file.ID, hash); err != nil {
		return err
	}

	return tx.Commit()
}

// Revert overwrites the files current contents with bytes.
func (s *Storage) Revert([]byte, string) error {
	return nil
}
