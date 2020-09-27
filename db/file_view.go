package db

import (
	"database/sql"
	"time"

	"github.com/knoebber/dotfile/dotfile"
	"github.com/pkg/errors"
)

// FileView contains a file record and its uncompressed content.
type FileView struct {
	FileRecord
	Content []byte
	Hash    string
}

// FileSummary summarizes a file.
type FileSummary struct {
	Alias      string
	Path       string
	NumCommits int
	UpdatedAt  string
}

func (fv *FileView) scan(row *sql.Row) error {
	if err := row.Scan(
		&fv.ID,
		&fv.UserID,
		&fv.Alias,
		&fv.Path,
		&fv.CurrentCommitID,
		&fv.Content,
		&fv.Hash,
	); err != nil {
		return err
	}
	buff, err := dotfile.Uncompress(fv.Content)
	if err != nil {
		return err
	}

	fv.Content = buff.Bytes()
	return nil
}

// UncompressFile gets a file and uncompresses its current commit.
func UncompressFile(e Executor, username string, alias string) (*FileView, error) {
	fv := new(FileView)

	row := e.QueryRow(`
SELECT files.id,
       files.user_id,
       files.alias,
       files.path,
       files.current_commit_id,
       commits.revision,
       commits.hash
FROM files
JOIN users ON user_id = users.id
JOIN commits ON current_commit_id = commits.id
WHERE username = ? AND alias = ?
`, username, alias)

	if err := fv.scan(row); err != nil {
		return nil, errors.Wrapf(err, "uncompress file: querying for %q %q", username, alias)
	}

	return fv, nil
}

// FilesByUsername returns all of a users files.
func FilesByUsername(e Executor, username string) ([]FileSummary, error) {
	var (
		alias, path, timezone *string
		updatedAt             time.Time
	)

	f := FileSummary{}

	result := []FileSummary{}
	rows, err := e.Query(`
SELECT alias,
       path,
       COUNT(commits.id) AS num_commits,
       timezone,
       updated_at
FROM users
JOIN files ON user_id = users.id
JOIN commits ON file_id = files.id
WHERE username = ?
GROUP BY files.id
ORDER BY alias`, username)
	if err != nil {
		return nil, errors.Wrapf(err, "querying user %q files", username)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&alias,
			&path,
			&f.NumCommits,
			&timezone,
			&updatedAt,
		); err != nil {
			return nil, errors.Wrapf(err, "scanning files for user %#v", username)
		}

		f.UpdatedAt = formatTime(updatedAt, timezone)
		f.Alias = *alias
		f.Path = *path

		result = append(result, f)
	}
	if len(result) == 0 {
		// User doesn't exist.
		return nil, sql.ErrNoRows
	}

	return result, nil
}
