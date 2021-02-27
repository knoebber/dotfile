package server

import (
	"net/http"
	"testing"
)

func TestFileFeed(t *testing.T) {
	router := setupTestRouter(t, createRSSFeed(Config{}))

	t.Run("ok", func(t *testing.T) {
		createTestFile(t, createTestUser(t))
		assertOK(t, router, testFilePath, http.MethodGet)
	})
}
