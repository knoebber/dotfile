package db

import (
	"time"
)

// Commit models the commits table.
type Commit struct {
	ID        int
	FileID    int    `validate:"required"`
	Hash      string `validate:"required"`
	Message   *string
	Timestamp time.Time
}

func (*Commit) createStmt() string {
	return `
CREATE TABLE IF NOT EXISTS commits(
id                   INTEGER PRIMARY KEY,
file_id              INTEGER NOT NULL,
hash                 TEXT NOT NULL UNIQUE,
message              TEXT,
timestamp            DATETIME NOT NULL,
FOREIGN KEY(file_id) REFERENCES files(id)
);`
}
