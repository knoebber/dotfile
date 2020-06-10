package server

import (
	"net/http"

	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/dotfile/file"
)

// Loads a diff: ?on VS ?against.
func loadDiff(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	alias := p.Vars["alias"]

	storage, err := db.NewReadOnlyStorage(p.Vars["username"], alias)
	if err != nil {
		return p.setError(w, err)
	}

	commits, err := db.GetCommitList(p.Vars["username"], alias)
	if err != nil {
		return p.setError(w, err)
	}

	on := r.URL.Query().Get("on")
	against := r.URL.Query().Get("against")

	diffs, err := file.Diff(storage, on, against)
	if err != nil {
		return p.setError(w, err)
	}

	p.Data["diffs"] = diffs
	p.Data["alias"] = alias
	p.Data["path"] = storage.Staged.Path
	p.Data["hash"] = on
	p.Data["against"] = against
	p.Data["commits"] = commits
	p.Title = alias + " - diff"
	return
}

func diffHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "diff.tmpl",
		loadData:     loadDiff,
	})
}
