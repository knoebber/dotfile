package local

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/knoebber/dotfile/file"
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

func initTestData(t *testing.T) {
	_ = os.Mkdir(testDir, 0755)
	writeTestFile(t, []byte(testContent))
}

func updateTestFile(t *testing.T) {
	writeTestFile(t, []byte(updatedTestContent))
}

func clearTestStorage() {
	_ = os.RemoveAll(testDir)
}

func resetTestStorage(t *testing.T) {
	clearTestStorage()
	initTestData(t)
}

func setupTestFile(t *testing.T) *Storage {
	clearTestStorage()
	initTestData(t)

	fullPath, err := filepath.Abs(testTrackedFile)
	if err != nil {
		t.Fatalf("getting full path for %#v: %v", testTrackedFile, err)
	}

	s := &Storage{
		Home:     testHome,
		dir:      testDir,
		Alias:    testAlias,
		jsonPath: filepath.Join(testDir, testAlias+".json"),
		FileData: &file.TrackingData{
			Path:    fullPath,
			Commits: []file.Commit{},
		},
	}

	if err := file.Init(s, fullPath, testAlias); err != nil {
		t.Fatalf("initializing test file: %v", err)
	}

	// Read the newly initialized test file.
	if err := s.SetTrackingData(testAlias); err != nil {
		t.Fatalf("reading test file tracking data: %v", err)
	}

	if !s.HasFile {
		t.Fatal("expected storage to have file")
	}

	return s
}

func writeTestFile(t *testing.T, contents []byte) {
	if err := ioutil.WriteFile(testTrackedFile, contents, 0644); err != nil {
		t.Fatalf("setting up %s: %v", testTrackedFile, err)
	}
}
