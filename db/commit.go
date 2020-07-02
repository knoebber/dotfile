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
SELECT revision
FROM commits
JOIN files ON files.id = file_id
WHERE file_id = ? AND hash = ?`
)

// Commit models the commits table.
type Commit struct {
	ID         int64
	ForkedFrom *int64 // A commit id.
	FileID     int64  `validate:"required"`
	Hash       string `validate:"required"` // Hash of the uncompressed file.
	Message    string
	Revision   []byte `validate:"required"` // Compressed version of file at hash.
	Timestamp  int64  `validate:"required"` // Unix time to stay synced with local commits.
}

// CommitSummary summarizes a commit.
type CommitSummary struct {
	ForkedFrom *int64
	Hash       string
	Message    string
	Current    bool
	Timestamp  int64
	DateString string
}

// CommitView is used for an individual commit view.
type CommitView struct {
	CommitSummary
	Path    string
	Content []byte
}

// Unique index prevents a file from having a duplicate hash.
func (*Commit) createStmt() string {
	return `
CREATE TABLE IF NOT EXISTS commits(
id          INTEGER PRIMARY KEY,
forked_from INTEGER REFERENCES commits,
file_id     INTEGER NOT NULL REFERENCES files,
hash        TEXT NOT NULL,
message     TEXT NOT NULL,
revision    BLOB NOT NULL,
timestamp   INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS commits_file_index ON commits(file_id);
CREATE INDEX IF NOT EXISTS commits_forked_from_index ON commits(forked_from);
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
INSERT INTO commits(forked_from, file_id, hash, message, revision, timestamp) VALUES(?, ?, ?, ?, ?, ?)`,
		c.ForkedFrom,
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
	rows, err := connection.Query(`
SELECT hash,
       forked_from,
       message, 
       hash = current_revision AS current,
       timestamp
FROM commits
JOIN files ON commits.file_id = files.id
JOIN users ON files.user_id = users.id
WHERE username = ? AND alias = ?
ORDER BY timestamp DESC
`, username, alias)
	if err != nil {
		return nil, errors.Wrapf(err, "querying commits for user %#v file %#v", username, alias)
	}
	defer rows.Close()

	for rows.Next() {
		c := CommitSummary{}

		if err := rows.Scan(
			&c.Hash,
			&c.ForkedFrom,
			&c.Message,
			&c.Current,
			&c.Timestamp,
		); err != nil {
			return nil, errors.Wrapf(err, "scanning commits for user %#v file %#v", username, alias)
		}
		c.DateString = formatTime(time.Unix(c.Timestamp, 0))
		result = append(result, c)
	}

	if len(result) == 0 {
		return nil, sql.ErrNoRows
	}

	return result, nil
}

// Returns the commit record.
func getCommit(username, alias, hash string) (*Commit, error) {
	result := new(Commit)

	err := connection.QueryRow(`
SELECT commits.*
FROM commits
JOIN files ON commits.file_id = files.id
JOIN users ON files.user_id = users.id
WHERE username = ? AND alias = ? AND hash = ?`, username, alias, hash).
		Scan(
			&result.ID,
			&result.ForkedFrom,
			&result.FileID,
			&result.Hash,
			&result.Message,
			&result.Revision,
			&result.Timestamp,
		)
	if err != nil {
		return nil, errors.Wrapf(err, "get commit: %#v %#v %#v", username, alias, hash)
	}

	return result, nil
}

// GetCommit gets a commit and uncompresses its contents.
func GetCommit(username, alias, hash string) (*CommitView, error) {
	result := new(CommitView)
	revision := []byte{}

	err := connection.QueryRow(`
SELECT hash,
       forked_from,
       message,
       path,
       hash = current_revision AS current,
       revision,
       timestamp
FROM commits
JOIN files ON commits.file_id = files.id
JOIN users ON files.user_id = users.id
WHERE username = ? AND alias = ? AND hash = ?
`, username, alias, hash).
		Scan(
			&result.Hash,
			&result.ForkedFrom,
			&result.Message,
			&result.Path,
			&result.Current,
			&revision,
			&result.Timestamp,
		)
	if err != nil {
		return nil, errors.Wrapf(err, "querying for %#v %#v %#v", username, alias, hash)
	}

	result.DateString = formatTime(time.Unix(result.Timestamp, 0))

	uncompressed, err := file.Uncompress(revision)
	if err != nil {
		return nil, err
	}
	result.Content = uncompressed.Bytes()

	return result, nil
}
