package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCommitsTable(t *testing.T) {
	createTestDB(t)
	c := &Commit{
		FileID:    testFileID,
		Hash:      testHash,
		Message:   testMessage,
		Revision:  []byte(testContent),
		Timestamp: time.Now().Unix(),
	}

	t.Run("has foreign key constraints", func(t *testing.T) {
		// Fails because file doesn't exist.
		_, err := insert(c, nil)
		assert.Error(t, err)

		createTestFile(t)

		// File exists - ok.
		_, err = insert(c, nil)
		assert.NoError(t, err)

		// FK should restrict delete not cascade.
		_, err = connection.Exec("DELETE FROM files WHERE id = ?", testFileID)
		assert.Error(t, err)
	})

	// Reset DB.
	createTestDB(t)

	t.Run("forked from commit must exist", func(t *testing.T) {
		createTestFile(t)
		nonExistentcommitID := int64(69)
		c.ForkedFrom = &nonExistentcommitID

		_, err := insert(c, nil)
		assert.Error(t, err)
	})
}
