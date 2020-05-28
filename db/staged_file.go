package db

import "database/sql"

type stagedFile struct {
	FileID       int64
	UserID       int64
	Alias        string
	Path         string
	DirtyContent []byte
	New          bool
}

// TODO test transaction logic more.
func setupStagedFile(tx *sql.Tx, userID int64, alias string) (*stagedFile, error) {
	var (
		dirtyContent []byte
		new          bool
	)

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
		// No existing file. User is initialzing a new file to track.
		new = true
		file, err = tempFile.save(tx)
		if err != nil {
			return nil, err
		}
	}

	if tempFile != nil {
		dirtyContent = tempFile.Content
	}

	return &stagedFile{
		New:          new,
		FileID:       file.ID,
		UserID:       file.UserID,
		Alias:        file.Alias,
		Path:         file.Path,
		DirtyContent: dirtyContent,
	}, nil
}
