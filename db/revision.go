package db

// Revision models the revisions table.
// It stores the compressed contents a file at a commit hash.
type Revision struct {
	ID                int
	CommitID          int    `validate:"required"`
	CompressedContent []byte `validate:"required"`
}

func (*Revision) createStmt() string {
	return `
CREATE TABLE IF NOT EXISTS revisions(
id                 INTEGER PRIMARY KEY,
commit_id          INTEGER NOT NULL REFERENCES commits,
compressed_content BLOB NOT NULL
);
CREATE INDEX IF NOT EXISTS revisions_commit_index ON revisions(commit_id);
`
}
