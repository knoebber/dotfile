package local

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRelativePath(t *testing.T) {
	initTestdata(t)

	home, _ := os.UserHomeDir()

	t.Run("returns error when file does not exist", func(t *testing.T) {
		_, err := RelativePath("", "")
		assert.Error(t, err)
	})

	t.Run("returns error when file is not in home", func(t *testing.T) {
		_, err := RelativePath("/dev/null", testHome)
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		if home == "" {
			t.Log("failed to get home in testing environment - skipping test")
		}

		relativePath, err := RelativePath(testTrackedFile, home)
		assert.Contains(t, relativePath, "~")
		assert.NoError(t, err)
	})
}

func TestCreateIfNotExist(t *testing.T) {
	t.Run("returns error when cannot create directory", func(t *testing.T) {
		created, err := createIfNotExist(testDir+testDir+testDir, "")
		assert.False(t, created)
		assert.Error(t, err)
	})

	t.Run("returns error when cannot create file", func(t *testing.T) {
		clearTestStorage()

		created, err := createIfNotExist(testDir, testDir+testDir+testDir)
		assert.False(t, created)
		assert.Error(t, err)
	})

	t.Run("returns true no error", func(t *testing.T) {
		clearTestStorage()

		created, err := createIfNotExist(testDir, testTrackedFile)
		assert.True(t, created)
		assert.NoError(t, err)
	})

	t.Run("returns false no error", func(t *testing.T) {
		initTestdata(t)
		created, err := createIfNotExist(testDir, testTrackedFile)
		assert.False(t, created)
		assert.NoError(t, err)
	})

}
