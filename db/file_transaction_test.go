package db

import (
	"bytes"
	"testing"

	"github.com/knoebber/dotfile/dotfile"
	"github.com/stretchr/testify/assert"
)

func TestNewFileTransaction(t *testing.T) {
	createTestDB(t)
	createTestUser(t, testUserID, testUsername, testEmail)

	tx, err := Connection.Begin()
	assert.NoError(t, err)
	ft, err := NewFileTransaction(tx, testUserID, testAlias)

	assert.NotNil(t, ft)
	assert.NoError(t, err)

	_ = ft.SaveFile(testUserID, testAlias, testPath)

	buff := bytes.NewBuffer([]byte(testContent))

	assert.NoError(t, ft.SaveCommit(buff, &dotfile.Commit{
		Hash:      testHash,
		Timestamp: 1,
		Message:   testMessage,
	}))

	assert.NoError(t, tx.Commit())
}
