package file

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAlias(t *testing.T) {
	t.Run("errors on non word characters", func(t *testing.T) {
		assert.Error(t, CheckAlias(" space "))
		assert.Error(t, CheckAlias("dash-dash"))
	})

	t.Run("ok", func(t *testing.T) {
		assert.NoError(t, CheckAlias("bashrc"))
		assert.NoError(t, CheckAlias("config_file"))
	})
}

func TestCheckPath(t *testing.T) {
	t.Run("error on empty", func(t *testing.T) {
		assert.Error(t, CheckPath(""))
	})
	t.Run("error on non absolute path", func(t *testing.T) {
		assert.Error(t, CheckPath("relative/file"))
	})
	t.Run("error on directory", func(t *testing.T) {
		for _, testcase := range []string{
			"/directory/path/",
			"/",
			"//",
			"~/Documents/",
		} {
			assert.Error(t, CheckPath(testcase), testcase)
		}
	})

	t.Run("ok", func(t *testing.T) {
		for _, testcase := range []string{
			"~/f",
			"~/.bashrc",
			"~/.config/nvim/init.vim",
			"/f",
			"/etc/aliases",
		} {
			assert.NoError(t, CheckPath(testcase), testcase)
		}

	})
}
