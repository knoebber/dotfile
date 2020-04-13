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
id                     INTEGER PRIMARY KEY,
commit_id              INTEGER NOT NULL,
compressed_content     BLOB NOT NULL,
FOREIGN KEY(commit_id) REFERENCES commits(id)
);`
}
