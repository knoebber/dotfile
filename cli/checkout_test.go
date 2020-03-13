package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckout(t *testing.T) {
	clearTestStorage()
	initTestFile(t)

	checkoutCommand := &checkoutCommand{
		getStorage: getTestStorageClosure(),
	}

	t.Run("returns error when file is not tracked", func(t *testing.T) {
		checkoutCommand.fileName = notTrackedFile
		assert.Error(t, checkoutCommand.run(nil))
	})

	t.Run("ok", func(t *testing.T) {
		checkoutCommand.fileName = trackedFileAlias
		assert.NoError(t, checkoutCommand.run(nil))
	})
}
