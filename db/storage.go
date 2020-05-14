package db

import (
	"bytes"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

// Storage implements the file.Storage interface using a sqlite database.
type Storage struct {
	file *File
	tx   *sql.Tx
}

func newStorage() (s *Storage, err error) {
	s = new(Storage)

	s.tx, err = connection.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "starting storage transaction")
	}
	return
}

// LoadFile returns a storage loaded with a users file.
func LoadFile(userID int64, alias string) (*Storage, error) {
	s, err := newStorage()
	if err != nil {
		return nil, err
	}

	s.file, err = getFile(userID, alias)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// InitFile initializes a new file to be tracked.
// Expects that a temp file is setup with the alias first.
func InitFile(userID int64, alias string) (*Storage, error) {
	s, err := newStorage()
	if err != nil {
		return nil, err
	}

	temp, err := getTempFileByAlias(userID, alias)
	if err != nil {
		return nil, err
	}

	s.file, err = temp.save(s.tx)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Close commits the current transaction.
// Must be called after the storage is done being used.
func (s *Storage) Close() error {
	if err := s.tx.Commit(); err != nil {
		return errors.Wrapf(err, "closing transaction for file %d", s.file.ID)
	}
	return nil
}

// HasCommit returns whether the file has a commit with hash.
func (s *Storage) HasCommit(hash string) (exists bool, err error) {
	return hasCommit(s.file.ID, hash)
}

// GetContents reads the bytes from the users temp file.
// Returns an error if the temp file is not set.
func (s *Storage) GetContents() ([]byte, error) {
	temp, err := getTempFileByAlias(s.file.UserID, s.file.Alias)
	if err != nil {
		return nil, err
	}

	return temp.Content, nil
}

// GetRevision pulls a user's revision at hash from the database.
func (s *Storage) GetRevision(hash string) ([]byte, error) {
	return getRevision(s.file.ID, hash)
}

// SaveCommit saves a commit to the database.
func (s *Storage) SaveCommit(buff *bytes.Buffer, hash, message string, timestamp time.Time) error {
	commit := &Commit{
		FileID:    s.file.ID,
		Hash:      hash,
		Message:   message,
		Revision:  buff.Bytes(),
		Timestamp: timestamp,
	}

	if _, err := insert(commit, s.tx); err != nil {
		return errors.Wrapf(err, "inserting commit for file %d", s.file.ID)
	}

	if err := updateRevision(s.tx, s.file.ID, hash); err != nil {
		return err
	}

	return nil
}

// Revert overwrites the files current contents with bytes.
func (s *Storage) Revert(buff *bytes.Buffer, hash string) error {
	return updateContent(s.tx, s.file.ID, buff, hash)
}
