package server

import (
	"fmt"
	"net/http"

	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/dotfile/file"
)

// Handles submitting the new file form.
func newTempFile(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	content := r.Form.Get("contents")
	path := r.Form.Get("path")

	alias, err := file.GetAlias(r.Form.Get("name"), path)
	if err != nil {
		return p.setError(w, err)
	}

	tempFile := &db.TempFile{
		UserID:  p.Session.UserID,
		Alias:   alias,
		Path:    path,
		Content: []byte(content),
	}

	if err := tempFile.Create(); err != nil {
		return p.setError(w, err)
	}

	http.Redirect(w, r, "/temp_file/"+alias+"/create", http.StatusSeeOther)
	return true
}

// Handles submitting the edit file form.
func editFile(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	content := r.Form.Get("contents")

	existingFile, err := db.GetFileByUsername(p.Session.Username, p.Vars["alias"])
	if err != nil {
		return p.setError(w, err)
	}

	alias := existingFile.Alias
	path := existingFile.Path

	tempFile := &db.TempFile{
		UserID:  p.Session.UserID,
		Alias:   alias,
		Path:    path,
		Content: []byte(content),
	}

	if err := tempFile.Create(); err != nil {
		return p.setError(w, err)
	}

	http.Redirect(w, r, "/temp_file/"+alias+"/commit", http.StatusSeeOther)

	return true
}

// Handles submitting the confirm file form.
// Either initializes a new file or makes a commit to an existing.
func confirmTempFile(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	var err error

	alias := p.Vars["alias"]
	storage, err := db.NewStorage(p.Session.UserID, alias)
	if err != nil {
		return p.setError(w, err)
	}

	if storage.Staged.New {
		err = file.Init(storage, alias)
	} else {
		err = file.NewCommit(storage, r.Form.Get("message"))
	}
	if err != nil {
		return p.setError(w, err)
	}

	if err := storage.Close(); err != nil {
		return p.setError(w, err)
	}

	path := fmt.Sprintf("/%s/%s", p.Session.Username, alias)

	http.Redirect(w, r, path, http.StatusSeeOther)
	return true
}

// Loads the contents of a file by its alias.
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

// Loads the contents of a users temp file for the create/edit form.
func loadTempFileForm(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	newFile := p.Vars["alias"] == ""
	p.Data["newFile"] = newFile

	if newFile {
		p.Title = "New File"
	} else {
		p.Title = "Edit File"
	}

	// Only load the previous temp file if there is an ?edit query param.
	editing := r.URL.Query().Get("edit") == "true"

	if newFile && !editing {
		return
	} else if !newFile && !editing {
		return loadFile(w, r, p)
	}

	tempFile, err := db.GetTempFile(p.Session.UserID, p.Vars["alias"])
	if err != nil && !db.NotFound(err) {
		return p.setError(w, err)
	}
	if tempFile == nil {
		return
	}

	p.Data["alias"] = tempFile.Alias
	p.Data["path"] = tempFile.Path
	p.Data["content"] = string(tempFile.Content)
	return
}

// Loads the contents of a users temp file for the confirm edit page.
func loadCommitConfirm(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	alias := p.Vars["alias"]
	storage, err := db.NewStorage(p.Session.UserID, alias)
	if err != nil {
		return p.setError(w, err)
	}

	// TODO show user when there are no changes. Currently shows the first and last
	// two lines of the unchanged file which is user friendly.
	diffs, err := file.Diff(storage, storage.Staged.CurrentRevision, "")
	if err != nil {
		return p.setError(w, err)
	}

	p.Data["diffs"] = diffs
	p.Data["alias"] = storage.Staged.Alias
	p.Data["path"] = storage.Staged.Path
	p.Data["editAction"] = fmt.Sprintf("/%s/%s/edit", p.Session.Username, alias)
	p.Title = "Edit " + storage.Staged.Alias
	return
}

// Loads the contents of a users temp file for the new file page.
func loadNewFileConfirm(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	tempFile, err := db.GetTempFile(p.Session.UserID, p.Vars["alias"])
	if err != nil {
		return p.setError(w, err)
	}

	p.Data["alias"] = tempFile.Alias
	p.Data["path"] = tempFile.Path
	p.Data["content"] = string(tempFile.Content)
	p.Data["editAction"] = "/new_file"
	p.Title = "New File"
	return
}

func createFileHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "file_form.tmpl",
		title:        "New File",
		handleForm:   newTempFile,
		loadData:     loadTempFileForm,
		protected:    true,
	})
}

func editFileHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "file_form.tmpl",
		handleForm:   editFile,
		loadData:     loadTempFileForm,
		protected:    true,
	})
}

func confirmNewFileHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "confirm_file.tmpl",
		handleForm:   confirmTempFile,
		loadData:     loadNewFileConfirm,
		protected:    true,
	})
}

func confirmEditHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "confirm_edit.tmpl",
		handleForm:   confirmTempFile,
		loadData:     loadCommitConfirm,
		protected:    true,
	})
}

func fileHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "file.tmpl",
		loadData:     loadFile,
	})
}