package server

import (
	"fmt"
	"net/http"

	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/dotfile/file"
)

// Creates a temp file; on success redirects to the confirm file page.
func handleCreateFile(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	var err error

	alias := r.Form.Get("name")
	path := r.Form.Get("path")
	content := r.Form.Get("contents")

	if alias == "" {
		alias, err = file.PathToAlias(path)
		if err != nil {
			return p.setError(w, err)
		}
	}

	tempFile := &db.TempFile{
		UserID:  p.Session.UserID,
		Alias:   alias,
		Path:    path,
		Content: []byte(content),
	}

	if err = tempFile.Create(); err != nil {
		return p.setError(w, err)
	}

	http.Redirect(w, r, "/new_file/"+tempFile.Alias, http.StatusSeeOther)
	return true
}

// Loads the results of handleCreateFile.
func loadTempFile(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	tempFile, err := db.GetTempFile(p.Session.UserID, p.Vars["alias"])
	if err != nil && !db.NotFound(err) {
		return p.setError(w, err)
	}

	if tempFile != nil {
		p.Data["alias"] = tempFile.Alias
		p.Data["path"] = tempFile.Path
		p.Data["content"] = string(tempFile.Content)
		p.Title = tempFile.Alias
	} else {
		p.Title = "New File"
	}

	return
}

func loadFile(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	file, err := db.GetFileByUsername(p.Vars["username"], p.Vars["alias"])
	if err != nil {
		return p.setError(w, err)
	}

	p.Data["path"] = file.Path
	p.Data["content"] = string(file.Content)

	p.Title = file.Alias

	return
}

func handleConfirmFile(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	alias := p.Vars["alias"]
	storage, err := db.NewStorage(p.Session.UserID, alias)
	if err != nil {
		return p.setError(w, err)
	}

	if err := file.Init(storage, alias); err != nil {
		return p.setError(w, err)
	}

	if err := storage.Close(); err != nil {
		return p.setError(w, err)
	}

	path := fmt.Sprintf("/%s/%s", p.Session.Username, alias)

	http.Redirect(w, r, path, http.StatusSeeOther)
	return true
}

func createFileHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "new_file.tmpl",
		title:        "New File",
		handleForm:   handleCreateFile,
		loadData:     loadTempFile,
		protected:    true,
	})
}

func confirmFileHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "confirm_file.tmpl",
		handleForm:   handleConfirmFile,
		loadData:     loadTempFile,
		protected:    true,
	})
}

func fileHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "file.tmpl",
		loadData:     loadFile,
	})
}
