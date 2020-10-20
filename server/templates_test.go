package server

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"testing"

	"github.com/knoebber/dotfile/db"
)

// Tests the contents of all the page templates.
func TestTemplatesHTML(t *testing.T) {
	var (
		mockRow struct { // Add columns that templates expect from table.Rows
			Username  string
			Path      string
			Alias     string
			UpdatedAt string
		}
	)

	if err := loadTemplates(); err != nil {
		t.Fatalf("loading templates: %v", err)
	}

	testData := make(map[string]interface{})
	vars := make(map[string]string)
	vars["username"] = "testusername"
	vars["alias"] = "testalias"
	testSession := &db.SessionRecord{}
	controls := new(db.PageControls)
	_ = controls.Set()

	p := Page{
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
