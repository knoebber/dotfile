package server

import (
	"net/http"

	"github.com/knoebber/dotfile/db"
)

func loadUserFiles(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	username := p.Vars["username"]
	// Not doing anything with the user yet
	// Error is used to throw 404 when the user doesn't exist.
	_, err := db.User(db.Connection, username)
	if err != nil {
		return p.setError(w, err)
	}

	p.Title = username

	files, err := db.FilesByUsername(db.Connection, username)
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
