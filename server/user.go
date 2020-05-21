package server

import (
	"net/http"

	"github.com/knoebber/dotfile/db"
)

func loadUserFiles(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	username := p.Vars["username"]
	files, err := db.GetFilesByUsername(username)
	if err != nil {
		return p.setError(w, err)
	}

	p.Data["files"] = files
	p.Title = username
	return
}

func userHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "user.tmpl",
		loadData:     loadUserFiles,
	})
}
