package server

import (
	"bytes"
	"encoding/xml"
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/knoebber/dotfile/db"
)

func TestSessionGetters(t *testing.T) {
	p := &Page{}
	assert.Empty(t, p.Timezone())
	assert.Empty(t, p.Username())
	assert.Empty(t, p.Email())
	assert.Empty(t, p.Theme())
	assert.Empty(t, p.CLIToken())
	assert.Empty(t, p.UserCreatedAt())
	assert.Empty(t, p.session())
	assert.Empty(t, p.Timezone())
	p = testPage(t)
	assert.NotEmpty(t, p.Timezone())
	assert.NotEmpty(t, p.Username())
	assert.NotEmpty(t, p.Email())
	assert.NotEmpty(t, p.Theme())
	assert.NotEmpty(t, p.CLIToken())
	assert.NotEmpty(t, p.UserCreatedAt())
	assert.NotEmpty(t, p.session())
	assert.NotEmpty(t, p.userID())
	assert.NotEmpty(t, p.Timezone())
}

// Tests the contents of all the page templates.
func TestTemplatesHTML(t *testing.T) {

	if err := loadTemplates(); err != nil {
		t.Fatalf("loading templates: %v", err)
	}

	p := testPage(t)
	for _, template := range pageTemplates.Templates() {
		curr := template.Name()
		if curr[0] == '_' || curr == "pages" {
			// Skip partials and the name of the root template.
			continue
		}

		buff := new(bytes.Buffer)

		p.templateName = curr
		if err := p.writeFromTemplate(buff); err != nil {
			t.Fatalf("failed to write from template: %v", err)
		}

		assertHTMLResponse(t, string(buff.Bytes()), curr)
	}
}

// Asserts that the html body is valid HTML and does not have any empty lines.
func assertHTMLResponse(t *testing.T, body, name string) {
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		if strings.Trim(line, " ") == "" {
			t.Log(body)
			t.Fatalf("%q body: line %d is empty", name, i)
		}
	}

	htmlReader := strings.NewReader(body)
	d := xml.NewDecoder(htmlReader)
	d.Strict = false
	d.AutoClose = xml.HTMLAutoClose
	d.Entity = xml.HTMLEntity

	for {
		_, err := d.Token()
		switch err {
		case io.EOF:
			return
		case nil:
		default:
			t.Log(body)
			t.Fatalf("%q: %v", name, err)
		}
	}

}
func testPage(t *testing.T) *Page {
	var (
		mockRow struct { // Add columns that templates expect from table.Rows
			Username        string
			Path            string
			Alias           string
			UpdatedAtString string
		}
	)
	testData := make(map[string]interface{})
	vars := make(map[string]string)
	vars["username"] = testUsername
	vars["alias"] = testAlias
	testEmail := testEmail
	testTZ := testTZ
	testSession := &db.UserSession{
		Session:       "sess123",
		IP:            "192.168.69",
		UserID:        1,
		Username:      testUsername,
		Email:         &testEmail,
		CLIToken:      "1234token",
		Timezone:      &testTZ,
		Theme:         db.UserThemeDark,
		UserCreatedAt: time.Now().Format(time.RFC3339),
	}
	controls := new(db.PageControls)
	if err := controls.Set(); err != nil {
		t.Fatalf("setting page controls: %s", err)
	}

	return &Page{
		Title:   "Test Page",
		Data:    testData,
		Vars:    vars,
		Session: testSession,
		Table: &db.HTMLTable{
			Columns:  []string{"Test", "Test2", "Test3"},
			Rows:     []interface{}{mockRow},
			Controls: controls,
		},
	}

}
