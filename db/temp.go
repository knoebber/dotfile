package db

// Temp models the temps table.
// It represents a changed/new file that has not yet been commited.
// Similar to an untracked or dirty file on the filesystem.
// This allows the user to make a change to a file on the server and view a diff before saving a commit.
type Temp struct {
	ID      int
	FileID  int    `validate:"required"`
	Content []byte `validate:"required"`
}

func (*Temp) createStmt() string {
	return `
CREATE TABLE IF NOT EXISTS temps(
id                   INTEGER PRIMARY KEY,
file_id              INTEGER NOT NULL REFERENCES files,
content              BLOB NOT NULL
);
CREATE INDEX IF NOT EXISTS temps_file_index ON temps(file_id);`
}
