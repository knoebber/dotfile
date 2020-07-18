package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/knoebber/dotfile/db"
)

func loadCommits(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	alias := p.Vars["alias"]
	commits, err := db.GetCommitList(p.Vars["username"], alias)
	if err != nil {
		return p.setError(w, err)
	}

	p.Data["commits"] = commits
	p.Title = fmt.Sprintf("%s commits", alias)
	return
}

func loadCommit(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	if !strings.Contains(r.Header.Get("Accept"), "text/html") {
		handleRawUncompressedCommit(w, r)
		return true
	}

	alias := p.Vars["alias"]
	hash := p.Vars["hash"]
	username := p.Vars["username"]

	commit, err := db.GetUncompressedCommit(username, alias, hash)
	if err != nil {
		return p.setError(w, err)
	}

	p.Data["hash"] = hash
	p.Data["message"] = commit.Message
	p.Data["dateString"] = commit.DateString
	p.Data["content"] = string(commit.Content)
	p.Data["path"] = commit.Path
	p.Data["current"] = commit.Current
	p.Data["forkedFrom"] = commit.ForkedFrom

	p.Title = fmt.Sprintf("%s@%s", alias, hash)

	return
}

// Handles submitting the restore form on the commits page.
// This route isn't protected because any user can view any commit.
// Adds permission checks so only an owner can restore their file.
func restoreFile(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	username := p.Vars["username"]
	alias := p.Vars["alias"]
	hash := p.Vars["hash"]

	if p.Session == nil {
		permissionDenied(w, "", username)
		return true
	}
	if !p.Owned() {
		permissionDenied(w, p.Session.Username, username)
		return true
	}

	if err := db.SetFileToHash(username, alias, hash); err != nil {
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
