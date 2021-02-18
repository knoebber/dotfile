package db

import (
	"strings"
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
			_, err := insert(Connection, c)
			assert.Error(t, err)
		})

		t.Run("ok when file exists", func(t *testing.T) {
			fv := initTestFile(t)
			c.FileID = fv.ID
			_, err := insert(Connection, c)
			assert.NoError(t, err)
		})

		t.Run("fk restricts delete", func(t *testing.T) {
			_, err := Connection.Exec("DELETE FROM files")
			assert.Error(t, err)
		})
	})

	t.Run("forked from commit must exist", func(t *testing.T) {
		resetTestDB(t)

		fv := initTestFile(t)
		nonExistentCommitID := int64(69)
		c.FileID = fv.ID
		c.ForkedFrom = &nonExistentCommitID

		_, err := insert(Connection, c)
		assert.Error(t, err)
	})
}

func TestCommitRecord_check(t *testing.T) {
	createTestDB(t)

	c := new(CommitRecord)
	c.Message = strings.Repeat("c", maxCommitMessageSize+1)

	t.Run("error when message longer than max length", func(t *testing.T) {
		assert.Error(t, c.check(Connection))
	})
}
