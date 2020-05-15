package db

import (
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

const (
	tempFileCountQuery   = "SELECT COUNT(*) FROM temp_files WHERE user_id = ?"
	tempFileQuery        = "SELECT * FROM temp_files"
	tempFileDeleteQuery  = "DELETE FROM temp_files WHERE id = ?"
	tempFileCleanupQuery = `
DELETE FROM temp_files 
WHERE id IN 
  (SELECT id 
   FROM temp_files 
   WHERE user_id = ? 
   ORDER BY created_at
   LIMIT 5)
`
)

// TempFile models the temp_files table.
// It represents a changed/new file that has not yet been commited.
// Similar to an untracked or dirty file on the filesystem.
// This allows the user to "stage" a file and view results before saving.
// Columns mirror the file table for the most part.
// This is split into its own table to keep the unique indexes simple.
type TempFile struct {
	ID        int64
	UserID    int64  `validate:"required"`
	Alias     string `validate:"required"`
	Path      string `validate:"required"`
	Content   []byte `validate:"required"`
	CreatedAt time.Time
}

func (*TempFile) createStmt() string {
	return `
CREATE TABLE IF NOT EXISTS temp_files(
id         INTEGER PRIMARY KEY,
user_id    INTEGER NOT NULL REFERENCES users,
alias      TEXT NOT NULL,
path       TEXT NOT NULL,
content    BLOB NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS temp_files_user_index ON temp_files(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS temp_files_user_alias_index ON temp_files(user_id, alias);
CREATE UNIQUE INDEX IF NOT EXISTS temp_files_user_path_index ON temp_files(user_id, path);
`
}

func (f *TempFile) check() error {
	var count int64

	if err := checkSize(f.Content, "File "+f.Alias); err != nil {
		return err
	}

	if err := connection.QueryRow(tempFileQuery, f.UserID).Scan(&count); err != nil {
		return errors.Wrapf(err, "counting user %d's temp files", f.UserID)
	}

	if count < maxTempsPerUser {
		return nil
	}

	// Delete the oldest five temp files.
	if _, err := connection.Exec(tempFileCleanupQuery, f.UserID); err != nil {
		return errors.Wrapf(err, "cleaning up user %d's temp files", f.UserID)
	}

	return nil
}

// https://sqlite.org/lang_UPSERT.html
func (f *TempFile) insertStmt(e executor) (sql.Result, error) {
	if err := checkSize(f.Content, "Temp file "+f.Alias); err != nil {
		return nil, err
	}

	return e.Exec(`
UPSERT INTO temp_files
(user_id, alias, path, content) 
VALUES
(?, ?, ?, ?)
ON CONFLICT DO UPDATE
SET alias = ?, path = ?, content = ?, created_at = ?
`,
		f.UserID,
		f.Alias,
		f.Path,
		f.Content,
		f.Alias,
		f.Path,
		f.Content,
		time.Now(),
	)
}

func (f *TempFile) save(tx *sql.Tx) (*File, error) {
	var err error

	file := &File{
		UserID:  f.UserID,
		Alias:   f.Alias,
		Path:    f.Path,
		Content: f.Content,
	}

	file.ID, err = insert(file, tx)
	if err != nil {
		return nil, rollback(tx, errors.Wrapf(err, "creating file %#v for user %d", f.Alias, f.UserID))
	}

	if _, err = tx.Exec(tempFileDeleteQuery, f.ID); err != nil {
		return nil, errors.Wrapf(err, "cleaning up temp file %d", f.ID)
	}

	return file, nil

}

func getTempFileByAlias(userID int64, path string) (*TempFile, error) {
	res := new(TempFile)

	if err := connection.
		QueryRow(tempFileQuery+" WHERE user_id = ? AND alias = ?", userID, path).
		Scan(
			&res.ID,
			&res.UserID,
			&res.Alias,
			&res.Path,
			&res.Content,
			&res.CreatedAt,
		); err != nil {
		return nil, errors.Wrapf(err, "querying for user %d's temp file %#v", userID, path)
	}

	return res, nil
}
