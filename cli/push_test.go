package cli

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPush(t *testing.T) {
	resetTestStorage(t)
	pc := &pushCommand{alias: "test"}
	t.Run("returns error when config not set", func(t *testing.T) {
		assert.Error(t, pc.run(nil))
	})
}
