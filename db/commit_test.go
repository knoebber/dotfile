package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCommitsTable(t *testing.T) {
	createTestDB(t)

	c := &CommitRecord{
		Hash:      testHash,
		Message:   testMessage,
		Revision:  []byte(testContent),
		Timestamp: time.Now().Unix(),
	}

	t.Run("has foreign key constraints", func(t *testing.T) {
		// Fails because file doesn't exist.
		t.Run("fails when file doesnt exist", func(t *testing.T) {
			_, err := insert(c, nil)
			assert.Error(t, err)
		})

		t.Run("ok when file exists", func(t *testing.T) {
			fv := initTestFile(t)
			c.FileID = fv.ID
			_, err := insert(c, nil)
			assert.NoError(t, err)
		})

		t.Run("fk restricts delete", func(t *testing.T) {
			_, err := connection.Exec("DELETE FROM files")
			assert.Error(t, err)
		})
	})

	t.Run("forked from commit must exist", func(t *testing.T) {
		createTestDB(t)

		fv := initTestFile(t)
		nonExistentcommitID := int64(69)
		c.FileID = fv.ID
		c.ForkedFrom = &nonExistentcommitID

		_, err := insert(c, nil)
		assert.Error(t, err)
	})
}
