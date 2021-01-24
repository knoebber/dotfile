package server

import (
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/knoebber/dotfile/db"
)

const (
	testUsername = "testusername"
	testPassword = "test_password"
	testAlias    = "test_alias"
	testEmail    = "test@example.com"
	testTZ       = "America/Los_Angeles"
	testFilePath = "/" + testUsername + "/" + testAlias + "/no_hash"
)

func setupTestDB(t *testing.T) {
	if err := db.Start(""); err != nil {
		t.Fatalf("creating test db: %s", err)
	}
}

func setupTestRouter(t *testing.T, handler http.HandlerFunc) *mux.Router {
	setupTestDB(t)

	if err := loadTemplates(); err != nil {
		t.Fatalf("loading templates: %v", err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/{username}/{alias}/{hash}", handler)
	return r
}

func setupTestPage(t *testing.T) (http.ResponseWriter, *http.Request, *Page) {
	p := new(Page)
	u := createTestUser(t)
	p.Session = &db.UserSession{
		UserID:   u.ID,
		Username: u.Username,
	}
	p.Vars = make(map[string]string)
	p.Data = make(map[string]interface{})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("", "/", nil)
	r.Form = url.Values{}

	return w, r, p
}

func clearTestUser(t *testing.T) {
	if err := db.DeleteUser(testUsername, testPassword); err != nil {
		t.Fatalf("clearing test user: %s", err)
	}
}

func sendTestRequest(router *mux.Router, route string, method string) *httptest.ResponseRecorder {
	r := httptest.NewRecorder()
	request, _ := http.NewRequest(method, route, nil)
	request.Header.Set("Accept", "text/html")
	router.ServeHTTP(r, request)
	return r
}

func assertNotFound(t *testing.T, router *mux.Router, route string, method string) {
	resp := sendTestRequest(router, route, method)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func assertOK(t *testing.T, router *mux.Router, route string, method string) {
	resp := sendTestRequest(router, route, method)
	body := resp.Body.String()
	assert.NotContains(t, body, "flash-error")
	assert.Equal(t, http.StatusOK, resp.Code)
}

func createTestUser(t *testing.T) *db.UserRecord {
	u, err := db.CreateUser(db.Connection, testUsername, testPassword)
	if err != nil {
		t.Fatalf("creating test user: %s", err)
	}
	return u
}

func createTestTempFile(t *testing.T, userID int64, content string) {
	tempFile := &db.TempFileRecord{
		UserID:  userID,
		Alias:   testAlias,
		Path:    "~/.test_file",
		Content: []byte(content),
	}

	if err := tempFile.Create(db.Connection); err != nil {
		t.Fatalf("creating temp file: %s", err)
	}

}

func createTestFile(t *testing.T, userID int64) {
	createTestTempFile(t, userID, "content!")

	if err := db.InitOrCommit(userID, testAlias, ""); err != nil {
		t.Fatalf("creating test file: %s", err)
	}
}
