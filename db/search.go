package db

import (
	"time"

	"github.com/pkg/errors"
)

// SearchResult is holds the result of a file search.
type SearchResult struct {
	FileSummary
	Username string
}

// Search looks for files by their alias or path.
func Search(e Executor, q string) ([]SearchResult, error) {
	var (
		timezone  *string
		updatedAt time.Time
	)

	q = "%" + q + "%"
	current := SearchResult{}
	result := []SearchResult{}
	rows, err := e.Query(`
SELECT 
       username,
       alias,
       path,
       COUNT(commits.id) AS num_commits,
       timezone,
       updated_at
FROM users
JOIN files ON user_id = users.id
JOIN commits ON file_id = files.id
WHERE alias LIKE ? OR path LIKE ?
GROUP BY files.id
ORDER BY username, alias`, q, q)
	if err != nil {
		return nil, errors.Wrapf(err, "querying for files LIKE %#v", q)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&current.Username,
			&current.Alias,
			&current.Path,
			&current.NumCommits,
			&timezone,
			&updatedAt,
		); err != nil {
			return nil, errors.Wrap(err, "scanning files for file search")
		}

		current.UpdatedAt = formatTime(updatedAt, timezone)

		result = append(result, current)
	}

	return result, nil
}
