package local

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestDefaultStorageDir(t *testing.T) {
	dir, err := DefaultStorageDir()
	assert.NotEmpty(t, dir)
	assert.NoError(t, err)
}

func TestList(t *testing.T) {
	setupTestFile(t)
	files, err := List(testDir, true)
	assert.NotEmpty(t, files)
	assert.NoError(t, err)
}

func TestInitializeFile(t *testing.T) {
	t.Run("error when path cannot be converted to alias", func(t *testing.T) {
		_, err := InitializeFile(testDir, "%%%%", "")
		assert.Error(t, err)
	})

	t.Run("error when file is already tracked", func(t *testing.T) {
		setupTestFile(t)
		_, err := InitializeFile(testDir, testTrackedFile, testAlias)
		assert.Error(t, err)
	})

	t.Run("error when path cannot be converted", func(t *testing.T) {
		resetTestStorage(t)
		_, err := InitializeFile(testDir, "/does/not/exist", testAlias)
		assert.Error(t, err)
	})

	t.Run("error when alias is bad format", func(t *testing.T) {
		resetTestStorage(t)
		_, err := InitializeFile(testDir, testTrackedFile, "$$badchar$$")
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		resetTestStorage(t)
		s, err := InitializeFile(testDir, testTrackedFile, "")
		assert.NoError(t, err)
		assert.NotEmpty(t, s.Alias)
	})
}

func TestConvertPath(t *testing.T) {
	t.Run("error when $HOME isn't set", func(t *testing.T) {
		defer os.Setenv("HOME", os.Getenv("HOME"))
		_ = os.Unsetenv("HOME")
		_, err := convertPath("~/.bashrc")
		assert.Error(t, err)
	})

}
