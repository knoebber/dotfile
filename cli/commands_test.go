package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/alecthomas/kingpin.v2"
)

func TestAddCommandsToApplication(t *testing.T) {
	app := kingpin.New("dotfile", "version control optimized for single files")
	assert.NoError(t, AddCommandsToApplication(app))
}
