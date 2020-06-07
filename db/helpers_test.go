package db

import (
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/knoebber/dotfile/file"
	"golang.org/x/crypto/bcrypt"
)

const (
	testDir            = "testdata/"
	testAlias          = "testalias"
	testPath           = "~/dotfile/test-file.txt"
	testFileID         = 1
	testUserID         = 1
	testContent        = "Testing content. Stored as a blob."
	testUpdatedContent = testContent + "\n New content!\n"
	testRevision       = "Commit revision contents"
	testHash           = "9abdbcf4ea4e2c1c077c21b8c2f2470ff36c31ce"
	testMessage        = "commit message"
	testUsername       = "genericusername"
	testPassword       = "ilovecatS!"
	testEmail          = "dot@dotfilehub.com"
	testCliToken       = "12345678"
)

func createTestDB(t *testing.T) {
	os.RemoveAll(testDir)
	os.Mkdir(testDir, 0755)

	if err := Start(testDir + "dotfilehub.db"); err != nil {
		t.Fatalf("creating test db: %s", err)
	}
}

func createTestUser(t *testing.T) {
	var count int

	err := connection.
		QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", testUserID).
		Scan(&count)
	if err != nil {
		t.Fatalf("counting test users: %s", err)
	}
	if count > 0 {
		return
	}

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

func createTestFile(t *testing.T) {
	createTestUser(t)
	var count int

	err := connection.
		QueryRow("SELECT COUNT(*) FROM files WHERE id = ?", testFileID).
		Scan(&count)
	if err != nil {
		t.Fatalf("counting test files: %s", err)
	}
	if count > 0 {
		return
	}

	_, err = connection.Exec(`
INSERT INTO files(id, user_id, alias, path, current_revision, content)
VALUES(?, ?, ?, ?, ?, ?)`,
		testFileID,
		testUserID,
		testAlias,
		testPath,
		testHash,
		[]byte(testContent),
	)
	if err != nil {
		t.Fatalf("creating test file: %s", err)
	}
}

func createTestTempFile(t *testing.T, content string) *TempFile {
	createTestUser(t)

	testTempFile := &TempFile{
		UserID:  testUserID,
		Alias:   testAlias,
		Path:    testPath,
		Content: []byte(content),
	}
	id, err := insert(testTempFile, nil)
	if err != nil {
		t.Fatalf("creating test temp file: %s", err)
	}
	testTempFile.ID = id
	return testTempFile
}

func createTestCommit(t *testing.T) {
	createTestFile(t)

	testCommit := &Commit{
		FileID:    testFileID,
		Hash:      testHash,
		Message:   testMessage,
		Revision:  []byte(testRevision),
		Timestamp: time.Now(),
	}

	_, err := insert(testCommit, nil)
	if err != nil {
		t.Fatalf("creating test commit: %s", err)
	}
}

func failIf(t *testing.T, err error, context ...string) {
	if err != nil {
		t.Log("failed test setup")
		t.Fatal(context, err)
	}
}

func removeTestFiles(t *testing.T) {
	_, err := connection.Exec("DELETE FROM files")
	if err != nil {
		t.Fatalf("cleaning up files: %s", err)
	}
}

func assertErrNoRows(t *testing.T, err error) {
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows, got error %s", err)
	}
}

func getTestStorage(t *testing.T) *Storage {
	createTestFile(t)
	s, err := NewStorage(testUserID, testAlias)
	failIf(t, err, "getting test storage")
	return s
}

func getTestTransaction(t *testing.T) *sql.Tx {
	tx, err := connection.Begin()
	failIf(t, err)
	return tx
}

func initTestFile(t *testing.T) {
	createTestTempFile(t, testContent)

	s, err := NewStorage(testUserID, testAlias)
	failIf(t, err, "initializing test file")
	file.Init(s, testAlias)
	failIf(t, s.Close(), "closing storage in init test file")
}
