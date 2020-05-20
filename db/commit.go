package db

import (
	"database/sql"
	"time"

	"github.com/knoebber/dotfile/file"
	"github.com/knoebber/dotfile/usererr"
	"github.com/pkg/errors"
)

const (
	commitCountQuery    = "SELECT COUNT(*) FROM commits WHERE file_id = ?"
	commitValidateQuery = "SELECT COUNT(*) FROM commits WHERE file_id = ? AND hash = ?"
	commitRevisionQuery = `
SELECT commits.revision
FROM commits
JOIN files ON files.id = file_id
WHERE file_id = ? AND hash = ?`
	commitListQuery = `
SELECT hash,
       message, 
       timestamp
FROM commits
JOIN files ON commits.file_id = files.id
JOIN users ON files.user_id = users.id
WHERE username = ? AND alias = ?`
	commitViewQuery = `
SELECT hash,
       message,
       path,
       timestamp,
       commits.revision
FROM commits
JOIN files ON commits.file_id = files.id
JOIN users ON files.user_id = users.id
WHERE username = ? AND alias = ? AND hash = ?
`
)

// Commit models the commits table.
type Commit struct {
	ID        int64
	FileID    int64  `validate:"required"`
	Hash      string `validate:"required"` // Hash of the uncompressed file.
	Message   string
	Revision  []byte    `validate:"required"` // Compressed version of file at hash.
	Timestamp time.Time `validate:"required"`
}

// CommitSummary is a summary of a commit.
type CommitSummary struct {
	Hash      string
	Message   string
	Timestamp string
}

// CommitView is used for an individual commit view.
type CommitView struct {
	CommitSummary
	Path    string
	Content string
}

// Unique index prevents a file from having a duplicate hash.
func (*Commit) createStmt() string {
	return `
CREATE TABLE IF NOT EXISTS commits(
id        INTEGER PRIMARY KEY,
file_id   INTEGER NOT NULL REFERENCES files,
hash      TEXT NOT NULL,
message   TEXT NOT NULL,
revision  BLOB NOT NULL,
timestamp DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS commits_file_index ON commits(file_id);
CREATE UNIQUE INDEX IF NOT EXISTS commits_file_hash_index ON commits(file_id, hash);`
}

func (c *Commit) check() error {
	var count int

	exists, err := hasCommit(c.FileID, c.Hash)
	if err != nil {
		return err
	}
	if exists {
		return usererr.Duplicate("File hash", c.Hash)
	}

	if err := checkSize(c.Revision, "Commit "+c.Hash); err != nil {
		return err
	}

	if err := connection.QueryRow(commitCountQuery, c.FileID).Scan(&count); err != nil {
		return errors.Wrapf(err, "counting file %d's commits", c.FileID)
	}

	if count > maxCommitsPerFile {
		return usererr.Invalid("File has maximum amount of commits")
	}
	return nil
}

func (c *Commit) insertStmt(e executor) (sql.Result, error) {
	return e.Exec(`
INSERT INTO commits(file_id, hash, message, revision, timestamp) VALUES(?, ?, ?, ?, ?)`,
		c.FileID,
		c.Hash,
		c.Message,
		c.Revision,
		c.Timestamp,
	)
}

func (c *Commit) create(tx *sql.Tx) error {
	id, err := insert(c, tx)
	if err != nil {
		return errors.Wrapf(err, "creating commit for file %d at %#v", c.FileID, c.Hash)
	}
	c.ID = id

	return nil
}

func getRevision(fileID int64, hash string) (revision []byte, err error) {
	err = connection.QueryRow(commitRevisionQuery, fileID, hash).Scan(&revision)
	if err != nil {
		err = errors.Wrapf(err, "querying for file %d at %#v", fileID, hash)
	}
	return
}

func hasCommit(fileID int64, hash string) (bool, error) {
	var count int

	err := connection.QueryRow(commitValidateQuery, fileID, hash).Scan(&count)
	if err != nil {
		return false, errors.Wrapf(err, "checking if commit exists for file %d at %#v", fileID, hash)
	}
	return count > 0, nil

}

// GetCommitList gets a summary of all commits for a file.
func GetCommitList(username, alias string) ([]CommitSummary, error) {
	result := []CommitSummary{}
	rows, err := connection.Query(commitListQuery, username, alias)
	if err != nil {
		return nil, errors.Wrapf(err, "querying commits for user %#v file %#v", username, alias)
	}

	for rows.Next() {
		c := CommitSummary{}
		timestamp := time.Time{}

		if err := rows.Scan(
			&c.Hash,
			&c.Message,
			&timestamp,
		); err != nil {
			return nil, errors.Wrapf(err, "scanning commits for user %#v file %#v", username, alias)
		}
		c.Timestamp = formatTime(timestamp)
		result = append(result, c)
	}
	return result, nil
}

// GetCommit gets a commit and uncompresses its contents.
func GetCommit(username, alias, hash string) (*CommitView, error) {
	result := new(CommitView)
	timestamp := time.Time{}
	revision := []byte{}

	err := connection.QueryRow(commitViewQuery, username, alias, hash).
		Scan(
			&result.Hash,
			&result.Message,
			&result.Path,
			&timestamp,
			&revision,
		)
	if err != nil {
		return nil, errors.Wrapf(err, "querying for user %v commit %#v %#v", username, alias, hash)
	}
	result.Timestamp = formatTime(timestamp)

	uncompressed, err := file.Uncompress(revision)
	if err != nil {
		return nil, err
	}
	result.Content = string(uncompressed.Bytes())

	return result, nil
}
