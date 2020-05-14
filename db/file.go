package db

import (
	"bytes"
	"database/sql"
	"time"

	"github.com/knoebber/dotfile/usererr"
	"github.com/pkg/errors"
)

const (
	fileCountQuery     = "SELECT COUNT(*) FROM files WHERE user_id = ?"
	fileValidateQuery  = "SELECT COUNT(*) FROM files WHERE user_id = ? AND alias = ?"
	getFileQuery       = "SELECT * FROM files WHERE user_id = ? AND alias = ?"
	updateCurrentQuery = "UPDATE files SET revision = ? WHERE id = ?"
	updateContentQuery = "UPDATE files SET content = ?, revision = ? WHERE id = ?"
	fileCommitsQuery   = `
SELECT file_id,
       alias, 
       path,
       current, 
       hash,
       message, 
       timestamp 
FROM files
JOIN commits ON commits.file_id = files.id
WHERE userID = ? AND alias = ?"`
)

// File models the files table.
// It stores the contents of a file at the current revision hash.
//
// Both aliases and paths must be unique for each user.
type File struct {
	ID        int64
	UserID    int64  `validate:"required"`
	Alias     string `validate:"required"` // Friendly name for a file: "bashrc"
	Path      string `validate:"required"` // Where the file lives: "~/.bashrc"
	Revision  string
	Content   []byte `validate:"required"`
	CreatedAt time.Time
}

// Unique indexes prevent a user from having duplicate alias / path.
func (*File) createStmt() string {
	return `
CREATE TABLE IF NOT EXISTS files(
id         INTEGER PRIMARY KEY,
user_id    INTEGER NOT NULL REFERENCES users,
alias      TEXT NOT NULL,
path       TEXT NOT NULL,
revision   TEXT NOT NULL,
content    BLOB NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS files_user_index ON files(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS files_user_alias_index ON files(user_id, alias);
CREATE UNIQUE INDEX IF NOT EXISTS files_user_path_index ON files(user_id, path);
`
}

func (f *File) check() error {
	var count int64

	exists, err := fileExists(f.UserID, f.Alias)
	if err != nil {
		return err
	}
	if exists {
		return usererr.Duplicate("File alias", f.Alias)
	}

	if err := checkSize(f.Content, "File "+f.Alias); err != nil {
		return err
	}

	if err := connection.QueryRow(fileCountQuery, f.UserID).Scan(&count); err != nil {
		return errors.Wrapf(err, "counting user %d's file", f.UserID)
	}

	if count > maxFilesPerUser {
		return usererr.Invalid("User has maximum amount of files")
	}

	return nil
}

func (f *File) insertStmt(e executor) (sql.Result, error) {
	return e.Exec(`
INSERT INTO files(user_id, alias, path, revision, content) VALUES(?, ?, ?, ?, ?)`,
		f.UserID,
		f.Alias,
		f.Path,
		f.Revision,
		f.Content,
	)
}

func updateContent(tx *sql.Tx, fileID int64, buff *bytes.Buffer, hash string) error {
	newContent := buff.Bytes()
	if err := checkSize(newContent, "File revision "+hash); err != nil {
		return err
	}

	if _, err := tx.Exec(updateCurrentQuery, newContent, hash, fileID); err != nil {
		return rollback(tx, errors.Wrapf(err, "updating file %d content to %#v", fileID, hash))
	}
	return nil
}

func updateRevision(tx *sql.Tx, fileID int64, newHash string) error {
	if _, err := tx.Exec(updateCurrentQuery, newHash, fileID); err != nil {
		return rollback(tx, errors.Wrapf(err, "setting file %d to revision %#v", fileID, newHash))
	}
	return nil
}

func getFile(userID int64, alias string) (*File, error) {
	file := new(File)

	err := connection.QueryRow(getFileQuery, userID, alias).
		Scan(
			&file.ID,
			&file.UserID,
			&file.Alias,
			&file.Path,
			&file.Revision,
			&file.Content,
		)
	if err != nil {
		return nil, errors.Wrapf(err, "querying for user %d's file %#v", userID, alias)
	}

	return file, nil
}

func getFileAndCommits(userID int64, alias string) (*File, []Commit, error) {
	file := new(File)
	commits := []Commit{}

	rows, queryErr := connection.Query(fileCommitsQuery, userID, alias)

	if errors.Is(queryErr, sql.ErrNoRows) {
		return nil, nil, nil
	} else if queryErr != nil {
		return nil, nil, errors.Wrapf(
			queryErr, "querying for user %d's file %#v", userID, alias)
	}

	for rows.Next() {
		commit := Commit{}
		if err := rows.Scan(
			&file.ID,
			&file.Alias,
			&file.Path,
			&file.Revision,
			&commit.Hash,
			&commit.Message,
			&commit.Timestamp,
		); err != nil {
			err = errors.Wrapf(err, "scanning user %d's tracked file %#v", userID, alias)
			return nil, nil, err
		}
		commits = append(commits, commit)
	}

	return file, commits, nil
}

func fileExists(userID int64, alias string) (bool, error) {
	var count int

	err := connection.QueryRow(fileValidateQuery, userID, alias).Scan(&count)
	if err != nil {
		return false, errors.Wrapf(err, "checking if file %#v exists for user %d", alias, userID)
	}
	return count > 0, nil
}
