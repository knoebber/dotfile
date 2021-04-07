package server

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLogin(t *testing.T) {
	setupTestDB(t)

	t.Run("invalid password", func(t *testing.T) {
		defer clearTestUser(t)
		w, r, p := setupTestPage(t)
		login(w, r, p, false)
		assert.NotEmpty(t, p.ErrorMessage)
	})

	t.Run("ok", func(t *testing.T) {
		w, r, p := setupTestPage(t)
		r.Form.Set("username", testUsername)
		r.Form.Set("password", testPassword)

		login(w, r, p, false)
		assert.Empty(t, p.ErrorMessage)
	})
}
