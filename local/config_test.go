package local

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestDefaultConfigPath(t *testing.T) {
	path, err := DefaultConfigPath()
	assert.NoError(t, err)
	assert.NotEmpty(t, path)
}

func TestReadConfig(t *testing.T) {
	_ = os.Mkdir(testDir, 0755)
	testConfigPath := testDir + "test_config.json"

	t.Run("error when json is invalid", func(t *testing.T) {
		_ = os.Remove(testConfigPath)
		if err := ioutil.WriteFile(testConfigPath, []byte("bad json"), 0644); err != nil {
			t.Fatalf("setting up %s: %v", testTrackedFile, err)
		}

		_, err := ReadConfig(testConfigPath)
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		_ = os.Remove(testConfigPath)
		assert.NoError(t, SetConfig(testConfigPath, "username", "test"))

		config, err := ReadConfig(testConfigPath)
		assert.NoError(t, err)
		assert.NotEmpty(t, config)
		assert.Equal(t, config.Username, "test")
	})
}
