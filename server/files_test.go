package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/knoebber/dotfile/db"
	"github.com/stretchr/testify/assert"
)

func TestFileHandler(t *testing.T) {
	router := setupTest(t, fileHandler())
	t.Run("404", func(t *testing.T) {
		assertNotFound(t, router, testFilePath, http.MethodGet)
	})

	u := createTestUser(t)
	createTestFile(t, u.ID)

	t.Run("ok", func(t *testing.T) {
		assertOK(t, router, testFilePath, http.MethodGet)
	})
}

func TestForkFile(t *testing.T) {
	// Example of testing a pageBuilder function.
	// Should generalize this eventually.

	p := new(Page)
	if err := db.Start(""); err != nil {
		t.Fatalf("creating test db: %s", err)
	}

	r := httptest.NewRecorder()
	request, _ := http.NewRequest(http.MethodPost, "", nil)
	forkFile(r, request, p)
	assert.NotEmpty(t, p.ErrorMessage)
}
