package local

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfigPath(t *testing.T) {
	t.Run("creates in config dir", func(t *testing.T) {
		resetTestStorage(t)
		configDir := filepath.Join(testDir, ".config")

		assert.NoError(t, createDir(configDir))

		dotfileConfigDir := filepath.Join(configDir, "dotfile")
		expected := filepath.Join(dotfileConfigDir, "config.json")

		// Use testDir as home.
		actual, err := GetConfigPath(testDir)
		assert.NoError(t, err)
		assert.True(t, exists(dotfileConfigDir), "%s exists", dotfileConfigDir)
		assert.Equal(t, expected, actual)
	})

	t.Run("creates at top level", func(t *testing.T) {
		resetTestStorage(t)

		expected := filepath.Join(testDir, ".dotfile-config.json")
		// Use testDir as home.
		actual, err := GetConfigPath(testDir)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}
