package db

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testDir = "testdata/"

func TestStart(t *testing.T) {
	// Tests that tables are able to be created without error.
	os.RemoveAll(testDir)
	os.Mkdir(testDir, 0755)

	assert.NoError(t, Start(testDir+"dotfilehub.db"))
}
