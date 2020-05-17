package db

import (
	"testing"

	"bytes"
	"github.com/knoebber/dotfile/file"
	"github.com/stretchr/testify/assert"
	"time"
)

func TestNewStorage(t *testing.T) {
	createTestDB(t)

	t.Run("no rows error", func(t *testing.T) {
		s, err := NewStorage(testUserID, testAlias)
		assert.Nil(t, s)
		assertErrNoRows(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		createTestFile(t)
		s, err := NewStorage(testUserID, testAlias)
		failIf(t, err)
		assert.NoError(t, s.Close())
	})

	t.Run("file.Init works", func(t *testing.T) {
		createTestTempFile(t)
		s, err := NewStorage(testUserID, testAlias)
		failIf(t, err)
		failIf(t, file.Init(s))
		assert.NoError(t, s.Close())
	})
}

func TestGetContents(t *testing.T) {
	t.Run("error when no dirty content", func(t *testing.T) {
		s := &Storage{
			staged: new(stagedFile),
		}
		_, err := s.GetContents()
		assert.Error(t, err)
	})
}

func TestGetRevision(t *testing.T) {
	createTestDB(t)

	t.Run("returns error when commit does not exist", func(t *testing.T) {
		s := getTestStorage(t)
		_, err := s.GetRevision(testHash)
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		createTestCommit(t)
		s := getTestStorage(t)
		revision, err := s.GetRevision(testHash)
		failIf(t, err)
		assert.Equal(t, testRevision, string(revision))
	})
}

func TestSaveCommit(t *testing.T) {
	createTestDB(t)

	t.Run("returns error on duplicate commit", func(t *testing.T) {
		createTestCommit(t)
		s := getTestStorage(t)

		err := s.SaveCommit(
			bytes.NewBuffer([]byte(testContent)),
			testHash,
			testMessage,
			time.Now(),
		)

		assert.Error(t, err)
	})

	t.Run("error when buffer is empty", func(t *testing.T) {
		s := getTestStorage(t)

		err := s.SaveCommit(
			bytes.NewBuffer([]byte{}),
			testHash,
			testMessage,
			time.Now(),
		)

		assert.Error(t, err)

	})
}
