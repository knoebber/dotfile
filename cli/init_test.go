package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	clearTestStorage()

	initCommand := new(initCommand)

	t.Run("returns error when file does not exist", func(t *testing.T) {
		initCommand.path = nonExistantFile
		assert.Error(t, initCommand.run(nil))
	})

	t.Run("no error when file exists", func(t *testing.T) {
		os.Mkdir(testDir, 0755)
		writeTestFile(t, []byte(initialTestFileContents))
		initCommand.path = trackedFile
		assert.NoError(t, initCommand.run(nil))
	})
}
