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
		createTestTempFile(t, testContent)
		s, err := NewStorage(testUserID, testAlias)
		failIf(t, err)
		failIf(t, file.Init(s, testPath, testAlias))
		assert.NoError(t, s.Close())
	})
}

func TestGetContents(t *testing.T) {
	t.Run("error when no dirty content", func(t *testing.T) {
		s := &Storage{
			Staged: new(stagedFile),
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

func TestRevert(t *testing.T) {
	createTestDB(t)
	initTestFile(t)

	f, err := getFileByUserID(testUserID, testAlias)
	failIf(t, err, "getting file by user ID")
	initialCommitHash := f.CurrentRevision

	// Stage updated test content.
	createTestTempFile(t, testUpdatedContent)
	s := getTestStorage(t)

	failIf(t, file.NewCommit(s, "Testing revert; updating to new content"), "creating test temp file")
	failIf(t, s.Close(), "closing storage for commit setup")

	f, err = getFileByUserID(testUserID, testAlias)
	failIf(t, err, "getting file by user ID")
	assert.Equal(t, testUpdatedContent, string(f.Content), "commit set file to updated content")

	s = getTestStorage(t)
	failIf(t, file.Checkout(s, initialCommitHash), "reverting file to", initialCommitHash)
	failIf(t, s.Close(), "closing storage after checkout")

	f, err = getFileByUserID(testUserID, testAlias)
	assert.Equal(t, testContent, string(f.Content), "reverted to content from initial commit")

}
