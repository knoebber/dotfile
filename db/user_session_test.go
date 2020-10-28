package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSession(t *testing.T) {
	createTestDB(t)
	createTestUser(t, testUserID, testUsername, testEmail)

	record, err := createSession(Connection, testUsername, "192.168.101")
	assert.NoError(t, err)
	assert.NotNil(t, record)

	session, err := Session(Connection, record.Session)
	assert.NoError(t, err)
	assert.NotNil(t, session)
}
