package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiff(t *testing.T) {
	clearTestStorage()
	initTestFile(t)

	diffCommand := new(diffCommand)

	t.Run("returns error when file is not tracked", func(t *testing.T) {
		diffCommand.fileName = notTrackedFile
		assert.Error(t, diffCommand.run(nil))
	})

	t.Run("ok", func(t *testing.T) {
		diffCommand.fileName = trackedFileAlias
		assert.NoError(t, diffCommand.run(nil))
	})
}
