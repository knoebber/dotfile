package cli

import (
	"fmt"
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

// Based on https://npf.io/2015/06/testing-exec-command/

var sneakyTestingReference *testing.T

func init() {
	home, err := getHome()
	if err != nil {
		panic(fmt.Errorf("$HOME is required for cli tests: %v", err.Error()))
	}

	config = cliConfig{
		home:       home,
		storageDir: testDir,
	}
}

func initTestFile(t *testing.T) {
	os.Mkdir(testDir, 0755)
	writeTestFile(t, []byte(initialTestFileContents))
	fullPath, err := filepath.Abs(trackedFile)
	if err != nil {
		t.Fatalf("getting full path for %#v: %v", trackedFile, err)
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
