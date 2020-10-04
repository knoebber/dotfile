package db

import (
	"time"

	"github.com/pkg/errors"
)

const (
	fileSearchSelect = `
SELECT
alias,
path,
username,
updated_at,
timezone
`
	fileSearchBody = `
FROM users
JOIN files ON user_id = users.id
WHERE alias LIKE ? OR path LIKE ?
`
)

// FileSearchResult is the result of a file search.
type FileSearchResult struct {
	Username  string
	Alias     string
	Path      string
	UpdatedAt string
}

// SearchFiles looks for files by their alias or path.
func SearchFiles(e Executor, controls *PageControls) (*HTMLTable, error) {
	var (
		timezone  *string
		updatedAt time.Time
	)

	res := &HTMLTable{
		Columns:  []string{"Alias", "Path", "Username", "Updated At"},
		Controls: controls,
	}
	if controls.query == "" {
		return res, nil
	}
	q := "%" + controls.query + "%"

	// Count the total rows and scan it into the page Controls.
	err := e.QueryRow("SELECT COUNT(*) "+fileSearchBody, q, q).Scan(&controls.totalRows)
	if err != nil {
		return nil, errors.Wrap(err, "counting rows for file search")
	}

	current := FileSearchResult{}
	rows, err := e.Query(fileSearchSelect+fileSearchBody+controls.sqlSuffix(), q, q)
	if err != nil {
		return nil, errors.Wrap(err, "file search Query")
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(
			&current.Alias,
			&current.Path,
			&current.Username,
			&updatedAt,
			&timezone,
		); err != nil {
			return nil, errors.Wrap(err, "scanning files for file search")
		}

		current.UpdatedAt = formatTime(updatedAt, timezone)
		res.Rows = append(res.Rows, current)
	}

	return res, nil
}
