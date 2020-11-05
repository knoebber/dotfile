package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLink_Class(t *testing.T) {
	t.Run("active", func(t *testing.T) {
		l := newLink("test", "test", "test")
		assert.NotEmpty(t, l.Class())
	})
	t.Run("not active", func(t *testing.T) {
		l := newLink("test", "test", "")
		assert.Empty(t, l.Class())
	})
}
