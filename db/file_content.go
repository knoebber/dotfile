package db

// FileContent implements file.Getter.
// It pulls content from temp_files and commits.
type FileContent struct {
	Username   string
	UserID     int64
	Alias      string
	Connection Executor
}

// DirtyContent returns content from the users temp file.
// Returns nil when UserID is not set - this is so that only file owners can see their temp.
func (fc *FileContent) DirtyContent() ([]byte, error) {
	if fc.UserID < 1 {
		return nil, nil
	}

	temp, err := TempFile(fc.Connection, fc.UserID, fc.Alias)
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
