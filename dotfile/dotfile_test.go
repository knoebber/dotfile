package dotfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlias(t *testing.T) {
	t.Run("lowercases given alias", func(t *testing.T) {
		alias, err := Alias("OK", "path")
		assert.NoError(t, err)
		assert.Equal(t, alias, "ok")
	})
	t.Run("creates expected aliases", func(t *testing.T) {
		for path, expected := range map[string]string{
			"file-name":                         "file-name",
			"file_nAME":                         "file_name",
			"~/.bashrc":                         "bashrc",
			".bash_profile":                     "bash_profile",
			"/etc/profile":                      "profile",
			"~/.config/i3/config":               "config",
			"~/.config/alacritty/alacritty.yml": "alacritty",
			"main.go":                           "main",
			"./main.go":                         "main",
			"./testdata/ok.txt":                 "ok",
			"/home/nicolas/projects/dotfile/.travis.yml": "travis",
			"/home/nicolas/.config/dotfile/dotfile.json": "dotfile",
			"/github.com/testdata/testfile.txt":          "testfile",
		} {
			alias, err := Alias("", path)
			assert.NoError(t, err)
			assert.Equal(t, expected, alias)
		}
	})

	t.Run("error with path that doesn't contain characters", func(t *testing.T) {
		_, err := Alias("", "/*/")
		assert.Error(t, err)
	})

}

func TestCheckAlias(t *testing.T) {
	t.Run("errors on non word characters", func(t *testing.T) {
		assert.Error(t, CheckAlias(" space "))
		assert.Error(t, CheckAlias("dash*dash"))
	})

	t.Run("ok", func(t *testing.T) {
		assert.NoError(t, CheckAlias("bashrc"))
		assert.NoError(t, CheckAlias("config_file"))
		assert.NoError(t, CheckAlias("config-file"))
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

func TestCompress(t *testing.T) {
	t.Run("error on empty ", func(t *testing.T) {
		_, err := Compress([]byte(""))
		assert.Error(t, err)
	})
}
