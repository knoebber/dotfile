package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	clearTestStorage(t)

	initCommand := new(initCommand)

	t.Run("returns error when file does not exist", func(t *testing.T) {
		initCommand.path = nonExistantFile
		assert.Error(t, initCommand.run(nil))
	})

	t.Run("ok", func(t *testing.T) {
		initTestFile(t)
	})

}
