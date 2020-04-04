package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/knoebber/dotfile/local"
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

func getTestStorageClosure() func() (*local.Storage, error) {
	home, _ := getHome()
	dir := testDir
	name := defaultStorageName
	return getStorageClosure(home, &dir, &name)
}

func initTestFile(t *testing.T) {
	os.Mkdir(testDir, 0755)
	writeTestFile(t, []byte(initialTestFileContents))
	initCommand := &initCommand{
		fileName:   trackedFile,
		getStorage: getTestStorageClosure(),
	}
	err := initCommand.run(nil)
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
