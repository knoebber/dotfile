package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFileTransaction(t *testing.T) {
	createTestDB(t)
	defer assertDBNotLocked(t)

	ft, err := NewFileTransaction(testUsername, testAlias)

	t.Run("ok when file does not exist", func(t *testing.T) {
		assert.NotNil(t, ft)
		assert.NoError(t, err)
		assert.NoError(t, ft.Close())
	})
}
