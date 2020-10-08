package cli

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	nonExistantFile         = "file_does_not_exist"
	notTrackedFile          = "/dev/null"
	testDir                 = "testdata/"
	trackedFile             = testDir + "testfile.txt"
	trackedFileAlias        = "testfile"
	initialTestFileContents = "Some stuff.\n"
	updatedTestFileContents = initialTestFileContents + "Some new content!\n"
)

func init() {
	flags = globalFlags{storageDir: testDir}
}

func initTestFile(t *testing.T) {
	os.Mkdir(testDir, 0755)
	writeTestFile(t, []byte(initialTestFileContents))
	fullPath, err := filepath.Abs(trackedFile)
	if err != nil {
		t.Fatalf("getting full path for %q: %v", trackedFile, err)
	}

	initCommand := &initCommand{path: fullPath}
	err = initCommand.run(nil)
	assert.NoError(t, err)
}

func updateTestFile(t *testing.T) {
	writeTestFile(t, []byte(updatedTestFileContents))
}

func clearTestStorage() {
	os.RemoveAll(testDir)
}

func writeTestFile(t *testing.T, contents []byte) {
	if err := ioutil.WriteFile(trackedFile, contents, 0644); err != nil {
		t.Fatalf("setting up %s: %v", trackedFile, err)
	}
}
