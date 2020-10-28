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

		_, err := insert(Connection, u)
		assert.Error(t, err)

	})

	t.Run("ok", func(t *testing.T) {
		_, err := CreateUser(Connection, "user1", testPassword)
		assert.NoError(t, err)

		_, err = CreateUser(Connection, "user2", testPassword)
		assert.NoError(t, err)
	})

}

func TestCheckPasswordResetToken(t *testing.T) {
	createTestDB(t)
	createTestUser(t, testUserID, testUsername, testEmail)

	t.Run("db error", func(t *testing.T) {
		tx, _ := Connection.Begin()
		_ = tx.Commit()

		_, err := CheckPasswordResetToken(tx, testPasswordResetToken)
		assert.Error(t, err)
	})

	t.Run("token not found error", func(t *testing.T) {
		_, err := CheckPasswordResetToken(Connection, "")
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		username, err := CheckPasswordResetToken(Connection, testPasswordResetToken)
		assert.NoError(t, err)
		assert.Equal(t, testUsername, username)
	})

	t.Run("multiple token error", func(t *testing.T) {
		createTestUser(t, testUserID+1, testUsername+"2", "other@example.com")
		_, err := CheckPasswordResetToken(Connection, testPasswordResetToken)
		assert.Error(t, err)
	})

}
