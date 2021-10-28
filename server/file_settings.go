package server

import (
	"net/http"

	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/usererror"
	"github.com/pkg/errors"
)

func updateFile(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	currentAlias := p.Vars["alias"]
	alias := r.Form.Get("alias")
	path := r.Form.Get("path")

	record, err := db.File(db.Connection, p.Username(), currentAlias)
	if err != nil {
		return p.setError(w, err)
	}

	if err := record.Update(db.Connection, alias, path); err != nil {
		return p.setError(w, err)
	}

	http.Redirect(w, r, "/"+p.Username()+"/"+alias, http.StatusSeeOther)
	return true
}

func clearFile(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	username := p.Vars["username"]
	alias := p.Vars["alias"]
	tx, err := db.Connection.Begin()
	if err != nil {
		return p.setError(w, errors.Wrap(err, "starting transaction for clear file"))
	}

	if err := db.ClearCommits(tx, username, alias); err != nil {
		return p.setError(w, db.Rollback(tx, err))
	}
	if err := tx.Commit(); err != nil {
		return p.setError(w, errors.Wrap(err, "committing transaction for clear file"))
	}

	http.Redirect(w, r, "/"+p.Username()+"/"+alias+"/commits", http.StatusSeeOther)
	return true
}

func deleteFile(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	deleteConfirm := r.Form.Get("delete")
	username := p.Vars["username"]
	alias := p.Vars["alias"]
	if alias != deleteConfirm {
		return p.setError(w, usererror.New("Alias does not match"))
	}

	tx, err := db.Connection.Begin()
	if err != nil {
		return p.setError(w, errors.Wrap(err, "starting transaction for delete file"))
	}

	if err := db.DeleteFile(tx, username, alias); err != nil {
		return p.setError(w, db.Rollback(tx, err))
	}
	if err := tx.Commit(); err != nil {
		return p.setError(w, errors.Wrap(err, "committing transaction for delete file"))
	}

	http.Redirect(w, r, "/"+p.Username(), http.StatusSeeOther)
	return true

}

func loadFile(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	username := p.Vars["username"]
	alias := p.Vars["alias"]

	file, err := db.File(db.Connection, username, alias)
	if err != nil {
		return p.setError(w, err)
	}

	p.Data["path"] = file.Path
	return
}

func fileSettingsHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "file_settings.tmpl",
		title:        "settings",
		protected:    true,
	})
}

func updateFileHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "update_file.tmpl",
		title:        "update",
		loadData:     loadFile,
		handleForm:   updateFile,
		protected:    true,
	})
}

func clearFileHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "delete_commits.tmpl",
		title:        "clear",
		handleForm:   clearFile,
		protected:    true,
	})
}
func deleteFileHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "delete_file.tmpl",
		title:        "delete",
		handleForm:   deleteFile,
		protected:    true,
	})
}
