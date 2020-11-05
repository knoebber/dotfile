package cli

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestMove(t *testing.T) {
	clearTestStorage()
	initTestFile(t)

	newPath := filepath.Join(testDir, "newpath.txt")
	moveCommand := &moveCommand{newPath: newPath}

	t.Run("returns error when file is not tracked", func(t *testing.T) {
		moveCommand.alias = notTrackedFile
		assert.Error(t, moveCommand.run(nil))
	})

	t.Run("ok", func(t *testing.T) {
		moveCommand.alias = trackedFileAlias
		assert.NoError(t, moveCommand.run(nil))

		// Assert that the new path exists.
		_, err := os.Stat(newPath)
		assert.NoError(t, err)
	})
}
