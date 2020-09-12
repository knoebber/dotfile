package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/dotfile/file"
	"github.com/knoebber/dotfile/usererror"
)

// Handles submitting the new file form.
func newTempFile(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	content := r.Form.Get("contents")
	path := r.Form.Get("path")

	alias, err := file.Alias(r.Form.Get("alias"), path)
	if err != nil {
		return p.setError(w, err)
	}

	// Expect not found error.
	_, err = db.File(p.Session.Username, alias)
	if err == nil {
		return p.setError(w, usererror.Duplicate("File", alias))
	}
	if !db.NotFound(err) {
		return p.setError(w, err)
	}

	tempFile := &db.TempFileRecord{
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

// Handles submitting the update file form.
func updateFile(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	currentAlias := p.Vars["alias"]
	alias := r.Form.Get("alias")
	path := r.Form.Get("path")

	// Delete the file when the user submits the alias into the form.
	delete := r.Form.Get("delete")

	file, err := db.File(p.Session.Username, currentAlias)
	if err != nil {
		return p.setError(w, err)
	}

	if currentAlias == delete {
		if err := file.Delete(); err != nil {
			return p.setError(w, err)
		}

		http.Redirect(w, r, "/"+p.Session.Username, http.StatusSeeOther)
		return true
	} else if delete != "" {
		return p.setError(w, usererror.Invalid("Alias does not match"))
	}

	if err := file.Update(alias, path); err != nil {
		return p.setError(w, err)
	}

	http.Redirect(w, r, fmt.Sprintf("/%s/%s", p.Session.Username, alias), http.StatusSeeOther)
	return true
}

// Handles submitting the edit file form.
func editFile(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	content := r.Form.Get("contents")

	existingFile, err := db.File(p.Session.Username, p.Vars["alias"])
	if err != nil {
		return p.setError(w, err)
	}

	alias := existingFile.Alias
	path := existingFile.Path

	tempFile := &db.TempFileRecord{
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
	tx, err := db.StageFile(p.Session.Username, alias)
	if err != nil {
		return p.setError(w, err)
	}

	if !tx.FileExists {
		err = file.Init(tx, tx.Staged.Path, alias)
	} else {
		err = file.NewCommit(tx, r.Form.Get("message"))
	}
	if err != nil {
		return p.setError(w, err)
	}
	if err := tx.Close(); err != nil {
		return p.setError(w, err)
	}

	path := fmt.Sprintf("/%s/%s", p.Session.Username, alias)

	http.Redirect(w, r, path, http.StatusSeeOther)
	return true
}

// Forks another user's file at hash.
func forkFile(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	username := r.Form.Get("username")
	alias := r.Form.Get("alias")
	hash := r.Form.Get("hash")

	err := db.ForkFile(username, alias, hash, p.Session.UserID)
	if err != nil {
		return p.setError(w, err)
	}

	return
}

// Loads the contents of a file by its alias.
func searchFiles(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	query := r.URL.Query().Get("q")
	if query == "" {
		return
	}

	result, err := db.Search(query)
	if err != nil {
		return p.setError(w, err)
	}

	p.Data["files"] = result
	p.Data["query"] = query

	return
}

// Loads the contents of a file by its alias.
func loadFile(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	if !strings.Contains(r.Header.Get("Accept"), "text/html") {
		handleRawFile(w, r)
		return true
	}

	username := p.Vars["username"]
	alias := p.Vars["alias"]

	file, err := db.File(username, alias)
	if err != nil {
		return p.setError(w, err)
	}

	p.Data["path"] = file.Path
	p.Data["content"] = string(file.Content)
	p.Data["hash"] = file.Hash

	p.Title = file.Alias

	return
}

// Loads data into the create/edit form.
// Fills the text area with content from a tempfile or a current file depending on query params.
func loadTempFileForm(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	pageAlias := p.Vars["alias"]
	newFile := pageAlias == ""
	p.Data["newFile"] = newFile

	if newFile {
		p.Title = "New File"
	} else {
		p.Title = pageAlias + " - Edit"
	}

	// Edit from a specific hash with the at param.
	at := r.URL.Query().Get("at")

	// Load the user's temp file if there is an ?edit query param.
	editing := r.URL.Query().Get("edit") == "true"

	if newFile && !editing {
		// A new file that is not being edited. No content needs to be loaded.
		return
	}
	if editing {
		tempFile, err := db.TempFile(p.Session.Username, pageAlias)
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

	if at == "" {
		// Load the current content.
		file, err := db.File(p.Vars["username"], pageAlias)
		if err != nil {
			return p.setError(w, err)
		}
		p.Data["path"] = file.Path
		p.Data["content"] = string(file.Content)
		return
	} else if !newFile && !editing && at != "" {
		commit, err := db.UncompressedCommit(p.Vars["username"], pageAlias, at)
		if err != nil {
			return p.setError(w, err)
		}

		p.Data["path"] = commit.Path
		p.Data["content"] = string(commit.Content)
	}

	return
}

// Loads the contents of a users temp file for the confirm edit page.
func loadCommitConfirm(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	alias := p.Vars["alias"]

	f, err := db.File(p.Session.Username, alias)
	if err != nil {
		return p.setError(w, err)
	}

	content := &db.FileContent{Username: p.Session.Username, Alias: alias}

	diffs, err := file.Diff(content, f.Hash, "")
	if err != nil {
		return p.setError(w, err)
	}

	p.Data["diffs"] = diffs
	p.Data["alias"] = f.Alias
	p.Data["path"] = f.Path
	p.Data["editAction"] = fmt.Sprintf("/%s/%s/edit", p.Session.Username, alias)
	p.Title = f.Alias + " - edit"
	return
}

// Loads the contents of a users temp file for the new file page.
func loadNewFileConfirm(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	tempFile, err := db.TempFile(p.Session.Username, p.Vars["alias"])
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

func indexHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "index.tmpl",
		title:        "Dotfilehub",
		loadData:     searchFiles,
	})
}

func newFileHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "file_form.tmpl",
		title:        "New File",
		handleForm:   newTempFile,
		loadData:     loadTempFileForm,
		protected:    true,
	})
}

func fileSettingsHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "file_settings.tmpl",
		handleForm:   updateFile,
		loadData:     loadFile,
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
		handleForm:   forkFile,
		loadData:     loadFile,
	})
}
