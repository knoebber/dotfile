package db

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	os.Mkdir("testdata/", 0755)

	assert.NoError(t, Start("testdata/dotfilehub.db"))
}
