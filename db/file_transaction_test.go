package db

import (
	"bytes"
	"testing"

	"github.com/knoebber/dotfile/dotfile"
	"github.com/stretchr/testify/assert"
)

func testNewFileTransaction(t *testing.T) *FileTransaction {
	createTestDB(t)
	createTestUser(t, testUserID, testUsername, testEmail)

	tx, err := Connection.Begin()
	assert.NoError(t, err)
	ft, err := NewFileTransaction(tx, testUserID, testAlias)

	assert.NotNil(t, ft)
	assert.NoError(t, err)

	return ft
}

func TestFileTransaction_SaveFile(t *testing.T) {
	ft := testNewFileTransaction(t)

	t.Run("error when no values are provided", func(t *testing.T) {
		assert.Error(t, ft.SaveFile(0, "", ""))
	})

	t.Run("ok", func(t *testing.T) {
		assert.NoError(t, ft.SaveFile(testUserID, testAlias, testPath))
	})

	assert.NoError(t, ft.tx.Commit())
}

func TestFileTransaction_SaveCommit(t *testing.T) {
	ft := testNewFileTransaction(t)
	c := &dotfile.Commit{
		Hash:      testHash,
		Timestamp: 1,
		Message:   testMessage,
	}
	buff := bytes.NewBuffer([]byte(testContent))

	t.Run("error with no file association", func(t *testing.T) {
		assert.Error(t, ft.SaveCommit(buff, c))
	})

	t.Run("ok", func(t *testing.T) {
		assert.NoError(t, ft.SaveFile(testUserID, testAlias, testPath))
		assert.NoError(t, ft.SaveCommit(buff, c))
	})

	assert.NoError(t, ft.tx.Commit())
}
