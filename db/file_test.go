package db

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilesTable(t *testing.T) {
	createTestDB(t)
	f := &File{
		UserID: testUserID,
		Alias:  testAlias,
		Path:   testPath,
	}

	t.Run("has foreign key constraints", func(t *testing.T) {
		t.Run("fails when user doesnt exist", func(t *testing.T) {
			_, err := insert(f, nil)
			assert.Error(t, err)
		})

		t.Run("ok when user exists", func(t *testing.T) {
			createTestUser(t, testUserID, testUsername, testEmail)
			_, err := insert(f, nil)
			assert.NoError(t, err)
		})

		t.Run("foreign key restricts record delete", func(t *testing.T) {
			_, err := connection.Exec("DELETE FROM users WHERE id = ?", testUserID)
			assert.Error(t, err)
		})
	})

	// Reset DB.
	createTestDB(t)

	t.Run("alias must be unique between users", func(t *testing.T) {
		initTestFile(t)

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

	t.Run("fork works", func(t *testing.T) {
		setup(t)
		f := initTestFile(t)
		err := ForkFile(testUsername, testAlias, f.Hash, otherUserID)

		t.Run("no error", func(t *testing.T) {
			assert.NoError(t, err)
		})

		t.Run("error on attempt to fork again", func(t *testing.T) {
			assert.Error(t, ForkFile(testUsername, testAlias, f.Hash, otherUserID))
		})
		assertDBNotLocked(t)
	})

	t.Run("fork copies commit revision content", func(t *testing.T) {
		setup(t)
		initialCommit, _ := initTestFileAndCommit(t)

		assert.NoError(t, ForkFile(testUsername, testAlias, initialCommit.Hash, otherUserID))
		f, err := GetFile(otherUsername, testAlias)
		failIf(t, err)
		assert.Equal(t, testContent, string(f.Content))

		assertDBNotLocked(t)
	})
}

func TestSetFileToHash(t *testing.T) {
	createTestDB(t)
	initial, _ := initTestFileAndCommit(t)

	t.Run("error when hash does not exist", func(t *testing.T) {
		err := SetFileToHash(testUsername, testAlias, "doesnt exist")
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		err := SetFileToHash(testUsername, testAlias, initial.Hash)
		assert.NoError(t, err)
		f, err := GetFile(testUsername, testAlias)
		assert.NoError(t, err)
		assert.Equal(t, initial.Hash, f.Hash)
	})
}
