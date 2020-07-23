package server

import (
	"net/http"

	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/dotfile/file"
)

// Loads a diff: ?on VS ?against.
func loadDiff(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	// TODO add line numbers to diff template
	alias := p.Vars["alias"]
	username := p.Vars["username"]

	on := r.URL.Query().Get("on")
	against := r.URL.Query().Get("against")

	commits, err := db.GetCommitList(username, alias)
	if err != nil {
		return p.setError(w, err)
	}

	p.Data["commits"] = commits
	p.Data["alias"] = alias
	p.Data["on"] = on
	p.Data["against"] = against

	p.Title = alias + " - diff"

	if on == "" || against == "" {
		return
	}

	content := &db.FileContent{Username: username, Alias: alias}

	diffs, err := file.Diff(content, on, against)
	if err != nil {
		return p.setError(w, err)
	}

	// p.Data["path"] = storage.Staged.Path
	p.Data["diffs"] = diffs

	return
}

func diffHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "diff.tmpl",
		loadData:     loadDiff,
	})
}
