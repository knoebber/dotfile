package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	clearTestStorage()

	initCommand := &initCommand{
		getStorage: getTestStorageClosure(),
	}

	t.Run("returns error when file does not exist", func(t *testing.T) {
		initCommand.fileName = nonExistantFile
		assert.Error(t, initCommand.run(nil))
	})

	t.Run("no error when file exists", func(t *testing.T) {
		writeTestFile(t, []byte(initialTestFileContents))
		initCommand.fileName = trackedFile
		assert.NoError(t, initCommand.run(nil))
	})
}
