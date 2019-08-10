package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const nonExistantFile = "this does not exist"

func TestInit(t *testing.T) {
	clearTestData()

	initCommand := &initCommand{
		data: getTestData(),
	}

	t.Run("returns error when file does not exist", func(t *testing.T) {
		initCommand.fileName = nonExistantFile
		assert.Error(t, initCommand.run(nil))
	})

	t.Run("no error when file exists", func(t *testing.T) {
		initCommand.fileName = arbitraryFile
		assert.NoError(t, initCommand.run(nil))
	})
}
