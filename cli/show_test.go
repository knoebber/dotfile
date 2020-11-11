package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShow(t *testing.T) {
	clearTestStorage(t)

	t.Run("error on attempt to show remote file without config set", func(t *testing.T) {
		showCommand := &showCommand{remote: true}
		assert.Error(t, showCommand.run(nil))
	})
	t.Run("error on attempt to show file that isn't set", func(t *testing.T) {
		showCommand := &showCommand{remote: true}
		assert.Error(t, showCommand.run(nil))
	})

	initTestFile(t)
	t.Run("ok", func(t *testing.T) {
		showCommand := &showCommand{alias: trackedFileAlias}
		assert.NoError(t, showCommand.run(nil))

	})

	t.Run("show data is ok", func(t *testing.T) {
		showCommand := &showCommand{alias: trackedFileAlias, data: true}
		assert.NoError(t, showCommand.run(nil))
	})
}
