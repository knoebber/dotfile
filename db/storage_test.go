package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadFile(t *testing.T) {
	createTestDB(t)

	t.Run("no rows error", func(t *testing.T) {
		s, err := LoadFile(testUserID, testAlias)
		assert.Nil(t, s)
		assertErrNoRows(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		createTestFile(t)
		s, err := LoadFile(testUserID, testAlias)
		assert.NotNil(t, s)
		assert.NoError(t, err)
	})
}

func TestInitFile(t *testing.T) {
	createTestDB(t)

	t.Run("no rows error when no temp file", func(t *testing.T) {
		s, err := InitFile(testUserID, testAlias)
		assert.Nil(t, s)
		assertErrNoRows(t, err)
	})

	/*
		t.Run("ok", func(t *testing.T) {
			createTestTempFile(t)
			s, err := InitFile(testUserID, testAlias)
			assert.NotNil(t, s)
			assert.NoError(t, err)
		})
	*/
}
