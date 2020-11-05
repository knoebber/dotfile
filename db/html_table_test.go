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

	t.Run("returns columns", func(t *testing.T) {
		controls := new(PageControls)
		failIf(t, controls.Set(), "setting empty page controls")
		table := &HTMLTable{Controls: controls, Columns: []string{"ok"}}
		assert.NotEmpty(t, table.Header())
	})

	t.Run("returns pages", func(t *testing.T) {
		controls := &PageControls{totalRows: 1000, page: 9}
		failIf(t, controls.Set(), "setting empty page controls")
		table := &HTMLTable{Controls: controls, Columns: []string{"ok"}}
		assert.NotEmpty(t, table.Pages())
	})

}
