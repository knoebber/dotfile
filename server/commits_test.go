package server

import (
	"net/http"
	"testing"
)

func TestCommitsHandler(t *testing.T) {
	router := setupTest(t, commitsHandler())
	t.Run("404", func(t *testing.T) {
		assertNotFound(t, router, testFilePath, http.MethodGet)
	})

	t.Run("ok", func(t *testing.T) {
		u := createTestUser(t)
		createTestFile(t, u.ID)
		assertOK(t, router, testFilePath, http.MethodGet)
	})
}
