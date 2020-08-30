package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadTemplates(t *testing.T) {
	assert.NoError(t, loadTemplates())
}
