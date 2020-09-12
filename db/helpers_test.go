package db

import (
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/knoebber/dotfile/file"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

const (
	testDir            = "testdata/"
	testAlias          = "testalias"
	testPath           = "~/dotfile/test-file.txt"
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

func createTestUser(t *testing.T, userID int64, username, email string) {
	var count int

	err := connection.
		QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", userID).
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
		userID,
		username,
		email,
		hashed,
		testCliToken,
	)
	if err != nil {
		t.Fatalf("creating test user: %s", err)
	}
}

func createTestTempFile(t *testing.T, content string) *TempFileRecord {
	createTestUser(t, testUserID, testUsername, testEmail)

	testTempFile := &TempFileRecord{
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

func getTestFileTransaction(t *testing.T) *FileTransaction {
	ft, err := NewFileTransaction(testUsername, testAlias)
	failIf(t, err, "getting test storage")
	return ft
}

func getTestTransaction(t *testing.T) *sql.Tx {
	tx, err := connection.Begin()
	failIf(t, err)
	return tx
}

func initTestFile(t *testing.T) *FileView {
	createTestUser(t, testUserID, testUsername, testEmail)
	createTestTempFile(t, testContent)

	ft, err := StageFile(testUsername, testAlias)
	failIf(t, err, "new storage in init test file")
	failIf(t, file.Init(ft, testPath, testAlias), "initialing test file")
	failIf(t, ft.Close())

	f, err := File(testUsername, testAlias)
	failIf(t, err, "getting file by username in init test file")
	return f
}

// Creates a test file, an initial commit, and an additional commit.
func initTestFileAndCommit(t *testing.T) (initialCommit CommitSummary, currentCommit CommitSummary) {
	initTestFile(t)

	// Latest commit will have this content.
	createTestTempFile(t, testUpdatedContent)

	ft, err := StageFile(testUsername, testAlias)
	failIf(t, err, "staging test file")

	// Ensure that the new commit has a different timestamp - unix time is by the second.
	time.Sleep(time.Second)

	failIf(t, file.NewCommit(ft, "Commiting test updated content"))
	failIf(t, ft.Close())

	lst, err := CommitList(testUsername, testAlias)
	failIf(t, err, "getting test commit")

	if len(lst) != 2 {
		t.Fatalf("expected commit list to be length 2, got %d", len(lst))
	}

	f, err := File(testUsername, testAlias)
	failIf(t, err, "initTestFileAndCommit: GetFileByUsername")

	currentCommit = lst[0]
	initialCommit = lst[1]

	assert.Equal(t, currentCommit.Hash, f.Hash)
	return
}

// Should run after any tests that use a transaction.
func assertDBNotLocked(t *testing.T) {
	_, err := connection.Exec("DELETE FROM sessions")
	if err != nil {
		t.Fatalf("asserting db is still open: %s", err)
	}

}
