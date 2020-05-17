package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupStagedFile(t *testing.T) {
	t.Run("error when cannot get file", func(t *testing.T) {
		createTestDB(t)
		_, err := connection.Exec("DROP TABLE files")
		failIf(t, err)
		tx := getTestTransaction(t)
		_, err = setupStagedFile(tx, testUserID, testAlias)
		assert.Error(t, err)
	})

	t.Run("error when cannot get temp file", func(t *testing.T) {
		createTestDB(t)
		_, err := connection.Exec("DROP TABLE temp_files")
		failIf(t, err)
		tx := getTestTransaction(t)
		_, err = setupStagedFile(tx, testUserID, testAlias)
		assert.Error(t, err)
	})

	createTestDB(t)

	t.Run("no rows error", func(t *testing.T) {
		tx := getTestTransaction(t)
		_, err := setupStagedFile(tx, testUserID, testAlias)
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		tx := getTestTransaction(t)
		createTestTempFile(t)
		_, err := setupStagedFile(tx, testUserID, testAlias)
		assert.NoError(t, err)
	})
}
