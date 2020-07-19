package db

import (
	"bytes"
	"database/sql"

	"github.com/knoebber/dotfile/file"
	"github.com/pkg/errors"
)

// FileTransaction is used for dotfile operations.
// This should be created with one of the exported functions not a literal.
// Any errors will result in rollback to initial state.
type FileTransaction struct {
	tx          *sql.Tx
	newCommitID int64

	FileExists      bool
	FileID          int64
	CurrentCommitID int64
	Hash            string
	Staged          *TempFile
}

// NewFileTransaction loads file information into a file transaction.
// FileID, CurrentCommitID, and Hash will be zero valued when the file doesn't exist.
func NewFileTransaction(username, alias string) (ft *FileTransaction, err error) {
	ft = new(FileTransaction)

	ft.tx, err = connection.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "starting storage transaction")
	}

	row := connection.
		QueryRow(`
SELECT files.id, current_commit_id, hash
FROM files 
JOIN users ON users.id = user_id
JOIN commits ON current_commit_id = commits.id
WHERE username = ? AND alias = ?`, username, alias)

	err = row.Scan(
		&ft.FileID,
		&ft.CurrentCommitID,
		&ft.Hash,
	)

	if NotFound(err) {
		return ft, nil
	}

	if err != nil {
		return nil, ft.Rollback(errors.Wrapf(err, "querying for user %q file %q", username, alias))
	}

	ft.FileExists = true

	return
}

// StageFile returns a file transaction loaded with a users temp file.
// Returns an error when the temp file does not exist.
func StageFile(username string, alias string) (ft *FileTransaction, err error) {
	ft, err = NewFileTransaction(username, alias)
	if err != nil {
		return
	}

	ft.Staged, err = GetTempFile(username, alias)
	if err != nil {
		return nil, ft.Rollback(err)
	}

	if ft.FileExists {
		return ft, nil
	}

	ft.FileID, err = ft.Staged.save(ft.tx)
	if err != nil {
		return
	}

	// File still does not exist after this - there is no commit association thus no content.
	// CurrentCommitID and hash are zero valued.
	return
}

// HasCommit returns whether the file has a commit with hash.
func (ft *FileTransaction) HasCommit(hash string) (exists bool, err error) {
	exists, err = hasCommit(ft.FileID, hash)
	if err != nil {
		return false, ft.Rollback(err)
	}

	return
}

// GetContents returns the bytes from the users temp file.
// Returns an error if the temp file is not set.
func (ft *FileTransaction) GetContents() ([]byte, error) {
	if ft.Staged == nil || len(ft.Staged.Content) == 0 {
		return nil, ft.Rollback(errors.New("temp file has no content"))
	}

	return ft.Staged.Content, nil
}

// SaveCommit saves a commit to the database.
// The files current revision will be set to the new commit.
func (ft *FileTransaction) SaveCommit(buff *bytes.Buffer, c *file.Commit) error {
	commit := &Commit{
		FileID:    ft.FileID,
		Revision:  buff.Bytes(),
		Hash:      c.Hash,
		Message:   c.Message,
		Timestamp: c.Timestamp,
	}

	newCommitID, err := insert(commit, ft.tx)
	if err != nil {
		return errors.Wrapf(err, "inserting commit for file %d", ft.Staged.ID)
	}

	ft.newCommitID = newCommitID
	return nil
}

// GetRevision returns the compressed content at hash.
func (ft *FileTransaction) GetRevision(hash string) ([]byte, error) {
	revision, err := getRevision(ft.FileID, hash)
	if err != nil {
		return nil, ft.Rollback(err)
	}

	return revision, nil
}

// SetRevision sets the file to the commit at hash.
func (ft *FileTransaction) SetRevision(hash string) error {
	row := connection.QueryRow("SELECT id FROM commits WHERE file_id = ? AND hash = ?", ft.FileID, hash)

	if err := row.Scan(&ft.newCommitID); err != nil {
		err = errors.Wrapf(err, "setting file %d to revision %q", ft.FileID, hash)
		return ft.Rollback(err)
	}

	return nil
}

// Close updates a files current commit and closes the transaction.
func (ft *FileTransaction) Close() error {
	if ft.newCommitID > 0 && ft.newCommitID != ft.CurrentCommitID {
		err := setFileToCommitID(ft.tx, ft.FileID, ft.newCommitID)
		if err != nil {
			return err
		}
	}

	if err := ft.tx.Commit(); err != nil {
		return errors.Wrapf(err, "closing transaction for file %d", ft.Staged.ID)
	}
	return nil
}

// Rollback rollsback the file transaction to its initial state.
func (ft *FileTransaction) Rollback(err error) error {
	return rollback(ft.tx, err)
}
