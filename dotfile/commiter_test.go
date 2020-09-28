package dotfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	t.Run("error on invalid alias", func(t *testing.T) {
		s := &MockStorer{}
		assert.Error(t, Init(s, "/valid/path", "alias cannot have spaces"))
	})

	t.Run("error on invalid path", func(t *testing.T) {
		s := &MockStorer{}
		assert.Error(t, Init(s, "/cant-be-directory/", "test"))
	})

	t.Run("error on new commit", func(t *testing.T) {
		s := &MockStorer{saveCommitErr: true}
		assert.Error(t, Init(s, "/valid/path", "test"))
	})

	t.Run("ok", func(t *testing.T) {
		s := &MockStorer{}
		assert.NoError(t, Init(s, "/valid/path", "test"))
	})
}

func TestNewCommit(t *testing.T) {
	t.Run("get contents error", func(t *testing.T) {
		s := &MockStorer{dirtyContentErr: true}
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
