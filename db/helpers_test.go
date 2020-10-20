package db

import (
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/knoebber/dotfile/dotfile"
	"github.com/knoebber/dotfile/usererror"
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
	testHash           = "9abdbcf4ea4e2c1c077c21b8c2f2470ff36c31ce"
	testMessage        = "commit message"
	testUsername       = "genericusername"
	testPassword       = "ilovecatS!"
	testEmail          = "dot@dotfilehub.com"
	testCliToken       = "12345678"
)

func assertUsererror(t *testing.T, err error) {
	var usererr *usererror.Error
	if !errors.As(err, &usererr) {
		t.Errorf("expected error to be usererror, received %s", err)
	}
}

func createTestDB(t *testing.T) {
	if Connection != nil {
		resetTestDB(t)
	}

	var err error

	// Set this to true to save a test database to testdir for debugging.
	const persistDB = false
	if persistDB {
		_ = os.RemoveAll(testDir)
		_ = os.Mkdir(testDir, 0755)
		err = Start(testDir + "dotfilehub.db")
	} else {
		err = Start("")
	}
	if err != nil {
		t.Fatalf("creating test db: %s", err)
	}
}

func testUserList(t *testing.T) (usernames []string) {
	var username string
	rows, err := Connection.
		Query("SELECT username FROM users")
	failIf(t, err, "listing test user")
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&username)
		failIf(t, err, "scanning test username")
		usernames = append(usernames, username)
	}
	return

}

func countTestUser(t *testing.T, username string) (count int) {
	err := Connection.
		QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).
		Scan(&count)
	failIf(t, err, "counting test user", username)
	return
}

func resetTestDB(t *testing.T) {
	usernames := testUserList(t)

	for _, u := range usernames {
		if err := DeleteUser(u, testPassword); err != nil {
			t.Fatalf("unable to delete test users: %s", err)
		}
	}
}

func createTestUser(t *testing.T, userID int64, username, email string) {
	if countTestUser(t, username) > 0 {
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("creating test password: %s", err)
	}

	_, err = Connection.Exec(`
INSERT INTO users(id, username, email, password_hash, cli_token) 
VALUES(?, ?, ?, ?, ?)`,
		userID,
		username,
		email,
		hashed,
		testCliToken,
	)
	if err != nil {
		t.Fatalf("creating test user %q: %s", username, err)
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

	id, err := insert(Connection, testTempFile)
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

func testTransaction(t *testing.T) *sql.Tx {
	tx, err := Connection.Begin()
	failIf(t, err, "starting transaction for test")
	return tx
}

func initTestFile(t *testing.T) *FileView {
	createTestUser(t, testUserID, testUsername, testEmail)
	createTestTempFile(t, testContent)

	tx := testTransaction(t)
	ft, err := StageFile(tx, testUsername, testAlias)
	failIf(t, err, "new storage in init test file")
	failIf(t, dotfile.Init(ft, testPath, testAlias), "initialing test file")
	failIf(t, tx.Commit())

	f, err := UncompressFile(Connection, testUsername, testAlias)
	failIf(t, err, "getting file by username in init test file")
	return f
}

// Creates a test file, an initial commit, and an additional commit.
func initTestFileAndCommit(t *testing.T) (initialCommit CommitSummary, currentCommit CommitSummary) {
	initTestFile(t)

	// Latest commit will have this content.
	createTestTempFile(t, testUpdatedContent)

	tx := testTransaction(t)
	ft, err := StageFile(tx, testUsername, testAlias)
	failIf(t, err, "staging test file")

	// Ensure that the new commit has a different timestamp - unix time is by the second.
	time.Sleep(time.Second)

	failIf(t, dotfile.NewCommit(ft, "Commiting test updated content"))
	failIf(t, tx.Commit())

	lst, err := CommitList(Connection, testUsername, testAlias)
	failIf(t, err, "getting test commit")

	if len(lst) != 2 {
		t.Fatalf("expected commit list to be length 2, got %d", len(lst))
	}

	f, err := UncompressFile(Connection, testUsername, testAlias)
	failIf(t, err, "initTestFileAndCommit: GetFileByUsername")

	currentCommit = lst[0]
	initialCommit = lst[1]

	assert.Equal(t, currentCommit.Hash, f.Hash)
	return
}
