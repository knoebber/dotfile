package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {
	clearTestStorage()
	initTestFile(t)

	logCommand := new(logCommand)

	t.Run("returns error when file is not tracked", func(t *testing.T) {
		logCommand.fileName = notTrackedFile
		assert.Error(t, logCommand.run(nil))
	})

	t.Run("ok", func(t *testing.T) {
		logCommand.fileName = trackedFileAlias
		assert.NoError(t, logCommand.run(nil))
	})
}
