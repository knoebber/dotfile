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

func getTestData() *file.Data {
	return &file.Data{
		Home: getHome(),
		Dir:  testDir,
		Name: defaultDataName,
	}
}

func initTestFile(t *testing.T) {
	initCommand := &initCommand{
		fileName: arbitraryFile,
		data:     getTestData(),
	}
	err := initCommand.run(nil)
	assert.NoError(t, err)
}

func clearTestData() {
	os.Remove(fmt.Sprintf("%s%s", testDir, defaultDataName))
}
