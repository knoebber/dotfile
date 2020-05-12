package local

import (
	"io/ioutil"
	"os"
	"testing"
)

const (
	testHome           = "/home/testing"
	testAlias          = "testalias"
	testMessage        = "test message"
	testHash           = "9abdbcf4ea4e2c1c077c21b8c2f2470ff36c31ce"
	testUpdatedHash    = "5d12fbbc6038e0b6a3e798dd790512ba03de7b6a"
	nonExistantFile    = "file_does_not_exist"
	notTrackedFile     = "/dev/null"
	testDir            = "testdata/"
	testTrackedFile    = testDir + "testfile.txt"
	testContent        = "Some stuff.\n"
	updatedTestContent = testContent + "Some new content!\n"
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

func setupTestFile(t *testing.T) *Storage {
	clearTestStorage()
	os.Mkdir(testDir, 0755)
	writeTestFile(t, []byte(testContent))

	_, err := InitFile(testHome, testDir, testTrackedFile, testAlias)
	if err != nil {
		t.Fatalf("initializing test file: %s", err)
	}

	s, err := LoadFile(testHome, testDir, testAlias)
	if err != nil {
		t.Fatalf("loading test file: %s", err)
	}

	return s
}

func writeTestFile(t *testing.T, contents []byte) {
	if err := ioutil.WriteFile(testTrackedFile, contents, 0644); err != nil {
		t.Fatalf("setting up %s: %v", testTrackedFile, err)
	}
}
