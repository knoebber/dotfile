package db

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilesTable(t *testing.T) {
	createTestDB(t)
	f := &FileRecord{
		UserID: testUserID,
		Alias:  testAlias,
		Path:   testPath,
	}

	t.Run("has foreign key constraints", func(t *testing.T) {
		t.Run("fails when user doesnt exist", func(t *testing.T) {
			_, err := insert(Connection, f)
			assert.Error(t, err)
		})

		t.Run("ok when user exists", func(t *testing.T) {
			createTestUser(t, testUserID, testUsername, testEmail)
			_, err := insert(Connection, f)
			assert.NoError(t, err)
		})

		t.Run("foreign key restricts record delete", func(t *testing.T) {
			_, err := Connection.Exec("DELETE FROM users WHERE id = ?", testUserID)
			assert.Error(t, err)
		})
	})

	t.Run("alias must be unique between users", func(t *testing.T) {
		resetTestDB(t)
		initTestFile(t)

		// Fails because alias already exists.
		_, err := insert(Connection, f)
		assert.Error(t, err)

		// Error, alias is uppercased but distinctness should be case insensitive.
		f2 := *f
		f2.Alias = strings.Title(testAlias)
		f2.Path = "/different/path"
		_, err = insert(Connection, &f2)
		assert.Error(t, err)
	})

	t.Run("alias cant be too long", func(t *testing.T) {
		resetTestDB(t)
		createTestUser(t, testUserID, testUsername, testEmail)
		f.Alias = strings.Repeat("a", maxStringSize+1)
		_, err := insert(Connection, f)
		assertUsererror(t, err)
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
		tx := testTransaction(t)
		f := initTestFile(t)
		err := ForkFile(tx, testUsername, testAlias, f.Hash, otherUserID)

		assert.NoError(t, err)
		assert.NoError(t, tx.Commit())

		// Error on attempt to fork again.
		tx = testTransaction(t)

		err = ForkFile(tx, testUsername, testAlias, f.Hash, otherUserID)
		assertUsererror(t, err)
		assert.NoError(t, tx.Rollback())
	})

	t.Run("fork copies commit revision content", func(t *testing.T) {
		setup(t)
		tx := testTransaction(t)
		initialCommit, _ := initTestFileAndCommit(t)

		assert.NoError(t, ForkFile(tx, testUsername, testAlias, initialCommit.Hash, otherUserID))
		assert.NoError(t, tx.Commit())
		f, err := UncompressFile(Connection, otherUsername, testAlias)
		failIf(t, err)
		assert.Equal(t, testContent, string(f.Content))
	})
}

func TestSetFileToHash(t *testing.T) {
	createTestDB(t)
	initial, _ := initTestFileAndCommit(t)

	t.Run("error when hash does not exist", func(t *testing.T) {
		err := SetFileToHash(Connection, testUsername, testAlias, "doesnt exist")
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		err := SetFileToHash(Connection, testUsername, testAlias, initial.Hash)
		assert.NoError(t, err)
		f, err := UncompressFile(Connection, testUsername, testAlias)
		assert.NoError(t, err)
		assert.Equal(t, initial.Hash, f.Hash)
	})
}
