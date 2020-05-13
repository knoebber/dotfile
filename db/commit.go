package db

import (
	"database/sql"
	"time"

	"github.com/knoebber/dotfile/usererr"
	"github.com/pkg/errors"
)

const commitRevisionQuery = `
SELECT revision
FROM commits
JOIN files ON files.id = file_id
WHERE file_id = ? AND hash = ?`

const commitCountQuery = "SELECT COUNT(*) FROM commits WHERE file_id = ?"
const commitValidateQuery = "SELECT COUNT(*) FROM commits WHERE file_id = ? AND hash = ?"

// Commit models the commits table.
type Commit struct {
	ID        int64
	FileID    int64  `validate:"required"`
	Hash      string `validate:"required"` // Hash of the uncompressed file.
	Message   string
	Revision  []byte    `validate:"required"` // Compressed version of file at hash.
	Timestamp time.Time `validate:"required"`
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

func (c *Commit) insertStmt(e executor) (sql.Result, error) {
	var count int64

	if err := checkSize(c.Revision, "Commit "+c.Hash); err != nil {
		return nil, err
	}

	if err := connection.QueryRow(commitCountQuery, c.FileID).Scan(&count); err != nil {
		return nil, errors.Wrapf(err, "counting file %d's commits", c.FileID)
	}

	if count > maxCommitsPerFile {
		return nil, usererr.Invalid("File has maximum amount of commits")
	}

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
	if err := validateCommit(c.FileID, c.Hash); err != nil {
		return err
	}

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

func validateCommit(fileID int64, hash string) error {
	var count int

	err := connection.QueryRow(commitValidateQuery, fileID, hash).Scan(&count)
	if err != nil {
		return errors.Wrapf(err, "checking duplicate commit for file %d at %#v", fileID, hash)

	}
	if count > 0 {
		return usererr.Duplicate("File hash", hash)
	}
	return nil

}
