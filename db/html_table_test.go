package db

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHTMLTable(t *testing.T) {
	t.Run("ok with defaults", func(t *testing.T) {
		controls := new(PageControls)
		failIf(t, controls.Set(), "setting empty page controls")
		table := &HTMLTable{Controls: controls}
		assert.Empty(t, table.Pages())
	})
}
