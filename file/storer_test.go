package file

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustGetTracked(t *testing.T) {
	t.Run("get tracked error", func(t *testing.T) {
		s := &MockStorer{getTrackedErr: true}

		_, err := MustGetTracked(s, testAlias)
		assert.Error(t, err)
	})

	t.Run("not tracked error", func(t *testing.T) {
		s := &MockStorer{testAliasNotTracked: true}

		_, err := MustGetTracked(s, testAlias)
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		tf, err := MustGetTracked(new(MockStorer), testAlias)
		assert.NoError(t, err)
		assert.NotNil(t, tf)
	})

}

func TestUncompressRevision(t *testing.T) {
	t.Run("get revision error", func(t *testing.T) {
		s := &MockStorer{getRevisionErr: true}

		_, err := UncompressRevision(s, testAlias, testHash)
		assert.Error(t, err)
	})

	t.Run("uncompress error", func(t *testing.T) {
		_, err := UncompressRevision(new(MockStorer), testAlias, testHash)
		assert.Error(t, err)
	})
}

func TestInit(t *testing.T) {
	t.Run("path to alias error", func(t *testing.T) {
		_, err := Init(new(MockStorer), invalidPath, "")
		assert.Error(t, err)
	})

	t.Run("get tracked error", func(t *testing.T) {
		s := &MockStorer{getTrackedErr: true}
		_, err := Init(s, testPath, testAlias)
		assert.Error(t, err)
	})

	t.Run("file is already tracked error", func(t *testing.T) {
		_, err := Init(new(MockStorer), testPath, testAlias)
		assert.Error(t, err)
	})

	t.Run("get contents error", func(t *testing.T) {
		s := &MockStorer{testAliasNotTracked: true, getContentsErr: true}
		_, err := Init(s, testPath, testAlias)
		assert.Error(t, err)
	})

	t.Run("save revision error", func(t *testing.T) {
		s := &MockStorer{testAliasNotTracked: true, saveRevisionErr: true}
		_, err := Init(s, testPath, testAlias)
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		s := &MockStorer{testAliasNotTracked: true}
		_, err := Init(s, testPath, testAlias)
		assert.NoError(t, err)
	})

}

func TestNewCommit(t *testing.T) {
	t.Run("get tracked error", func(t *testing.T) {
		s := &MockStorer{getTrackedErr: true}
		err := NewCommit(s, testPath, testMessage)
		assert.Error(t, err)
	})

	t.Run("save revision error", func(t *testing.T) {
		s := &MockStorer{saveRevisionErr: true}
		err := NewCommit(s, testPath, testMessage)
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		err := NewCommit(new(MockStorer), testPath, testMessage)
		assert.NoError(t, err)
	})
}

func TestCheckout(t *testing.T) {
	t.Run("get tracked error", func(t *testing.T) {
		s := &MockStorer{getTrackedErr: true}
		err := Checkout(s, testPath, testHash)
		assert.Error(t, err)
	})

	t.Run("get revision error", func(t *testing.T) {
		s := &MockStorer{getRevisionErr: true}
		err := Checkout(s, testPath, testHash)
		assert.Error(t, err)
	})
}
