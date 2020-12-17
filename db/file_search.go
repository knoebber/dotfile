package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

const (
	fileSearchSelect = "SELECT alias, path, username, updated_at"
	fileSearchBody   = " FROM users JOIN files ON user_id = users.id"
	fileSearchWhere  = " WHERE alias LIKE ? OR path LIKE ?"
)

// FileSearchResult is the result of a file search.
type FileSearchResult struct {
	Username        string
	Alias           string
	Path            string
	UpdatedAtString string
	UpdatedAt       time.Time
}

func scanFileSearchResult(rows *sql.Rows, timezone *string) (FileSearchResult, error) {
	var result FileSearchResult

	if err := rows.Scan(
		&result.Alias,
		&result.Path,
		&result.Username,
		&result.UpdatedAt,
	); err != nil {
		return result, errors.Wrap(err, "scanning file for file search")
	}

	result.UpdatedAtString = formatTime(result.UpdatedAt, timezone)
	return result, nil
}

// FileFeed returns n of the most recently updated files.
func FileFeed(e Executor, n int, timezone *string) ([]FileSearchResult, error) {
	var result []FileSearchResult

	query := fileSearchSelect + fileSearchBody + " ORDER BY updated_at DESC" + fmt.Sprintf(" LIMIT %d", n)

	rows, err := e.Query(query)
	if err != nil {
		return nil, errors.Wrap(err, "file feed")
	}
	defer rows.Close()

	for rows.Next() {
		current, err := scanFileSearchResult(rows, timezone)
		if err != nil {
			return nil, err
		}
		result = append(result, current)
	}

	return result, nil
}

// SearchFiles looks for files by their alias or path.
func SearchFiles(e Executor, controls *PageControls, timezone *string) (*HTMLTable, error) {
	res := &HTMLTable{
		Columns:  []string{"Alias", "Path", "Username", "Updated At"},
		Controls: controls,
	}
	if controls.query == "" {
		return res, nil
	}
	q := "%" + controls.query + "%"
	countQuery := "SELECT COUNT(*) " + fileSearchBody + fileSearchWhere

	// Count the total rows and scan it into the page Controls.
	err := e.QueryRow(countQuery, q, q).Scan(&controls.totalRows)
	if err != nil {
		return nil, errors.Wrap(err, "counting rows for file search")
	}

	query := fileSearchSelect + fileSearchBody + fileSearchWhere + controls.sqlSuffix()
	rows, err := e.Query(query, q, q)
	if err != nil {
		return nil, errors.Wrap(err, "file search Query")
	}
	defer rows.Close()
	for rows.Next() {
		current, err := scanFileSearchResult(rows, timezone)
		if err != nil {
			return nil, err
		}

		res.Rows = append(res.Rows, current)
	}

	return res, nil
}
