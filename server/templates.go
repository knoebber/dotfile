package server

import (
	"html/template"
	"net/http"
)

const (
	indexTitle = "Dotfilehub"
	aboutTitle = "About"
	loginTitle = "Login"
)

var templates *template.Template

// Root is the common data that all templates expect.
type Root struct {
	Title string
	Links []Link

	templateName string
}

// Link populates a navbar link.
type Link struct {
	URL    string
	Title  string
	Active bool
}

// GetClass returns a class name for active styling.
func (l *Link) GetClass() string {
	if l.Active {
		return "active"
	}
	return ""
}

func newLink(url, title, currentTitle string) Link {
	return Link{
		URL:    url,
		Title:  title,
		Active: title == currentTitle,
	}
}

func getStaticLinks(currentTitle string) []Link {
	return []Link{
		newLink("/", indexTitle, currentTitle),
		newLink("/about", aboutTitle, currentTitle),
		newLink("/login", loginTitle, currentTitle),
	}
}

func newStaticView(title string) *Root {
	return &Root{
		Title: title,
		Links: getStaticLinks(title),
	}
}

func loadTemplates() (err error) {
	templates, err = template.ParseGlob("tmpl/*")

	return
}

// Renders pages without dynamic content.
func renderStatic(w http.ResponseWriter, templateName, title string) error {
	return templates.ExecuteTemplate(w, templateName, newStaticView(title))
}
