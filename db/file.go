package db

import (
	"database/sql"
	"github.com/knoebber/dotfile/file"
	"github.com/pkg/errors"
)

const trackedFileQuery = `
SELECT alias, 
       path,
       current, 
       hash,
       message, 
       timestamp 
FROM files
JOIN commits ON commits.file_id = files.id
`

// File models the files table.
// It stores the contents of a file at the current revision hash.
type File struct {
	ID      int
	UserID  int    `validate:"required"`
	Alias   string `validate:"required"`
	Path    string `validate:"required"` // Where the file lives - ~/.bashrc
	Current string `validate:"required"` // The current commit hash.
	Content []byte `validate:"required"` // The content of the file at current.
}

func (*File) createStmt() string {
	return `
CREATE TABLE IF NOT EXISTS files(
id       INTEGER PRIMARY KEY,
user_id  INTEGER NOT NULL REFERENCES users,
alias    TEXT NOT NULL,
path     TEXT NOT NULL,
current  TEXT NOT NULL,
content  BLOB NOT NULL
);
CREATE INDEX IF NOT EXISTS files_user_index ON files(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS files_user_alias_index ON files(user_id, alias);`
}

func (f *File) insertStmt() (sql.Result, error) {
	return connection.Exec("INSERT INTO files(user_id, alias, path, current, content) VALUES(?, ?, ?, ?)",
		f.UserID,
		f.Alias,
		f.Path,
		f.Current,
		f.Content,
	)
}

func getTrackedFile(userID int64, alias string) (*file.Tracked, error) {
	tf := new(file.Tracked)
	tf.Commits = []file.Commit{}

	rows, err := connection.Query(trackedFileQuery+"WHERE userID = ? AND alias = ?", userID, alias)
	if err != nil {
		return nil, errors.Wrapf(err, "querying for user %d's tracked file %#v", userID, alias)
	}

	for rows.Next() {
		commit := file.Commit{}
		if err := rows.Scan(
			&tf.Alias,
			&tf.RelativePath,
			&tf.Revision,
			&commit.Hash,
			&commit.Message,
			&commit.Timestamp,
		); err != nil {
			return nil, errors.Wrapf(err, "scanning user %d's tracked file %#v", userID, alias)
		}
		tf.Commits = append(tf.Commits, commit)
	}

	return tf, nil
}
