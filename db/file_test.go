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

		createTestUser(t)

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
