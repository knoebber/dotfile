package db

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	createTestDB(t)

	t.Run("username must be unique / case insensitive", func(t *testing.T) {
		createTestUser(t, testUserID, testUsername, testEmail)
		u := &UserRecord{
			Username:     strings.Title(testUsername),
			CLIToken:     testCliToken,
			PasswordHash: []byte(testPassword),
		}

		_, err := insert(u, nil)
		assert.Error(t, err)

	})

	t.Run("ok", func(t *testing.T) {
		_, err := CreateUser("user1", "testpassword")
		assert.NoError(t, err)

		_, err = CreateUser("user2", "testpassword")
		assert.NoError(t, err)
	})

}
