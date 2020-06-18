package db

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilesTable(t *testing.T) {
	createTestDB(t)
	f := &File{
		UserID:          testUserID,
		Alias:           testAlias,
		Path:            testPath,
		CurrentRevision: testHash,
		Content:         []byte(testContent),
	}

	t.Run("has foreign key constraints", func(t *testing.T) {
		// Fails because user doesn't exist.
		_, err := insert(f, nil)
		assert.Error(t, err)

		createTestUser(t, testUserID, testUsername, testEmail)

		// User exists - ok.
		_, err = insert(f, nil)
		assert.NoError(t, err)

		// FK should restrict delete not cascade.
		_, err = connection.Exec("DELETE FROM users WHERE id = ?", testUserID)
		assert.Error(t, err)
	})

	// Reset DB.
	createTestDB(t)

	t.Run("alias must be unique between users", func(t *testing.T) {
		createTestFile(t)

		// Fails because alias already exists.
		_, err := insert(f, nil)
		assert.Error(t, err)

		// Error, alias is uppercased but distinctness should be case insensitive.
		f2 := *f
		f2.Alias = strings.Title(testAlias)
		f2.Path = "/different/path"
		_, err = insert(&f2, nil)
		assert.Error(t, err)
	})
}

func TestForkFile(t *testing.T) {
	otherUserID := int64(testUserID + 1)
	otherUsername := "user2"
	otherEmail := "user2@example.com"

	setup := func(t *testing.T) {
		createTestDB(t)

		createTestUser(t, otherUserID, otherUsername, otherEmail)
	}

	t.Run("forks with and without error", func(t *testing.T) {
		setup(t)
		f := initTestFile(t)
		assert.NoError(t, ForkFile(testUsername, testAlias, f.CurrentRevision, otherUserID))

		// Second fork is a duplicate alias error.
		assert.Error(t, ForkFile(testUsername, testAlias, f.CurrentRevision, otherUserID))
	})

	t.Run("fork copies commit revision content", func(t *testing.T) {
		setup(t)
		initialCommit, _ := initTestFileAndCommit(t)

		assert.NoError(t, ForkFile(testUsername, testAlias, initialCommit.Hash, otherUserID))
		f, err := GetFileByUsername(otherUsername, testAlias)
		failIf(t, err)
		assert.Equal(t, testContent, string(f.Content))
	})
}
