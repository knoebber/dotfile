package db

import (
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"os"
	"testing"
)

const (
	testDir      = "testdata/"
	testAlias    = "testalias"
	testPath     = "~/dotfile/test-file.txt"
	testUserID   = 1
	testContent  = "Testing content. Stored as a blob."
	testHash     = "9abdbcf4ea4e2c1c077c21b8c2f2470ff36c31ce"
	testUsername = "dotfile_user"
	testPassword = "ilovecatS!"
	testEmail    = "dot@dotfilehub.com"
	testCliToken = "12345678"
)

func createTestDB(t *testing.T) {
	os.RemoveAll(testDir)
	os.Mkdir(testDir, 0755)

	if err := Start(testDir + "dotfilehub.db"); err != nil {
		t.Fatalf("creating test db: %s", err)
	}
}

func createTestUser(t *testing.T) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("creating test password: %s", err)
	}

	_, err = connection.Exec(`
INSERT INTO users(id, username, email, password_hash, cli_token) 
VALUES(?, ?, ?, ?, ?)`,
		testUserID,
		testUsername,
		testEmail,
		hashed,
		testCliToken,
	)
	if err != nil {
		t.Fatalf("creating test user: %s", err)
	}
}

func createTestFile(t *testing.T) *File {
	createTestUser(t)

	testFile := &File{
		UserID:   testUserID,
		Alias:    testAlias,
		Path:     testPath,
		Revision: testHash,
		Content:  []byte(testContent),
	}
	id, err := insert(testFile, nil)
	if err != nil {
		t.Fatalf("creating test file: %s", err)
	}
	testFile.ID = id
	return testFile
}

func createTestTempFile(t *testing.T) *TempFile {
	createTestUser(t)

	testTempFile := &TempFile{
		UserID:  testUserID,
		Alias:   testAlias,
		Path:    testPath,
		Content: []byte(testContent),
	}
	id, err := insert(testTempFile, nil)
	if err != nil {
		t.Fatalf("creating test file: %s", err)
	}
	testTempFile.ID = id
	return testTempFile
}

func assertErrNoRows(t *testing.T, err error) {
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows, got error %s", err)
	}
}
