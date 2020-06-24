package db

import (
	"database/sql"
	"time"

	"github.com/knoebber/dotfile/file"
	"github.com/pkg/errors"
)

// TODO prevent user from creating temp file with alias that already exists.

const tempFileCountQuery = "SELECT COUNT(*) FROM temp_files WHERE user_id = ?"

// TempFile models the temp_files table.
// It represents a changed/new file that has not yet been commited.
// Similar to an untracked or dirty file on the filesystem.
// This allows the user to "stage" a file and view results before saving.
// User to TempFile is a one to one relationship.
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
CREATE UNIQUE INDEX IF NOT EXISTS temp_files_user_index ON temp_files(user_id);
`
}

func (f *TempFile) check() error {
	var count int

	if err := file.CheckPath(f.Path); err != nil {
		return err
	}

	if err := file.CheckAlias(f.Alias); err != nil {
		return err
	}

	if err := checkSize(f.Content, "File "+f.Alias); err != nil {
		return err
	}

	if err := connection.QueryRow(tempFileCountQuery, f.UserID).Scan(&count); err != nil {
		return errors.Wrapf(err, "counting user %d's temp files", f.UserID)
	}

	return nil
}

// Inserts or updates a user's previous temp file.
// Uses an UPSERT statement: https://sqlite.org/lang_UPSERT.html
func (f *TempFile) insertStmt(e executor) (sql.Result, error) {
	return e.Exec(`
INSERT INTO temp_files
(user_id, alias, path, content) 
VALUES
(?, ?, ?, ?)
ON CONFLICT(user_id) DO UPDATE
SET alias = ?, path = ?, content = ?, created_at = ?`,
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

// Create creates a new temp file.
func (f *TempFile) Create() error {
	id, err := insert(f, nil)
	if err != nil {
		return err
	}
	f.ID = id

	return nil
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
		return nil, errors.Wrapf(err, "creating file %#v for user %d", f.Alias, f.UserID)
	}

	_, err = tx.Exec("DELETE FROM temp_files WHERE user_id = ?", f.UserID)
	if err != nil {
		return nil, rollback(tx, errors.Wrapf(err, "deleting temp file %#v for user %d", f.Alias, f.UserID))
	}

	return file, nil

}

// GetTempFile finds a user's temp file.
// Users can only have one temp file at a time, so alias can be empty.
// When alias is present, ensures that temp file exists with alias.
func GetTempFile(userID int64, alias string) (*TempFile, error) {
	res := new(TempFile)

	if err := connection.
		QueryRow("SELECT * FROM temp_files WHERE user_id = ? AND (? = '' OR alias = ?)", userID, alias, alias).
		Scan(
			&res.ID,
			&res.UserID,
			&res.Alias,
			&res.Path,
			&res.Content,
			&res.CreatedAt,
		); err != nil {
		return nil, errors.Wrapf(err, "querying for user %d's temp file %#v", userID, alias)
	}

	return res, nil
}
