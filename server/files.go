package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/dotfile/dotfile"
	"github.com/knoebber/usererror"
)

// Handles submitting the new file form.
func newTempFile(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	content := r.Form.Get("contents")
	path := r.Form.Get("path")

	alias, err := dotfile.Alias(r.Form.Get("alias"), path)
	if err != nil {
		return p.setError(w, err)
	}
	if err := db.ValidateFileNotExists(db.Connection, p.userID(), alias, path); err != nil {
		return p.setError(w, err)
	}

	tempFile := &db.TempFileRecord{
		UserID:  p.userID(),
		Alias:   alias,
		Path:    path,
		Content: []byte(content),
	}

	if err := tempFile.Create(db.Connection); err != nil {
		return p.setError(w, err)
	}

	http.Redirect(w, r, fmt.Sprintf("/%s/%s/init", p.Username(), alias), http.StatusSeeOther)
	return true
}

// Handles submitting the edit file form.
func editFile(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	content := r.Form.Get("contents")

	existingFile, err := db.File(db.Connection, p.Username(), p.Vars["alias"])
	if err != nil {
		return p.setError(w, err)
	}

	alias := existingFile.Alias
	path := existingFile.Path

	tempFile := &db.TempFileRecord{
		UserID:  p.userID(),
		Alias:   alias,
		Path:    path,
		Content: []byte(content),
	}

	if err := tempFile.Create(db.Connection); err != nil {
		return p.setError(w, err)
	}

	http.Redirect(w, r, fmt.Sprintf("/%s/%s/commit", p.Username(), alias), http.StatusSeeOther)
	return true
}

// Handles submitting the confirm file form.
// Either initializes a new file or makes a commit to an existing.
func confirmTempFile(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	if err := db.InitOrCommit(p.userID(), p.Vars["alias"], r.Form.Get("message")); err != nil {
		return p.setError(w, err)
	}

	path := fmt.Sprintf("/%s/%s", p.Username(), p.Vars["alias"])
	http.Redirect(w, r, path, http.StatusSeeOther)
	return true
}

// Forks another user's file at hash.
func forkFile(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	username := r.Form.Get("username")
	alias := r.Form.Get("alias")
	hash := r.Form.Get("hash")
	loggedInUserID := p.userID()

	if loggedInUserID < 1 {
		return p.setError(w, usererror.New("Must be logged in to fork file."))
	}

	if err := db.ForkFile(username, alias, hash, loggedInUserID); err != nil {
		return p.setError(w, err)
	}

	return
}

// Uncompress file and sets content to page data.
func uncompressFile(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	if !strings.Contains(r.Header.Get("Accept"), "text/html") {
		handleRawFile(w, r)
		return true
	}

	username := p.Vars["username"]
	alias := p.Vars["alias"]

	file, err := db.UncompressFile(db.Connection, username, alias)
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
		p.Title = "new file"
	} else {
		p.Title = "edit"
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
		tempFile, err := db.TempFile(db.Connection, p.userID(), pageAlias)
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
		file, err := db.UncompressFile(db.Connection, p.Vars["username"], pageAlias)
		if err != nil {
			return p.setError(w, err)
		}
		p.Data["path"] = file.Path
		p.Data["content"] = string(file.Content)
		return
	} else if !newFile && !editing && at != "" {
		commit, err := db.UncompressCommit(db.Connection, p.Vars["username"], pageAlias, at, p.Timezone())
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

	f, err := db.UncompressFile(db.Connection, p.Username(), alias)
	if err != nil {
		return p.setError(w, err)
	}

	p.Data["editAction"] = fmt.Sprintf("/%s/%s/edit", p.Username(), alias)
	p.Data["alias"] = f.Alias
	p.Data["path"] = f.Path

	diff, err := getHtmlDiff(&db.FileContent{
		Connection: db.Connection,
		Username:   p.Username(),
		UserID:     p.userID(),
		Alias:      alias,
	}, f.Hash, "")

	if err != nil {
		return p.setError(w, err)
	}

	p.Data["diff"] = diff
	return
}

// Loads the contents of a users temp file for the new file page.
func loadNewFileConfirm(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	tempFile, err := db.TempFile(db.Connection, p.userID(), p.Vars["alias"])
	if err != nil {
		return p.setError(w, err)
	}

	p.Data["alias"] = tempFile.Alias
	p.Data["path"] = tempFile.Path
	p.Data["content"] = string(tempFile.Content)
	p.Data["editAction"] = "/new_file"
	p.Title = "confirm new file"
	return
}

func newFileHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "file_form.tmpl",
		title:        "new file",
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
		title:        "commit",
		handleForm:   confirmTempFile,
		loadData:     loadCommitConfirm,
		protected:    true,
	})
}

func fileHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "file.tmpl",
		handleForm:   forkFile,
		loadData:     uncompressFile,
	})
}
