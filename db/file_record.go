package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/knoebber/dotfile/dotfile"
	"github.com/knoebber/usererror"
	"github.com/pkg/errors"
)

// FileRecord models the files table.
// It stores the contents of a file at the current revision hash.
//
// Both aliases and paths must be unique for each user.
type FileRecord struct {
	ID              int64
	UserID          int64  `validate:"required"`
	Alias           string `validate:"required"` // Friendly name for a file: bashrc
	Path            string `validate:"required"` // Where the file lives: ~/.bashrc
	CurrentCommitID *int64 // The commit that the file is at.
}

// Unique indexes prevent a user from having duplicate alias / path.
func (*FileRecord) createStmt() string {
	return `
CREATE TABLE IF NOT EXISTS files(
id                 INTEGER PRIMARY KEY,
user_id            INTEGER NOT NULL REFERENCES users,
alias              TEXT NOT NULL COLLATE NOCASE,
path               TEXT NOT NULL COLLATE NOCASE,
current_commit_id  INTEGER REFERENCES commits,
created_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
updated_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS files_user_index ON files(user_id);
CREATE INDEX IF NOT EXISTS files_commit_index ON files(current_commit_id);
CREATE UNIQUE INDEX IF NOT EXISTS files_user_alias_index ON files(user_id, alias);
CREATE UNIQUE INDEX IF NOT EXISTS files_user_path_index ON files(user_id, path);
`
}

func (f *FileRecord) check(e Executor) error {
	var count int

	if err := checkFile(f.Alias, f.Path); err != nil {
		return err
	}
	if err := ValidateFileNotExists(e, f.UserID, f.Alias, f.Path); err != nil {
		return err
	}

	if err := e.QueryRow("SELECT COUNT(*) FROM files WHERE user_id = ?", f.UserID).
		Scan(&count); err != nil {
		return errors.Wrapf(err, "counting user %d file", f.UserID)
	}

	if count > maxFilesPerUser {
		return usererror.New("Maximum amount of files reached")
	}

	return nil
}

func (f *FileRecord) insertStmt(e Executor) (sql.Result, error) {
	return e.Exec(`
INSERT INTO files(user_id, alias, path, current_commit_id) VALUES(?, ?, ?, ?)`,
		f.UserID,
		strings.ToLower(f.Alias),
		f.Path,
		f.CurrentCommitID,
	)
}

// Update updates the alias or path if they are different.
func (f *FileRecord) Update(e Executor, newAlias, newPath string) error {
	newAlias = strings.ToLower(newAlias)

	if f.Alias == newAlias && f.Path == newPath {
		return nil
	}

	if err := checkFile(newAlias, newPath); err != nil {
		return err
	}

	_, err := e.Exec(`
UPDATE files
SET alias = ?, path = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`, newAlias, newPath, f.ID)
	if err != nil {
		return errors.Wrapf(err, "updating file %d to %q %q", f.ID, newAlias, newPath)
	}

	return nil
}

// DeleteFile deletes a users file.
func DeleteFile(tx *sql.Tx, username, alias string) error {
	record, err := File(tx, username, alias)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE files SET current_commit_id = NULL WHERE id = ?", record.ID)
	if err != nil {
		return errors.Wrapf(err, "setting current commit id to null for %q %q", username, alias)
	}

	_, err = tx.Exec(`
UPDATE commits SET forked_from = NULL
WHERE id IN (SELECT forked.id
             FROM commits
             JOIN commits AS forked ON forked.forked_from = commits.ID
             WHERE commits.file_id = ?)`, record.ID)
	if err != nil {
		return errors.Wrapf(err, "setting forked_from to null for %q %q", username, alias)
	}

	_, err = tx.Exec("DELETE FROM commits WHERE file_id = ?", record.ID)
	if err != nil {
		return errors.Wrapf(err, "deleting commits for %q %q", username, alias)
	}

	_, err = tx.Exec("DELETE FROM files WHERE id = ?", record.ID)
	if err != nil {
		return errors.Wrapf(err, "deleting file %q %q", username, alias)
	}

	return nil
}

// File retrieves a file record.
func File(e Executor, username string, alias string) (*FileRecord, error) {
	record := new(FileRecord)

	err := e.QueryRow(`
SELECT files.id, 
       user_id, 
       alias, 
       path, 
       current_commit_id
FROM files 
JOIN users ON user_id = users.id 
WHERE username = ? AND alias = ?`, username, alias).
		Scan(
			&record.ID,
			&record.UserID,
			&record.Alias,
			&record.Path,
			&record.CurrentCommitID,
		)
	if err != nil {
		return nil, errors.Wrapf(err, "querying file for %q %q", username, alias)
	}

	return record, nil
}

