package db

import (
	"database/sql"

	"github.com/knoebber/usererror"
	"github.com/pkg/errors"
)

const maxCommitMessageSize = 1000

// CommitRecord models the commits table.
type CommitRecord struct {
	ID         int64
	ForkedFrom *int64 // A commit id.
	FileID     int64  `validate:"required"`
	Hash       string `validate:"required"` // Hash of the uncompressed file.
	Message    string
	Revision   []byte `validate:"required"` // Compressed version of file at hash.
	Timestamp  int64  `validate:"required"` // Unix time to stay synced with local commits.
}

// Unique index prevents a file from having a duplicate hash.
func (*CommitRecord) createStmt() string {
	return `
CREATE TABLE IF NOT EXISTS commits(
id          INTEGER PRIMARY KEY,
forked_from INTEGER REFERENCES commits,
file_id     INTEGER NOT NULL REFERENCES files,
hash        TEXT NOT NULL COLLATE NOCASE,
message     TEXT NOT NULL,
revision    BLOB NOT NULL,
timestamp   INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS commits_file_index ON commits(file_id);
CREATE INDEX IF NOT EXISTS commits_forked_from_index ON commits(forked_from);
CREATE UNIQUE INDEX IF NOT EXISTS commits_file_hash_index ON commits(file_id, hash);`
}

func (c *CommitRecord) check(e Executor) error {
	var count int

	if len(c.Message) > maxCommitMessageSize {
		return usererror.Format("The maximum commit message length is %d characters", maxCommitMessageSize)

	}

	exists, err := hasCommit(e, c.FileID, c.Hash)
	if err != nil {
		return err
	}
	if exists {
		return usererror.Format("File hash %q already exists", c.Hash)
	}

	if err := checkSize(c.Revision, "Commit "+c.Hash); err != nil {
		return err
	}

	if err := e.
		QueryRow("SELECT COUNT(*) FROM commits WHERE file_id = ?", c.FileID).
		Scan(&count); err != nil {
		return errors.Wrapf(err, "counting file %d's commits", c.FileID)
	}

	if count > maxCommitsPerFile {
		return usererror.New("File has maximum amount of commits")
	}
	return nil
}

func (c *CommitRecord) insertStmt(e Executor) (sql.Result, error) {
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

func (c *CommitRecord) create(e Executor) error {
	id, err := insert(e, c)
	if err != nil {
		return errors.Wrapf(err, "creating commit for file %d %q", c.FileID, c.Hash)
	}
	c.ID = id

	return nil
}

func revision(e Executor, fileID int64, hash string) (revision []byte, err error) {
	err = e.QueryRow(`
SELECT revision
FROM commits
JOIN files ON files.id = file_id
WHERE file_id = ? AND hash = ?
`, fileID, hash).Scan(&revision)
	if err != nil {
		err = errors.Wrapf(err, "querying for file %d at %q", fileID, hash)
	}
	return
}

func hasCommit(e Executor, fileID int64, hash string) (bool, error) {
	var count int

	err := e.QueryRow(`
SELECT COUNT(*) FROM commits
WHERE file_id = ? AND hash = ?`, fileID, hash).
		Scan(&count)
	if err != nil {
		return false, errors.Wrapf(err, "checking if commit exists for file %d at %q", fileID, hash)
	}
	return count > 0, nil

}

func usernameFromCommitID(e Executor, commitID int64) (string, error) {
	var username string

	err := e.QueryRow(`
SELECT username
FROM commits
JOIN files ON commits.file_id = files.id
JOIN users ON files.user_id = users.id
WHERE commits.id = ?`, commitID).Scan(&username)
	if err != nil {
		return "", errors.Wrapf(err, "username from commit: %d", commitID)
	}

	return username, nil
}

// Commit returns the commit record.
func Commit(e Executor, username, alias, hash string) (*CommitRecord, error) {
	result := new(CommitRecord)

	err := e.QueryRow(`
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
		return nil, errors.Wrapf(err, "querying for %q %q %q", username, alias, hash)
	}

	return result, nil
}

// ClearCommits deletes all commits for a file except the current.
func ClearCommits(tx *sql.Tx, username, alias string) error {
	file, err := File(tx, username, alias)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
UPDATE commits SET forked_from = NULL
WHERE id IN (SELECT forked.id
             FROM commits
             JOIN files ON files.id = commits.file_id
             JOIN commits AS forked ON forked.forked_from = commits.ID
             WHERE commits.file_id = ? AND commits.id != files.current_commit_id)`, file.ID)
	if err != nil {
		return errors.Wrapf(err, "setting forked_from to null for %q %q", username, alias)
	}

	_, err = tx.Exec(`
DELETE FROM commits WHERE id IN (
SELECT commits.id FROM commits 
JOIN files ON files.id = file_id
JOIN users ON users.id = user_id
WHERE username = ? AND alias = ? AND current_commit_id != commits.id)`, username, alias)
	if err != nil {
		return errors.Wrapf(err, "clearing commits for %q %q", username, alias)
	}

	return nil
}
