package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/knoebber/dotfile/db"
	"github.com/stretchr/testify/assert"
)

func TestFileHandler(t *testing.T) {
	router := setupTestRouter(t, fileHandler())
	t.Run("404", func(t *testing.T) {
		assertNotFound(t, router, testFilePath, http.MethodGet)
	})

	createTestFile(t, createTestUser(t))

	t.Run("ok", func(t *testing.T) {
		assertOK(t, router, testFilePath, http.MethodGet)
	})
}

func TestNewTempFile(t *testing.T) {
	setupTestDB(t)
	w, r, p := setupTestPage(t)

	t.Run("error when form not set", func(t *testing.T) {
		newTempFile(w, r, p)
		assert.NotEmpty(t, p.ErrorMessage)
	})

	t.Run("ok", func(t *testing.T) {
		p.ErrorMessage = ""

		r.Form.Set("alias", testAlias)
		r.Form.Set("path", "/path")
		r.Form.Set("contents", "stuff!")

		newTempFile(w, r, p)
		assert.Empty(t, p.ErrorMessage)
	})
}

func TestEditFile(t *testing.T) {
	setupTestDB(t)
	w, r, p := setupTestPage(t)
	createTestFile(t, &db.UserRecord{Username: p.Session.Username, ID: p.Session.UserID})
	p.Vars["alias"] = testAlias

	t.Run("error when content empty", func(t *testing.T) {
		editFile(w, r, p)
		assert.NotEmpty(t, p.ErrorMessage)
	})

	t.Run("ok", func(t *testing.T) {
		p.ErrorMessage = ""
		r.Form.Set("contents", "new commit")
		editFile(w, r, p)
		assert.Empty(t, p.ErrorMessage)
	})
}

func TestConfirmTempFile(t *testing.T) {
	setupTestDB(t)
	w, r, p := setupTestPage(t)
	createTestTempFile(t, p.Session.UserID, "confirm temp file content")

	t.Run("error when alias empty", func(t *testing.T) {
		confirmTempFile(w, r, p)
		assert.NotEmpty(t, p.ErrorMessage)
	})

	t.Run("ok", func(t *testing.T) {
		p.Vars["alias"] = testAlias
		p.ErrorMessage = ""
		confirmTempFile(w, r, p)
		assert.Empty(t, p.ErrorMessage)
	})
}

func TestLoadCommitConfirm(t *testing.T) {
	setupTestDB(t)

	w, r, p := setupTestPage(t)
	createTestFile(t, &db.UserRecord{Username: p.Session.Username, ID: p.Session.UserID})
	createTestTempFile(t, p.Session.UserID, "load commit confirm content")

	t.Run("ok", func(t *testing.T) {
		p.Vars["alias"] = testAlias

		loadCommitConfirm(w, r, p)
		assert.Empty(t, p.ErrorMessage)
	})
}

func TestForkFile(t *testing.T) {
	setupTestDB(t)
	p := new(Page)

	t.Run("error when user not logged in", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodPost, "", nil)
		forkFile(w, r, p)
		assert.NotEmpty(t, p.ErrorMessage)
	})
}
