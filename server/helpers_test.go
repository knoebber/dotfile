package server

import (
	"os"
	"testing"

	"github.com/knoebber/dotfile/db"
)

const (
	testDir = "testdata/"
)

func setupTest(t *testing.T) {
	os.RemoveAll(testDir)
	os.Mkdir(testDir, 0755)

	if err := db.Start(testDir + "dotfilehub.db"); err != nil {
		t.Fatalf("creating test db: %s", err)
	}

	if err := loadTemplates(); err != nil {
		t.Fatalf("loading templates: %v", err)
	}
}
