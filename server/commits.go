package server

import (
	"fmt"
	"net/http"

	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/dotfile/file"
)

func loadCommits(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	alias := p.Vars["alias"]
	commits, err := db.GetCommitList(p.Vars["username"], alias)
	if err != nil {
		return p.setError(w, err)
	}

	p.Data["commits"] = commits
	p.Title = "Commits"
	p.Title = fmt.Sprintf("%s commits", alias)
	return
}

func loadCommit(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	alias := p.Vars["alias"]
	hash := p.Vars["hash"]

	commit, err := db.GetCommit(p.Vars["username"], alias, hash)
	if err != nil {
		return p.setError(w, err)
	}

	p.Data["message"] = commit.Message
	p.Data["timestamp"] = commit.Timestamp
	p.Data["content"] = commit.Content
	p.Data["path"] = commit.Path
	p.Data["current"] = commit.Current

	p.Title = fmt.Sprintf("%s@%s", alias, hash)

	return
}

// Handles submitting the restore form on the commits page.
// This route isn't protected because any user can view any commit.
// Adds permission checks so only an owner can restore their file.
func restoreFile(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	if p.Session == nil {
		permissionDenied(w, "", p.Vars["username"])
		return true
	}
	if !p.Owned() {
		permissionDenied(w, p.Session.Username, p.Vars["username"])
		return true
	}

	alias := p.Vars["alias"]
	hash := p.Vars["hash"]
	storage, err := db.NewStorage(p.Session.UserID, alias)
	if err != nil {
		return p.setError(w, err)
	}

	if err := file.Checkout(storage, hash); err != nil {
		return p.setError(w, err)
	}

	if err := storage.Close(); err != nil {
		return p.setError(w, err)
	}

	return

}

func commitsHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "commits.tmpl",
		loadData:     loadCommits,
	})
}

func commitHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "commit.tmpl",
		loadData:     loadCommit,
		handleForm:   restoreFile,
	})
}
