package db

import (
	"bytes"
	"testing"

	"github.com/knoebber/dotfile/file"
	"github.com/stretchr/testify/assert"
)

func TestNewFileTransaction(t *testing.T) {
	createTestDB(t)
	createTestUser(t, testUserID, testUsername, testEmail)
	defer assertDBNotLocked(t)

	ft, err := NewFileTransaction(testUsername, testAlias)

	assert.NotNil(t, ft)
	assert.NoError(t, err)

	ft.SaveFile(testUserID, testAlias, testPath)

	buff := bytes.NewBuffer([]byte(testContent))

	assert.NoError(t, ft.SaveCommit(buff, &file.Commit{
		Hash:      testHash,
		Timestamp: 1,
		Message:   testMessage,
	}))

	assert.NoError(t, ft.Close())
}
