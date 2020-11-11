package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommit(t *testing.T) {
	clearTestStorage(t)
	initTestFile(t)

	commitCommand := &commitCommand{
		commitMessage: "test commit",
	}

	t.Run("returns error when file is not tracked", func(t *testing.T) {
		commitCommand.alias = notTrackedFile
		assert.Error(t, commitCommand.run(nil))
	})

	t.Run("ok", func(t *testing.T) {
		updateTestFile(t)
		commitCommand.alias = trackedFileAlias
		assert.NoError(t, commitCommand.run(nil))
	})
}
