package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/knoebber/dotfile/dotfile"
	"github.com/knoebber/dotfile/usererror"
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

// FileView contains a file record and its uncompressed content.
type FileView struct {
	FileRecord
	Content []byte
	Hash    string
}

// FileSummary summarizes a file.
type FileSummary struct {
	Alias      string
	Path       string
	NumCommits int
	UpdatedAt  string
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

func (f *FileRecord) check() error {
	var count int

	if err := checkFile(f.Alias, f.Path); err != nil {
		return err
	}

	exists, err := fileExists(f.UserID, f.Alias, f.Path)
	if err != nil {
		return err
	}
	if exists {
		return usererror.Duplicate("File", f.Alias)
	}

	if err := connection.QueryRow("SELECT COUNT(*) FROM files WHERE user_id = ?", f.UserID).
		Scan(&count); err != nil {
		return errors.Wrapf(err, "counting user %d file", f.UserID)
	}

	if count > maxFilesPerUser {
		return usererror.Invalid("Maximum amount of files reached")
	}

	return nil
}

func (f *FileRecord) insertStmt(e executor) (sql.Result, error) {
	return e.Exec(`
INSERT INTO files(user_id, alias, path, current_commit_id) VALUES(?, ?, ?, ?)`,
		f.UserID,
		f.Alias,
		f.Path,
		f.CurrentCommitID,
	)
}

func (f *FileView) scan(row *sql.Row) error {
	if err := row.Scan(
		&f.ID,
		&f.UserID,
		&f.Alias,
		&f.Path,
		&f.CurrentCommitID,
		&f.Content,
		&f.Hash,
	); err != nil {
		return err
	}
	buff, err := dotfile.Uncompress(f.Content)
	if err != nil {
		return err
	}

	f.Content = buff.Bytes()
	return nil
}

// Update updates the alias or path if they are different.
func (f *FileRecord) Update(newAlias, newPath string) error {
	if f.Alias == newAlias && f.Path == newPath {
		return nil
	}

	if err := checkFile(newAlias, newPath); err != nil {
		return err
	}

	_, err := connection.Exec(`
UPDATE files
SET alias = ?, path = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`, newAlias, newPath, f.ID)
	if err != nil {
		return errors.Wrapf(err, "updating file %d to %q %q", f.ID, newAlias, newPath)
	}

	return nil
}

// Delete deletes the file.
func (f *FileRecord) Delete() error {
	tx, err := connection.Begin()
	if err != nil {
		return errors.Wrap(err, "starting transaction for file delete")
	}

	_, err = tx.Exec("UPDATE files SET current_commit_id = NULL WHERE id = ?", f.ID)
	if err != nil {
		return rollback(tx, errors.Wrapf(err, "setting current commit id to null for file %d %q", f.ID, f.Alias))
	}

	_, err = tx.Exec(`
UPDATE commits SET forked_from = NULL
WHERE id IN (SELECT forked.id
             FROM commits
             JOIN commits AS forked ON forked.forked_from = commits.ID
             WHERE commits.file_id = ?)`, f.ID)
	if err != nil {
		return rollback(tx, errors.Wrapf(err, "setting forked_from to null for file %d %q", f.ID, f.Alias))
	}

	_, err = tx.Exec("DELETE FROM commits WHERE file_id = ?", f.ID)
	if err != nil {
		return rollback(tx, errors.Wrapf(err, "deleting commits for file %d %q", f.ID, f.Alias))
	}

	_, err = tx.Exec("DELETE FROM files WHERE id = ?", f.ID)
	if err != nil {
		return rollback(tx, errors.Wrapf(err, "deleting file %d %q", f.ID, f.Alias))
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "commiting file delete transaction")
	}

	return nil
}

// File retrieves a file.
func File(username string, alias string) (*FileView, error) {
	fv := new(FileView)

	row := connection.QueryRow(`
SELECT files.id,
       files.user_id,
       files.alias,
       files.path,
       files.current_commit_id,
       commits.revision,
       commits.hash
FROM files
JOIN users ON user_id = users.id
JOIN commits ON current_commit_id = commits.id
WHERE username = ? AND alias = ?
`, username, alias)

	if err := fv.scan(row); err != nil {
		return nil, errors.Wrapf(err, "querying for user %#v file %#v", username, alias)
	}

	return fv, nil
}

// FilesByUsername returns all of a users files.
func FilesByUsername(username string) ([]FileSummary, error) {
	var (
		alias, path, timezone *string
		updatedAt             time.Time
	)

	f := FileSummary{}

	result := []FileSummary{}
	rows, err := connection.Query(`
SELECT 
       alias,
       path,
       COUNT(commits.id) AS num_commits,
       timezone,
       updated_at
FROM users
JOIN files ON user_id = users.id
JOIN commits ON file_id = files.id
WHERE username = ?
GROUP BY files.id`, username)
	if err != nil {
		return nil, errors.Wrapf(err, "querying user %#v files", username)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&alias,
			&path,
			&f.NumCommits,
			&timezone,
			&updatedAt,
		); err != nil {
			return nil, errors.Wrapf(err, "scanning files for user %#v", username)
		}

		f.UpdatedAt = formatTime(updatedAt, timezone)
		f.Alias = *alias
		f.Path = *path

		result = append(result, f)
	}
	if len(result) == 0 {
		// User doesn't exist.
		return nil, sql.ErrNoRows
	}

	return result, nil
}

// ForkFile creates a copy of username/alias/hash for the user newUserID.
func ForkFile(username, alias, hash string, newUserID int64) error {
	tx, err := connection.Begin()
	if err != nil {
		return errors.Wrap(err, "starting fork file transaction")
	}

	fileForkee, err := File(username, alias)
	if err != nil {
		return rollback(tx, err)
	}

	newFile := &FileRecord{
		UserID: newUserID,
		Alias:  alias,
		Path:   fileForkee.Path,
	}

	newFileID, err := insert(newFile, tx)
	if err != nil {
		return err
	}

	commitForkee, err := Commit(username, alias, hash)
	if err != nil {
		return rollback(tx, err)
	}

	newCommit := commitForkee
	newCommit.FileID = newFileID
	newCommit.ForkedFrom = &commitForkee.ID
	newCommit.Message = fmt.Sprintf("Forked from %s", username)
	newCommit.Timestamp = time.Now().Unix()

	newCommitID, err := insert(newCommit, tx)
	if err != nil {
		return err
	}

	if err := setFileToCommitID(tx, newFileID, newCommitID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "closing fork file transaction")
	}

	return nil
}

func setFileToCommitID(tx *sql.Tx, fileID int64, commitID int64) error {
	_, err := tx.Exec(`
UPDATE files
SET current_commit_id = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`, commitID, fileID)

	if err != nil {
		return rollback(tx, errors.Wrapf(err, "updating content in file %d", fileID))
	}

	return nil
}

// SetFileToHash sets file to the commit at hash.
func SetFileToHash(username, alias, hash string) error {
	result, err := connection.Exec(`
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

func fileExists(userID int64, alias, path string) (bool, error) {
	var count int

	err := connection.
		QueryRow(`
SELECT COUNT(*) FROM files
WHERE user_id = ?
AND (alias = ? OR path = ?)`, userID, alias, path).
		Scan(&count)

	if err != nil {
		return false, errors.Wrapf(err, "checking if file %#v exists for user %d", alias, userID)
	}
	return count > 0, nil
}
