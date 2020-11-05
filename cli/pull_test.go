package cli

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPull(t *testing.T) {
	resetTestStorage()
	pc := &pullCommand{alias: "test"}
	t.Run("returns error when config not set", func(t *testing.T) {
		assert.Error(t, pc.run(nil))
	})
}
