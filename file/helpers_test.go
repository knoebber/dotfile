package file

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testFile  = "storage_test.go"
	testAlias = "storage_test"
	testDir   = "testdata/"
	testName  = "files.json"
)

func clearTestStorage() {
	os.Remove(fmt.Sprintf("%s%s", testDir, testName))
}

func getTestStorage() *Storage {
	home, _ := os.UserHomeDir()
	s := &Storage{}
	s.Setup(home, testDir, testName)
	return s
}

func initTestFile(t *testing.T, s *Storage) {
	_, err := Init(s, testFile, testAlias)
	assert.NoError(t, err)
}
