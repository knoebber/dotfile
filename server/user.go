package server

import (
	"net/http"

	"github.com/knoebber/dotfile/db"
)

func loadUserFiles(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	username := p.Vars["username"]
	p.Title = username

	files, err := db.GetFilesByUsername(username)
	if db.NotFound(err) {
		return
	} else if err != nil {
		return p.setError(w, err)
	}

	p.Data["files"] = files
	return
}

func userHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "user.tmpl",
		loadData:     loadUserFiles,
	})
}
