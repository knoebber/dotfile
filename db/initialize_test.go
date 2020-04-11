package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	assert.NoError(t, Start("testdata/dotfilehub.db"))
}
