package server

import (
	"fmt"
	"net/http"

	"github.com/knoebber/dotfile/db"
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

	p.Title = fmt.Sprintf("%s@%s", alias, hash)

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
	})
}
