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

var rootTemplate *template.Template

// Root is the data the root template expects.
type Root struct {
	Title string
	Links []Link
}

// Static is the data a static page expects.
type Static struct {
	Root
	Body string
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

func loadTemplates() (err error) {
	rootTemplate, err = template.ParseFiles("tmpl/root.html")
	return
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
		newLink("/about.html", aboutTitle, currentTitle),
		newLink("/login.html", loginTitle, currentTitle),
	}
}

func newStatic(title, body string) *Static {
	s := new(Static)
	s.Title = title
	s.Body = body
	s.Links = getStaticLinks(title)
	return s
}

// Renders pages without dynamic content.
func renderStatic(w http.ResponseWriter, s *Static) {
	err := rootTemplate.ExecuteTemplate(w, "root.html", s)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
