package server

import (
	"net/http"
	"testing"
)

func TestIndexHandler(t *testing.T) {
	router := setupTest(t, indexHandler())
	t.Run("ok", func(t *testing.T) {
		assertOK(t, router, testFilePath, http.MethodGet)
	})
}
