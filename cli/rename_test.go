package cli

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRename(t *testing.T) {
	clearTestStorage(t)
	initTestFile(t)

	renameCommand := new(renameCommand)

	t.Run("returns error when file is not tracked", func(t *testing.T) {
		renameCommand.alias = notTrackedFile
		assert.Error(t, renameCommand.run(nil))
	})

	t.Run("ok", func(t *testing.T) {
		renameCommand.alias = trackedFileAlias
		renameCommand.newAlias = "new_name"
		assert.NoError(t, renameCommand.run(nil))

		// Assert that commands can be ran on the new alias.
		showCmd := &showCommand{alias: "new_name"}
		assert.NoError(t, showCmd.run(nil))
	})
}
