package server

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("error when database directory doesnt exist", func(t *testing.T) {
		_, err := New(Config{DBPath: "/not/exist"})
		assert.Error(t, err)
	})

	t.Run("error when smtp path doesnt exist", func(t *testing.T) {
		_, err := New(Config{SMTPConfigPath: "smtp.json"})
		assert.Error(t, err)
	})

	t.Run("ok with default config", func(t *testing.T) {
		s, err := New(Config{})
		assert.NotNil(t, s)
		assert.NoError(t, err)
		assert.NoError(t, s.Close())
	})
}
