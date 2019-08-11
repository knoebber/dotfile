package cli

import (
	"fmt"
	"os"
	"testing"

	"github.com/knoebber/dotfile/file"
	"github.com/stretchr/testify/assert"
)

const (
	arbitraryFile = "helpers_test.go"
	testAlias     = "helpers_test"
	testDir       = "testdata/"
)

func getTestStorage() *file.Storage {
	return &file.Storage{
		Home: getHome(),
		Dir:  testDir,
		Name: defaultStorageName,
	}
}

func initTestFile(t *testing.T) {
	initCommand := &initCommand{
		fileName: arbitraryFile,
		storage:  getTestStorage(),
	}
	err := initCommand.run(nil)
	assert.NoError(t, err)
}

func clearTestStorage() {
	os.Remove(fmt.Sprintf("%s%s", testDir, defaultStorageName))
}
