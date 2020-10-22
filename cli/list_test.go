package cli

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestList(t *testing.T) {

	t.Run("error on attempt to list remote files without config set", func(t *testing.T) {
		listCommand := &listCommand{remote: true}
		assert.Error(t, listCommand.run(nil))
	})

	// Running test remote commands creates a config.json file in the same directory (testdata/) as tracking files.
	// This will break the local list.
	clearTestStorage()
	initTestFile(t)

	t.Run("ok", func(t *testing.T) {
		listCommand := new(listCommand)
		assert.NoError(t, listCommand.run(nil))
	})
	t.Run("ok with path", func(t *testing.T) {
		listCommand := &listCommand{path: true}
		assert.NoError(t, listCommand.run(nil))
	})

}
