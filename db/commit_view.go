package db

import (
	"database/sql"
	"time"

	"github.com/knoebber/dotfile/dotfile"
	"github.com/pkg/errors"
)

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

// CommitList gets a summary of all commits for a file.
func CommitList(e Executor, username, alias string) ([]CommitSummary, error) {
	var (
		timezone   *string
		forkedFrom *int64
	)

	result := []CommitSummary{}
	rows, err := e.Query(`
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
			username, err := usernameFromCommitID(e, *forkedFrom)
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

// UncompressCommit gets a commit and uncompresses its contents.
func UncompressCommit(e Executor, username, alias, hash string) (*CommitView, error) {
	var (
		timezone   *string
		forkedFrom *int64
	)

	result := new(CommitView)
	revision := []byte{}

	err := e.QueryRow(`
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
		username, err := usernameFromCommitID(e, *forkedFrom)
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
