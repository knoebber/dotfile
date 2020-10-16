package server

import (
	"net/http"

	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/dotfile/dotfile"
)

// Loads a diff: ?on VS ?against.
func loadDiff(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	// TODO add line numbers to diff template
	alias := p.Vars["alias"]
	username := p.Vars["username"]

	on := r.URL.Query().Get("on")
	against := r.URL.Query().Get("against")

	commits, err := db.CommitList(db.Connection, username, alias)
	if err != nil {
		return p.setError(w, err)
	}

	for i, c := range commits {
		if c.Message == "" {
			commits[i].Message = c.DateString
		}
		if on != "" {
			continue
		}

		// Set 'on' to the commit before against.
		if c.Hash == against && i != len(commits)-1 {
			on = commits[i+1].Hash
		}
	}

	p.Data["commits"] = commits
	p.Data["alias"] = alias
	p.Data["against"] = against
	p.Data["on"] = on

	if on == "" || against == "" {
		return
	}

	content := &db.FileContent{Connection: db.Connection, Username: username, Alias: alias}
	diff, err := dotfile.DiffPrettyHTML(content, on, against)
	if err != nil {
		return p.setError(w, err)
	}

	p.Data["diff"] = diff
	return
}

func diffHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "diff.tmpl",
		title:        "diff",
		loadData:     loadDiff,
	})
}
