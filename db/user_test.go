package db

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsersTable(t *testing.T) {
	createTestDB(t)

	t.Run("username must be unique / case insensitive", func(t *testing.T) {
		createTestUser(t, testUserID, testUsername, testEmail)
		u := &User{
			Username:     strings.Title(testUsername),
			CLIToken:     testCliToken,
			PasswordHash: []byte(testPassword),
		}

		_, err := insert(u, nil)
		println(err.Error())
		assert.Error(t, err)

	})
}