// FileData returns the files dotfile data structure.
func FileData(e Executor, username, alias string) (*dotfile.TrackingData, error) {
	var (
		path, hash, message string
		current             bool
		timestamp           int64
	)

	result := new(dotfile.TrackingData)

	rows, err := e.Query(`
SELECT path,
       hash,
       message,
       timestamp,
       current_commit_id = commits.id AS current
FROM users
JOIN files ON files.user_id = users.id
JOIN commits ON commits.file_id = files.id
WHERE username = ? AND alias = ?`, username, alias)
	if err != nil {
		return nil, errors.Wrapf(err, "file data: querying for %q %q", username, alias)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&path,
			&hash,
			&message,
			&timestamp,
			&current,
		); err != nil {
			return nil, errors.Wrapf(err, "file data %q %q", username, alias)
		}
		result.Path = path
		if current {
			result.Revision = hash
		}

		result.Commits = append(result.Commits, dotfile.Commit{
			Hash:      hash,
			Message:   message,
			Timestamp: timestamp,
		})
	}
	if len(result.Commits) == 0 {
		return nil, sql.ErrNoRows
	}

	return result, nil

}

// SetFileToHash sets file to the commit at hash.
func SetFileToHash(e Executor, username, alias, hash string) error {
	result, err := e.Exec(`
WITH new_commit(id, file_id) AS (
SELECT commits.id,
       file_id
FROM commits
JOIN files ON files.id = commits.file_id
JOIN users ON files.user_id = users.id
WHERE username = ? AND alias = ? AND hash = ?
)
UPDATE files
SET current_commit_id = (SELECT new_commit.id FROM new_commit), updated_at = CURRENT_TIMESTAMP
WHERE id = (SELECT file_id FROM new_commit)
`, username, alias, hash)
	if err != nil {
		return errors.Wrapf(err, "setting %q %q to hash %q", username, alias, hash)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "getting rows affected in set file to hash")
	}
	if affected == 0 {
		return fmt.Errorf("commit %q %q %q not found", username, alias, hash)
	}

	return nil

}

// ForkFile creates a copy of username/alias/hash for the user newUserID.
func ForkFile(username, alias, hash string, newUserID int64) error {
	tx, err := Connection.Begin()
	if err != nil {
		return errors.Wrap(err, "starting fork file transaction")
	}
	if err := forkFile(tx, username, alias, hash, newUserID); err != nil {
		return Rollback(tx, err)
	}
	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "committing fork file transaction")
	}
	return nil
}

func forkFile(tx *sql.Tx, username, alias, hash string, newUserID int64) error {
	original, err := File(tx, username, alias)
	if err != nil {
		return err
	}

	newFile := &FileRecord{
		UserID: newUserID,
		Alias:  alias,
		Path:   original.Path,
	}

	newFileID, err := insert(tx, newFile)
	if err != nil {
		return err
	}

	currentCommit, err := Commit(tx, username, alias, hash)
	if err != nil {
		return err
	}

	newCommit := currentCommit
	newCommit.FileID = newFileID
	newCommit.ForkedFrom = &currentCommit.ID
	newCommit.Message = fmt.Sprintf("Forked from %s", username)
	newCommit.Timestamp = time.Now().Unix()

	newCommitID, err := insert(tx, newCommit)
	if err != nil {
		return err
	}

	if err := setFileToCommitID(tx, newFileID, newCommitID); err != nil {
		return err
	}

	return nil
}

// ValidateFileNotExists validates that no other file for user exists with alias or path.
func ValidateFileNotExists(e Executor, userID int64, alias, path string) error {
	var count int

	err := e.QueryRow(`
SELECT COUNT(*) FROM files
WHERE user_id = ?
AND (alias = ? OR path = ?)`, userID, alias, path).
		Scan(&count)

	if err != nil {
		return errors.Wrapf(err, "checking if file %q exists for user %d", alias, userID)
	}
	if count > 0 {
		return usererror.Format("File %q already exists", alias)
	}

	return nil
}

func setFileToCommitID(e Executor, fileID int64, commitID int64) error {
	_, err := e.Exec(`
UPDATE files
SET current_commit_id = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`, commitID, fileID)

	if err != nil {
		return errors.Wrapf(err, "updating content in file %d", fileID)
	}

	return nil
}
