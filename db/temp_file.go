package db

import (
	"time"
)

// TempFile models the temp_files table.
// It represents a changed/new file that has not yet been commited.
// Similar to an untracked or dirty file on the filesystem.
// This allows the user to make a change to a file on the server and view a diff before saving a commit.
// Columns mirror the file table for the most part. It's split into its own table to keep the unique indexes simple.
type TempFile struct {
	ID        int
	UserID    int    `validate:"required"`
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
CREATE UNIQUE INDEX IF NOT EXISTS temp_files_user_alias_index ON temp_files(user_id, alias);`
}
