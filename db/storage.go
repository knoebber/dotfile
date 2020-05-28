package db

import (
	"bytes"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

// Storage implements the file.Storage interface using a sqlite database.
type Storage struct {
	Staged *stagedFile
	tx     *sql.Tx
}

// NewStorage returns a storage loaded with a users file information.
// Storage should always be created this way.
func NewStorage(userID int64, alias string) (s *Storage, err error) {
	s = new(Storage)

	s.tx, err = connection.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "starting storage transaction")
	}

	s.Staged, err = setupStagedFile(s.tx, userID, alias)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Close commits the current transaction.
// Must be called after the storage is done being used.
func (s *Storage) Close() error {
	if err := s.tx.Commit(); err != nil {
		return errors.Wrapf(err, "closing transaction for file %d", s.Staged.FileID)
	}
	return nil
}

// HasCommit returns whether the file has a commit with hash.
func (s *Storage) HasCommit(hash string) (exists bool, err error) {
	return hasCommit(s.Staged.FileID, hash)
}

// GetContents returns the bytes from the users temp file.
// Returns an error if the temp file is not set.
func (s *Storage) GetContents() ([]byte, error) {
	if len(s.Staged.DirtyContent) == 0 {
		return nil, errors.New("temp file has not content")
	}

	return s.Staged.DirtyContent, nil
}

// GetRevision returns a commits contents.
func (s *Storage) GetRevision(hash string) ([]byte, error) {
	return getRevision(s.Staged.FileID, hash)
}

// SaveCommit saves a commit to the database.
func (s *Storage) SaveCommit(buff *bytes.Buffer, hash, message string, timestamp time.Time) error {
	commit := &Commit{
		FileID:    s.Staged.FileID,
		Hash:      hash,
		Message:   message,
		Revision:  buff.Bytes(),
		Timestamp: timestamp,
	}

	if _, err := insert(commit, s.tx); err != nil {
		return errors.Wrapf(err, "inserting commit for file %d", s.Staged.FileID)
	}

	return updateContent(s.tx, s.Staged.FileID, s.Staged.DirtyContent, hash)
}

// Revert overwrites the files current contents with bytes.
func (s *Storage) Revert(buff *bytes.Buffer, hash string) error {
	return updateContent(s.tx, s.Staged.FileID, buff.Bytes(), hash)
}
