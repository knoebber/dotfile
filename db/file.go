package db

import (
	"database/sql"

	"github.com/pkg/errors"
)

const (
	getFileQuery       = "SELECT * FROM files WHERE userID = ? AND alias = ?"
	updateCurrentQuery = "UPDATE files SET revision = ? WHERE id = ?"
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
	ID       int64
	UserID   int64  `validate:"required"`
	Alias    string `validate:"required"` // Friendly name for a file: "bashrc"
	Path     string `validate:"required"` // Where the file lives: "~/.bashrc"
	Revision string `validate:"required"` // The current commit hash.
	Content  []byte `validate:"required"` // The content of the file at current.
}

// Unique indexes prevent a user from having duplicate alias / path.
func (*File) createStmt() string {
	return `
CREATE TABLE IF NOT EXISTS files(
id       INTEGER PRIMARY KEY,
user_id  INTEGER NOT NULL REFERENCES users,
alias    TEXT NOT NULL,
path     TEXT NOT NULL,
revision TEXT NOT NULL,
content  BLOB NOT NULL
);
CREATE INDEX IF NOT EXISTS files_user_index ON files(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS files_user_alias_index ON files(user_id, alias);
CREATE UNIQUE INDEX IF NOT EXISTS files_user_path_index ON files(user_id, path);
`
}

func (f *File) insertStmt(e executor) (sql.Result, error) {
	if err := checkSize(f.Content, "File "+f.Alias); err != nil {
		return nil, err
	}

	return e.Exec(`
INSERT INTO files(user_id, alias, path, revision, content) VALUES(?, ?, ?, ?, ?)`,
		f.UserID,
		f.Alias,
		f.Path,
		f.Revision,
		f.Content,
	)
}

func updateRevision(tx *sql.Tx, fileID int64, newHash string) error {
	if _, err := connection.Exec(updateCurrentQuery, newHash, fileID); err != nil {
		return rollback(tx, errors.Wrapf(err, "setting file %d to revision %#v", fileID, newHash))
	}
	return nil
}

func getFile(userID int64, alias string) (*File, error) {
	file := new(File)

	err := connection.QueryRow(getFileQuery, userID, alias).
		Scan(&file.ID,
			&file.UserID,
			&file.Alias,
			&file.Path,
			&file.Revision,
			&file.Content,
		)
	if err != nil {
		return nil, errors.Wrapf(err, "querying for user %d's file: %#v", userID, alias)
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
			queryErr, "querying for user %d's tracked file %#v", userID, alias)
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
