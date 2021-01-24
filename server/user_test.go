package server

import (
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
