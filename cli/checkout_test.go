package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckout(t *testing.T) {
	clearTestStorage(t)
	initTestFile(t)

	checkoutCommand := new(checkoutCommand)

	t.Run("returns error when file is not tracked", func(t *testing.T) {
		checkoutCommand.alias = notTrackedFile
		assert.Error(t, checkoutCommand.run(nil))
	})

	t.Run("ok", func(t *testing.T) {
		checkoutCommand.alias = trackedFileAlias
		assert.NoError(t, checkoutCommand.run(nil))
	})
}
