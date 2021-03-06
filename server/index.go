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

	p.Table, err = db.SearchFiles(db.Connection, controls, p.Timezone())
	if err != nil {
		return p.setError(w, err)
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
