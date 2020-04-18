package server

import (
	"html/template"
	"net/http"

	"github.com/knoebber/dotfile/db"
	"github.com/pkg/errors"
)

var templates *template.Template

// Page is used for rendering pages and tracking request state.
// Exported fields/methods are called within templates.
type Page struct {
	Title        string
	ErrorMessage string
	Links        []Link

	templateName string
	title        string
	session      *db.Session
}

func (p *Page) setSession(r *http.Request) error {
	if p.session != nil {
		return nil
	}

	cookie, err := r.Cookie(sessionCookie)
	if errors.Is(err, http.ErrNoCookie) {
		return nil
	} else if err != nil {
		return err
	}

	p.session, err = db.GetSession(cookie.Value)
	if err != nil {
		return err
	}

	return nil
}

func (p *Page) setLinks() {
	var userLink Link

	username := p.session.Username

	if p.session != nil {
		userLink = newLink("/"+username, username, username)
	} else {
		userLink = newLink("/login", loginTitle, p.Title)
	}

	p.Links = []Link{
		userLink,
		newLink("/about", aboutTitle, p.Title),
		newLink("/login", loginTitle, p.Title),
	}
}

func (p *Page) render(w http.ResponseWriter) error {
	return templates.ExecuteTemplate(w, p.templateName, p)
}

func newPage(r *http.Request, templateName, title string) (*Page, error) {
	p := &Page{
		Title:        title,
		templateName: templateName,
	}

	if err := p.setSession(r); err != nil {
		return nil, err
	}

	p.setLinks()

	return p, nil
}

func loadTemplates() (err error) {
	templates, err = template.ParseGlob("tmpl/*.tmpl")
	return
}
