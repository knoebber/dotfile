package db

import (
	"database/sql"
	"time"

	"github.com/knoebber/dotfile/dotfile"
	"github.com/knoebber/dotfile/usererror"
	"github.com/pkg/errors"
)

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

// CommitSummary summarizes a commit.
type CommitSummary struct {
	Hash               string
	Message            string
	Current            bool
	Timestamp          int64
	DateString         string
	ForkedFromUsername *string // The owner of the file that this commit was forked from.
}

// CommitView is used for an individual commit view.
type CommitView struct {
	CommitSummary
	Path    string
	Content []byte
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

func (c *CommitRecord) check() error {
	var count int

	exists, err := hasCommit(c.FileID, c.Hash)
	if err != nil {
		return err
	}
	if exists {
		return usererror.Duplicate("File hash", c.Hash)
	}

	if err := checkSize(c.Revision, "Commit "+c.Hash); err != nil {
		return err
	}

	if err := connection.
		QueryRow("SELECT COUNT(*) FROM commits WHERE file_id = ?", c.FileID).
		Scan(&count); err != nil {
		return errors.Wrapf(err, "counting file %d's commits", c.FileID)
	}

	if count > maxCommitsPerFile {
		return usererror.Invalid("File has maximum amount of commits")
	}
	return nil
}

func (c *CommitRecord) insertStmt(e executor) (sql.Result, error) {
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

func (c *CommitRecord) create(tx *sql.Tx) error {
	id, err := insert(c, tx)
	if err != nil {
		return errors.Wrapf(err, "creating commit for file %d at %q", c.FileID, c.Hash)
	}
	c.ID = id

	return nil
}

func revision(fileID int64, hash string) (revision []byte, err error) {
	err = connection.QueryRow(`
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

func hasCommit(fileID int64, hash string) (bool, error) {
	var count int

	err := connection.
		QueryRow("SELECT COUNT(*) FROM commits WHERE file_id = ? AND hash = ?", fileID, hash).
		Scan(&count)
	if err != nil {
		return false, errors.Wrapf(err, "checking if commit exists for file %d at %q", fileID, hash)
	}
	return count > 0, nil

}

func usernameFromCommitID(commitID int64) (string, error) {
	var username string

	err := connection.QueryRow(`
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

// CommitList gets a summary of all commits for a file.
func CommitList(username, alias string) ([]CommitSummary, error) {
	var (
		timezone   *string
		forkedFrom *int64
	)

	result := []CommitSummary{}
	rows, err := connection.Query(`
SELECT hash,
       forked_from,
       message, 
       current_commit_id = commits.id AS current,
       timezone,
       timestamp
FROM commits
JOIN files ON commits.file_id = files.id
JOIN users ON files.user_id = users.id
WHERE username = ? AND alias = ?
ORDER BY timestamp DESC
`, username, alias)
	if err != nil {
		return nil, errors.Wrapf(err, "querying commits for user %q file %q", username, alias)
	}
	defer rows.Close()

	for rows.Next() {
		c := CommitSummary{}

		if err := rows.Scan(
			&c.Hash,
			&forkedFrom,
			&c.Message,
			&c.Current,
			&timezone,
			&c.Timestamp,
		); err != nil {
			return nil, errors.Wrapf(err, "scanning commits for user %q file %q", username, alias)
		}
		c.DateString = formatTime(time.Unix(c.Timestamp, 0), timezone)
		if forkedFrom != nil {
			username, err := usernameFromCommitID(*forkedFrom)
			if err != nil {
				return nil, err
			}
			c.ForkedFromUsername = &username
		}
		result = append(result, c)
	}

	if len(result) == 0 {
		return nil, sql.ErrNoRows
	}

	return result, nil
}

// Commit returns the commit record.
func Commit(username, alias, hash string) (*CommitRecord, error) {
	result := new(CommitRecord)

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
		return nil, errors.Wrapf(err, "querying for %q %q %q", username, alias, hash)
	}

	return result, nil
}

// UncompressedCommit gets a commit and uncompresses its contents.
func UncompressedCommit(username, alias, hash string) (*CommitView, error) {
	var (
		timezone   *string
		forkedFrom *int64
	)

	result := new(CommitView)
	revision := []byte{}

	err := connection.QueryRow(`
SELECT hash,
       forked_from,
       message,
       path,
       current_commit_id = commits.id AS current,
       revision,
       timezone,
       timestamp
FROM commits
JOIN files ON commits.file_id = files.id
JOIN users ON files.user_id = users.id
WHERE username = ? AND alias = ? AND hash = ?
`, username, alias, hash).
		Scan(
			&result.Hash,
			&forkedFrom,
			&result.Message,
			&result.Path,
			&result.Current,
			&revision,
			&timezone,
			&result.Timestamp,
		)
	if err != nil {
		return nil, errors.Wrapf(err, "querying for uncompressed %q %q %q", username, alias, hash)
	}

	result.DateString = formatTime(time.Unix(result.Timestamp, 0), timezone)
	if forkedFrom != nil {
		username, err := usernameFromCommitID(*forkedFrom)
		if err != nil {
			return nil, err
		}
		result.ForkedFromUsername = &username
	}

	uncompressed, err := dotfile.Uncompress(revision)
	if err != nil {
		return nil, err
	}
	result.Content = uncompressed.Bytes()

	return result, nil
}
