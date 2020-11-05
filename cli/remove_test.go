package cli

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRemove(t *testing.T) {
	clearTestStorage()
	initTestFile(t)

	removeCommand := new(removeCommand)

	t.Run("returns error when file is not tracked", func(t *testing.T) {
		removeCommand.alias = notTrackedFile
		assert.Error(t, removeCommand.run(nil))
	})

	t.Run("ok", func(t *testing.T) {
		removeCommand.alias = trackedFileAlias
		assert.NoError(t, removeCommand.run(nil))

		// Assert that using alias throws an error now.
		showCmd := &showCommand{alias: trackedFileAlias}
		assert.Error(t, showCmd.run(nil))
	})
}
