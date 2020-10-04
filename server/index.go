package server

import (
	"github.com/knoebber/dotfile/db"
	"net/http"
)

// Loads the contents of a file by its alias.
func searchFiles(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	var err error

	controls := &db.PageControls{Values: r.URL.Query()}
	if err := controls.Set(); err != nil {
		return p.setError(w, err)
	}
	// TODO timestamps aren't in users timezone. Session should have a timezone field.
	p.Table, err = db.SearchFiles(db.Connection, controls)
	if err != nil {
		p.setError(w, err)
		return
	}

	return
}

func indexHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "index.tmpl",
		title:        "Dotfilehub",
		loadData:     searchFiles,
	})
}
