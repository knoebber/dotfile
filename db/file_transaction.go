package db

import (
	"bytes"
	"database/sql"

	"github.com/knoebber/dotfile/dotfile"
	"github.com/pkg/errors"
)

// FileTransaction implements dotfile interfaces.
// This should be created with one of the exported functions not a literal.
type FileTransaction struct {
	tx              *sql.Tx
	FileExists      bool
	FileID          int64
	CurrentCommitID int64
	Hash            string
	Path            string
	Staged          *TempFileRecord
}

// NewFileTransaction loads file information into a file transaction.
// Exported fields will be zero valued when the file doesn't exist.
func NewFileTransaction(tx *sql.Tx, username, alias string) (*FileTransaction, error) {
	ft := &FileTransaction{tx: tx}

	row := ft.tx.
		QueryRow(`
SELECT files.id, current_commit_id, hash, path
FROM files  
JOIN users ON users.id = user_id
JOIN commits ON current_commit_id = commits.id
WHERE username = ? AND alias = ?`, username, alias)

	err := row.Scan(
		&ft.FileID,
		&ft.CurrentCommitID,
		&ft.Hash,
		&ft.Path,
	)

	if NotFound(err) {
		return ft, nil
	}

	if err != nil {
		return nil, errors.Wrapf(err, "querying for user %q file %q", username, alias)
	}

	ft.FileExists = true

	return ft, nil
}

// StageFile returns a file transaction loaded with a users temp file.
// Returns an error when the temp file does not exist.
func StageFile(tx *sql.Tx, username string, alias string) (*FileTransaction, error) {
	ft, err := NewFileTransaction(tx, username, alias)
	if err != nil {
		return nil, err
	}

	ft.Staged, err = TempFile(tx, username, alias)
	if err != nil {
		return nil, err
	}

	if ft.FileExists {
		return ft, nil
	}

	ft.FileID, err = ft.Staged.save(ft.tx)
	if err != nil {
		return nil, err
	}

	// File still does not exist after this - there is no commit association thus no content.
	// CurrentCommitID and hash are zero valued.
	return ft, nil
}

// SaveFile saves a new file that does yet have any commits.
// Callers should call SaveCommit in the same transaction.
func (ft *FileTransaction) SaveFile(userID int64, alias, path string) error {
	f := &FileRecord{
		UserID: userID,
		Alias:  alias,
		Path:   path,
	}

	fileID, err := insert(ft.tx, f)
	if err != nil {
		return err
	}

	ft.FileID = fileID
	ft.FileExists = true
	return nil
}

// HasCommit returns whether the file has a commit with hash.
func (ft *FileTransaction) HasCommit(hash string) (exists bool, err error) {
	exists, err = hasCommit(ft.tx, ft.FileID, hash)
	if err != nil {
		return false, err
	}

	return
}

// DirtyContent returns the bytes from the users temp file.
// Returns an error if the temp file is not set.
func (ft *FileTransaction) DirtyContent() ([]byte, error) {
	if ft.Staged == nil || len(ft.Staged.Content) == 0 {
		return nil, errors.New("temp file has no content")
	}

	return ft.Staged.Content, nil
}

// InsertCommit saves a new commit without changing the files current revision.
func (ft *FileTransaction) InsertCommit(buff *bytes.Buffer, c *dotfile.Commit) (int64, error) {
	commit := &CommitRecord{
		FileID:    ft.FileID,
		Revision:  buff.Bytes(),
		Hash:      c.Hash,
		Message:   c.Message,
		Timestamp: c.Timestamp,
	}

	newCommitID, err := insert(ft.tx, commit)
	if err != nil {
		return 0, errors.Wrapf(err, "inserting commit for file %d", ft.Staged.ID)
	}

	return newCommitID, nil
}

// SaveCommit saves a commit to the database.
// Sets the files current revision to the new commit.
func (ft *FileTransaction) SaveCommit(buff *bytes.Buffer, c *dotfile.Commit) error {
	newCommitID, err := ft.InsertCommit(buff, c)
	if err != nil {
		return err
	}

	return setFileToCommitID(ft.tx, ft.FileID, newCommitID)
}

// Revision returns the compressed content at hash.
func (ft *FileTransaction) Revision(hash string) ([]byte, error) {
	return revision(ft.tx, ft.FileID, hash)
}

// SetRevision sets the file to the commit at hash.
func (ft *FileTransaction) SetRevision(hash string) error {
	var newCommitID int64

	row := ft.tx.QueryRow("SELECT id FROM commits WHERE file_id = ? AND hash = ?", ft.FileID, hash)

	if err := row.Scan(&newCommitID); err != nil {
		return errors.Wrapf(err, "setting file %d to revision %q", ft.FileID, hash)
	}

	return setFileToCommitID(ft.tx, ft.FileID, newCommitID)
}
