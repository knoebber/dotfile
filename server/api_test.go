package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/dotfile/usererror"
	"github.com/stretchr/testify/assert"
)

func TestHandleFileJSON(t *testing.T) {
	router := setupTestRouter(t, handleFileJSON)

	t.Run("404", func(t *testing.T) {
		assertNotFound(t, router, testFilePath, http.MethodGet)
	})

	t.Run("ok", func(t *testing.T) {
		createTestFile(t, createTestUser(t))
		assertOK(t, router, testFilePath, http.MethodGet)
	})
}

func TestHandleFileListJSON(t *testing.T) {
	router := setupTestRouter(t, handleFileListJSON)

	t.Run("500 on db error", func(t *testing.T) {
		db.Connection.Exec("DROP TABLE users")
		assertInternalError(t, router, testFilePath, http.MethodGet)
		setupTestDB(t)
	})

	t.Run("200 with path", func(t *testing.T) {
		createTestFile(t, createTestUser(t))
		assertOK(t, router, testFilePath+"?path=true", http.MethodGet)
	})
}

func TestHandleRawCompressedCommit(t *testing.T) {
	router := setupTestRouter(t, handleFileJSON)

	t.Run("404", func(t *testing.T) {
		assertNotFound(t, router, testFilePath, http.MethodGet)
	})

	t.Run("ok", func(t *testing.T) {
		u := createTestUser(t)
		f := createTestFile(t, u)
		path := fmt.Sprintf("/%s/%s/%s", u.Username, f.Alias, f.Hash)
		assertOK(t, router, path, http.MethodGet)
	})
}

func TestValidateAPIUser(t *testing.T) {
	var (
		w httptest.ResponseRecorder
		r http.Request
	)

	userID := validateAPIUser(&w, &r)
	assert.Empty(t, userID)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIError(t *testing.T) {
	t.Run("user error triggers 400", func(t *testing.T) {
		var w httptest.ResponseRecorder
		usererr := usererror.Invalid("test invalid")
		apiError(&w, usererr)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
