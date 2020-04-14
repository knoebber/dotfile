package db

// File models the files table.
// It stores the contents of a file at the current revision hash.
type File struct {
	ID      int
	UserID  int    `validate:"required"`
	Alias   string `validate:"required"`
	Path    string `validate:"required"`
	Current string `validate:"required"`
	Content []byte `validate:"required"`
}

func (*File) createStmt() string {
	return `
CREATE TABLE IF NOT EXISTS files(
id       INTEGER PRIMARY KEY,
user_id  INTEGER NOT NULL REFERENCES users,
alias    TEXT NOT NULL,
path     TEXT NOT NULL,
current  TEXT NOT NULL,
content  BLOB NOT NULL
);
CREATE INDEX IF NOT EXISTS files_user_index ON files(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS files_user_alias_index ON files(user_id, alias);`
}
