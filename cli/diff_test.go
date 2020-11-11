package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiff(t *testing.T) {
	clearTestStorage(t)
	initTestFile(t)

	diffCommand := new(diffCommand)

	t.Run("returns error when file is not tracked", func(t *testing.T) {
		diffCommand.alias = notTrackedFile
		assert.Error(t, diffCommand.run(nil))
	})

	t.Run("error when file has no differences", func(t *testing.T) {
		diffCommand.alias = trackedFileAlias
		assert.Error(t, diffCommand.run(nil))
	})
}
