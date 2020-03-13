package local

import (
	"io/ioutil"
	"os"
	"testing"
)

const (
	testFiles            = "files.json"
	testHome             = "/home/testing"
	testAlias            = "testalias"
	testHash             = "9abdbcf4ea4e2c1c077c21b8c2f2470ff36c31ce"
	nonExistantFile      = "file_does_not_exist"
	notTrackedFile       = "/dev/null"
	testDir              = "testdata/"
	testTrackedFile      = testDir + "testfile.txt"
	testTrackedFileAlias = "testfile"
	testContent          = "Some stuff.\n"
	updatedTestContent   = testContent + "Some new content!\n"
)

func initTestdata(t *testing.T) {
	_ = os.Mkdir(testDir, 0755)
	writeTestFile(t, []byte(testContent))
}

func updateTestFile(t *testing.T) {
	writeTestFile(t, []byte(updatedTestContent))
}

func clearTestStorage() {
	_ = os.RemoveAll(testDir)
}

func setupTestStorage() *Storage {
	clearTestStorage()
	s, _ := NewStorage(testHome, testDir, testFiles)
	return s
}

func writeTestFile(t *testing.T, contents []byte) {
	if err := ioutil.WriteFile(testTrackedFile, contents, 0644); err != nil {
		t.Fatalf("setting up %s: %v", testTrackedFile, err)
	}
}
