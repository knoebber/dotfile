package db

import "github.com/pkg/errors"

// FileContent implements file.Getter.
// It pulls content from temp_files and commits.
type FileContent struct {
	Username   string
	Alias      string
	Connection Executor
}

// Content returns the content from the users temp_file.
func (fc *FileContent) Content() ([]byte, error) {
	temp, err := TempFile(fc.Connection, fc.Username, fc.Alias)
	if err != nil {
		return nil, err
	}

	return temp.Content, nil
}

// Revision returns the compressed content at hash.
func (fc *FileContent) Revision(hash string) ([]byte, error) {
	commit, err := Commit(fc.Connection, fc.Username, fc.Alias, hash)
	if err != nil {
		return nil, err
	}

	return commit.Revision, nil
}

// HasCommit determines if a commit at hash exists.
func (fc *FileContent) HasCommit(hash string) (exists bool, err error) {
	var count int

	err = fc.Connection.QueryRow(`
SELECT COUNT(*) 
FROM commits
JOIN files ON files.id = file_id
JOIN users ON users.id = user_id
WHERE username = ? AND alias = ? AND hash = ?`, fc.Username, fc.Alias, hash).
		Scan(&count)
	if err != nil {
		return false, errors.Wrapf(err, "has commit %q %q %q", fc.Username, fc.Alias, hash)
	}
	return count > 0, nil

}
