package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/alecthomas/kingpin.v2"
)

func TestAddCommandsToApplication(t *testing.T) {
	app := kingpin.New("dotfile", "version control optimized for single files")
	t.Run("error", func(t *testing.T) {
		defer os.Setenv("HOME", os.Getenv("HOME"))
		_ = os.Unsetenv("HOME")

		assert.Error(t, AddCommandsToApplication(app))
	})

	t.Run("ok", func(t *testing.T) {
		assert.NoError(t, AddCommandsToApplication(app))
	})
}
