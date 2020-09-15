package file

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUncompressRevision(t *testing.T) {
	t.Run("get revision error", func(t *testing.T) {
		s := &MockStorer{getRevisionErr: true}

		_, err := UncompressRevision(s, testHash)
		assert.Error(t, err)
	})

	t.Run("uncompress error", func(t *testing.T) {
		s := &MockStorer{uncompressErr: true}
		_, err := UncompressRevision(s, testHash)
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		_, err := UncompressRevision(new(MockStorer), testHash)
		assert.NoError(t, err)
	})

}

func TestDiff(t *testing.T) {
	t.Run("uncompress error", func(t *testing.T) {
		s := &MockStorer{uncompressErr: true}
		_, err := Diff(s, testHash, testHash)
		assert.Error(t, err)
	})

	t.Run("get content error", func(t *testing.T) {
		s := &MockStorer{getContentsErr: true}
		_, err := Diff(s, testHash, "")
		assert.Error(t, err)
	})

	t.Run("no changes error", func(t *testing.T) {
		s := &MockStorer{}
		_, err := Diff(s, testHash, testHash)
		assert.True(t, errors.Is(err, ErrNoChanges))
	})

}
