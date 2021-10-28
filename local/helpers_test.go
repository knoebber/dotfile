package local

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/knoebber/dotfile/dotfile"
)

const (
	testAlias          = "testalias"
	testMessage        = "test message"
	testHash           = "9abdbcf4ea4e2c1c077c21b8c2f2470ff36c31ce"
	testUpdatedHash    = "5d12fbbc6038e0b6a3e798dd790512ba03de7b6a"
	testDir            = "testdata/"
	testConfigPath     = testDir + "test_config.json"
	testTrackedFile    = testDir + "testfile.txt"
	testContent        = "Some stuff.\n"
	testUpdatedContent = testContent + "Some new content.\nNew lines!\n"
)

func initTestData(t *testing.T) {
	_ = os.Mkdir(testDir, 0755)
	writeTestFile(t, []byte(testContent))
}

func updateTestFile(t *testing.T) {
	writeTestFile(t, []byte(testUpdatedContent))
}

func clearTestStorage() {
	_ = os.RemoveAll(testDir)
}

func resetTestStorage(t *testing.T) {
	clearTestStorage()
	initTestData(t)
}

func testStorage() *Storage {
	return &Storage{
		Dir:   testDir,
		Alias: testAlias,
	}
}

func setupTestFile(t *testing.T) *Storage {
	clearTestStorage()
	initTestData(t)

	fullPath, err := filepath.Abs(testTrackedFile)
	if err != nil {
		t.Fatalf("getting full path for %q: %v", testTrackedFile, err)
	}

	s := testStorage()
	s.FileData = &dotfile.TrackingData{
		Path:    fullPath,
		Commits: []dotfile.Commit{},
	}

	if err := dotfile.Init(s, fullPath, testAlias); err != nil {
		t.Fatalf("initializing test file: %v", err)
	}

	// Read the newly initialized test file.
	if err := s.SetTrackingData(); err != nil {
		t.Fatalf("reading test file tracking data: %v", err)
	}

	return s
}

func writeTestFile(t *testing.T, contents []byte) {
	if err := os.WriteFile(testTrackedFile, contents, 0644); err != nil {
		t.Fatalf("setting up %s: %v", testTrackedFile, err)
	}
}
