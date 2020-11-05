package db

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestSearch(t *testing.T) {
	createTestDB(t)
	initTestFile(t)

	t.Run("empty result with no query", func(t *testing.T) {
		table, err := SearchFiles(Connection, new(PageControls), nil)
		assert.NoError(t, err)
		assert.Empty(t, table.Rows)
		assert.Empty(t, table.TotalRows())
	})

	t.Run("ok with query", func(t *testing.T) {
		v := url.Values{}
		v.Set("q", testAlias)
		controls := &PageControls{Values: v}
		failIf(t, controls.Set(), "setting page controls for file test")
		table, err := SearchFiles(Connection, controls, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, table.Rows)
		assert.NotEmpty(t, table.TotalRows())
	})
}
