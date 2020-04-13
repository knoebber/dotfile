package db

// File models the files table.
// It stores the contents of a file at the current revision hash.
type File struct {
	ID       int
	UserID   int    `validate:"required"`
	Alias    string `validate:"required"`
	Path     string `validate:"required"`
	Current  string `validate:"required"`
	Contents []byte `validate:"required"`
}

func (*File) createStmt() string {
	return `
CREATE TABLE IF NOT EXISTS files(
id                   INTEGER PRIMARY KEY,
user_id              INTEGER NOT NULL,
alias                TEXT NOT NULL,
path                 TEXT NOT NULL,
currents             TEXT NOT NULL,
content              BLOB NOT NULL,
FOREIGN KEY(user_id) REFERENCES users(id)
);`
}
