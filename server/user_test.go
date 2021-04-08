package server

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestUserHandler(t *testing.T) {
	router := setupTestRouter(t, userHandler())
	t.Run("404", func(t *testing.T) {
		assertNotFound(t, router, testFilePath, http.MethodGet)
	})

	t.Run("ok", func(t *testing.T) {
		_ = createTestUser(t)
		assertOK(t, router, testFilePath, http.MethodGet)
	})
}

func TestReloadSession(t *testing.T) {
	setupTestDB(t)

	t.Run("ok on GET", func(t *testing.T) {
		defer clearTestUser(t)
		w, r, p := setupTestPage(t)
		r.Method = "GET"

		reloadSession(w, r, p)
		assert.Empty(t, p.ErrorMessage)
	})

	t.Run("ok on POST", func(t *testing.T) {
		w, r, p := setupTestPage(t)
		r.Method = "POST"

		reloadSession(w, r, p)
		assert.Empty(t, p.ErrorMessage)
	})
}

func TestLoadThemes(t *testing.T) {
	setupTestDB(t)

	w, r, p := setupTestPage(t)
	loadThemes(w, r, p)
	assert.Empty(t, p.ErrorMessage)
}

func TestHandleTimezone(t *testing.T) {
	setupTestDB(t)

	t.Run("ok", func(t *testing.T) {
		w, r, p := setupTestPage(t)
		handleTimezone(w, r, p)
		assert.Empty(t, p.ErrorMessage)
	})
}

func TestHandleTheme(t *testing.T) {
	setupTestDB(t)

	t.Run("ok", func(t *testing.T) {
		w, r, p := setupTestPage(t)
		handleTheme(w, r, p)
		assert.Empty(t, p.ErrorMessage)
	})
}

func TestHandleTokenForm(t *testing.T) {
	setupTestDB(t)

	t.Run("error on token mismatch", func(t *testing.T) {
		defer clearTestUser(t)
		w, r, p := setupTestPage(t)
		handleTokenForm(w, r, p)
		assert.NotEmpty(t, p.ErrorMessage)
	})

	t.Run("ok", func(t *testing.T) {
		w, r, p := setupTestPage(t)
		r.Form.Set("token", p.CLIToken())
		handleTokenForm(w, r, p)
		assert.Empty(t, p.ErrorMessage)
	})

}
