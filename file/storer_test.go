package file

import (
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

func TestInit(t *testing.T) {
	t.Run("error on new commit", func(t *testing.T) {
		s := &MockStorer{saveCommitErr: true}
		assert.Error(t, Init(s))
	})
}

func TestNewCommit(t *testing.T) {
	t.Run("get contents error", func(t *testing.T) {
		s := &MockStorer{getContentsErr: true}
		err := NewCommit(s, testMessage)
		assert.Error(t, err)
	})

	t.Run("has commit error", func(t *testing.T) {
		s := &MockStorer{hasCommitErr: true}
		err := NewCommit(s, testMessage)
		assert.Error(t, err)
	})

	t.Run("already exists error", func(t *testing.T) {
		s := &MockStorer{hasCommit: true}
		err := NewCommit(s, testMessage)
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		s := &MockStorer{}
		err := NewCommit(s, testMessage)
		assert.NoError(t, err)
	})
}

func TestCheckout(t *testing.T) {
	t.Run("has commit error", func(t *testing.T) {
		s := &MockStorer{hasCommitErr: true}
		err := Checkout(s, testHash)
		assert.Error(t, err)
	})

	t.Run("revision does not exist error", func(t *testing.T) {
		s := &MockStorer{}
		err := Checkout(s, testHash)
		assert.Error(t, err)
	})

	t.Run("revert error", func(t *testing.T) {
		s := &MockStorer{hasCommit: true, revertErr: true}
		err := Checkout(s, testHash)
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		s := &MockStorer{hasCommit: true}
		err := Checkout(s, testHash)
		assert.NoError(t, err)
	})
}
