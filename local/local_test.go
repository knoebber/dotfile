package local

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultStorageDir(t *testing.T) {
	dir, err := DefaultStorageDir()
	assert.NotEmpty(t, dir)
	assert.NoError(t, err)
}

func TestList(t *testing.T) {
	setupTestFile(t)
	files, err := List(testDir, true)
	assert.NotEmpty(t, files)
	assert.NoError(t, err)
}
