package cli

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestForget(t *testing.T) {
	clearTestStorage(t)
	initTestFile(t)

	forgetCommand := &forgetCommand{}

	t.Run("returns error when file is not tracked", func(t *testing.T) {
		forgetCommand.alias = notTrackedFile
		assert.Error(t, forgetCommand.run(nil))
	})

	t.Run("ok", func(t *testing.T) {
		forgetCommand.alias = trackedFileAlias
		assert.NoError(t, forgetCommand.run(nil))

		// Assert that the json file does not exist.
		_, err := os.Stat(filepath.Join(testDir, trackedFileAlias+".json"))
		assert.Error(t, err)
	})
}
