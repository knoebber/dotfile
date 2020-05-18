package db

import "database/sql"

type stagedFile struct {
	FileID       int64
	UserID       int64
	Alias        string
	Path         string
	DirtyContent []byte
}

// TODO test transaction logic more.
func setupStagedFile(tx *sql.Tx, userID int64, alias string) (*stagedFile, error) {
	var dirtyContent []byte

	file, err := getFileByUserID(userID, alias)
	if err != nil && !NotFound(err) {
		return nil, err
	}

	tempFile, err := GetTempFile(userID, alias)
	if err != nil && !NotFound(err) {
		return nil, err
	}

	if file == nil && tempFile == nil {
		return nil, sql.ErrNoRows
	}

	if file == nil {
		file, err = tempFile.save(tx)
		if err != nil {
			return nil, err
		}
	}

	if tempFile != nil {
		dirtyContent = tempFile.Content
	}

	return &stagedFile{
		FileID:       file.ID,
		UserID:       file.UserID,
		Alias:        file.Alias,
		Path:         file.Path,
		DirtyContent: dirtyContent,
	}, nil
}
