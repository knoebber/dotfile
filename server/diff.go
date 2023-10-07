package server

import (
	"html"
	"html/template"
	"net/http"
	"strings"

	"github.com/hexops/gotextdiff"
	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/dotfile/dotfile"
)

// Returns HTML that is ready to be added to a template.
func getHtmlDiff(content *db.FileContent, on, against string) (template.HTML, error) {
	var buff strings.Builder

	unified, err := dotfile.Diff(content, on, against)
	if err != nil {
		return "", err
	}

	for _, hunk := range unified.Hunks {
		if len(unified.Hunks) > 1 {
			_, _ = buff.WriteString("<hr/><strong>HUNK</strong><hr/>")
		}
		for _, line := range hunk.Lines {
			text := html.EscapeString(line.Content)
			switch line.Kind {
			case gotextdiff.Insert:
				_, _ = buff.WriteString("<ins>")
				_, _ = buff.WriteString(text)
				_, _ = buff.WriteString("</ins>")
			case gotextdiff.Delete:
				_, _ = buff.WriteString("<del>")
				_, _ = buff.WriteString(text)
				_, _ = buff.WriteString("</del>")
			case gotextdiff.Equal:
				_, _ = buff.WriteString("<span>")
				_, _ = buff.WriteString(text)
				_, _ = buff.WriteString("</span>")
			}
		}
	}
	return template.HTML(buff.String()), nil
}

// Loads a diff: ?on VS ?against.
func loadDiff(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	// TODO add line numbers to diff template
	alias := p.Vars["alias"]
	username := p.Vars["username"]

	on := r.URL.Query().Get("on")
	against := r.URL.Query().Get("against")

	commits, err := db.CommitList(db.Connection, username, alias, p.Timezone())
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
	diff, err := getHtmlDiff(&db.FileContent{Connection: db.Connection, Username: username, Alias: alias}, on, against)

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
