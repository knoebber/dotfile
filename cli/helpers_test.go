package cli

import (
	"fmt"
	"os"
	"testing"

	"github.com/knoebber/dotfile/file"
	"github.com/stretchr/testify/assert"
)

const (
	nonExistantFile  = "file_does_not_exist"
	notTrackedFile   = "/dev/null"
	trackedFile      = "helpers_test.go"
	trackedFileAlias = "helpers_test"
	testDir          = "testdata/"
)

func getTestStorageClosure() func() (*file.Storage, error) {
	home, _ := getHome()
	dir := testDir
	name := defaultStorageName
	return getStorageClosure(home, &dir, &name)
}

func initTestFile(t *testing.T) {
	initCommand := &initCommand{
		fileName:   trackedFile,
		getStorage: getTestStorageClosure(),
	}
	err := initCommand.run(nil)
	assert.NoError(t, err)
}

func clearTestStorage() {
	os.Remove(fmt.Sprintf("%s%s", testDir, defaultStorageName))
}
