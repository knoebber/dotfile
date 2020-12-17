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

	t.Run("does query", func(t *testing.T) {
		v := url.Values{}
		v.Set("q", testAlias)
		controls := &PageControls{Values: v}
		failIf(t, controls.Set(), "setting page controls for file test")

		t.Run("error", func(t *testing.T) {
			tx, _ := Connection.Begin()
			_ = tx.Commit()
			_, err := SearchFiles(tx, controls, nil)
			assert.Error(t, err)
		})

		t.Run("ok", func(t *testing.T) {
			table, err := SearchFiles(Connection, controls, nil)
			assert.NoError(t, err)
			assert.NotEmpty(t, table.Rows)
			assert.NotEmpty(t, table.TotalRows())
		})
	})
}

func TestFileFeed(t *testing.T) {
	createTestDB(t)
	initTestFile(t)

	t.Run("error with bad connection", func(t *testing.T) {
		tx, _ := Connection.Begin()
		_ = tx.Commit()
		_, err := FileFeed(tx, 0, nil)
		assert.Error(t, err)
	})

	t.Run("returns result", func(t *testing.T) {
		res, err := FileFeed(Connection, 1, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)
	})
}
