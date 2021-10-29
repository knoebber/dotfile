package server

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/dotfile/dotfile"
	"github.com/knoebber/usererror"
	"github.com/pkg/errors"
)

// Titles of pages that have links on the nav bar.
// Important these stay constant as they are referenced in setLinks()
const (
	indexTitle    = "Dotfilehub"
	aboutTitle    = "About"
	settingsTitle = "Settings"
	signupTitle   = "Signup"
	loginTitle    = "Login"
)

var (
	baseTemplate  *template.Template
	pageTemplates *template.Template
)

// Page renders pages and tracks request state.
// Exported fields/methods may be used within templates.
type Page struct {
	Title          string
	SuccessMessage string
	ErrorMessage   string
	Links          []Link
	Vars           map[string]string
	Data           map[string]interface{}

	Table        *db.HTMLTable
	Session      *db.UserSession
	templateName string
	htmlFile     string
	// Page access is restricted to their owners when true.
	protected bool
}

// Dark returns whether dark mode is turned on.
func (p *Page) Dark() bool {
	return p.Session != nil && p.Session.Theme == db.UserThemeDark
}

// Owned returns whether the logged in user owns the page.
// Pages without the {username} var are owned by every user.
func (p *Page) Owned() bool {
	pageOwner := p.Vars["username"]

	return pageOwner == "" ||
		(p.Session != nil && strings.EqualFold(p.Session.Username, pageOwner))
}

// Timezone returns the logged in user's timezone.
func (p *Page) Timezone() *string {
	if p.Session != nil {
		return p.Session.Timezone
	}
	return nil
}

// Username returns the logged in user's username.
func (p *Page) Username() string {
	if p.Session != nil {
		return p.Session.Username
	}
	return ""
}

// Email returns the logged in user's email.
func (p *Page) Email() string {
	if p.Session != nil && p.Session.Email != nil {
		return *p.Session.Email
	}

	return ""
}

// Theme returns the logged in user's theme.
func (p *Page) Theme() db.UserTheme {
	if p.Session != nil {
		return p.Session.Theme
	}

	return ""
}

// CLIToken returns the logged in user's CLI token.
func (p *Page) CLIToken() string {
	if p.Session != nil {
		return p.Session.CLIToken
	}

	return ""
}

// UserCreatedAt returns the logged in user's creation date.
func (p *Page) UserCreatedAt() string {
	if p.Session != nil {
		return p.Session.UserCreatedAt
	}

	return ""
}

func (p *Page) session() string {
	if p.Session != nil {
		return p.Session.Session
	}
	return ""
}

func (p *Page) userID() int64 {
	if p.Session != nil {
		return p.Session.UserID
	}

	return -1
}

func (p *Page) flashSuccess(msg string) {
	p.SuccessMessage = msg
}

func (p *Page) setError(w http.ResponseWriter, err error) (done bool) {
	if db.NotFound(err) {
		log.Printf("resource not found: %s", err)
		w.WriteHeader(http.StatusNotFound)
		p.htmlFile = "404.html"
		if err := p.writeFromHTML(w); err != nil {
			staticError(w, "404", err)
		}

		return true
	}

	if uErr := usererror.Convert(err); uErr != nil {
		log.Printf("flashing usererror: %s", err)
		p.ErrorMessage = uErr.Message
	} else {
		log.Print("flashing fallback from unexpected error: ", err)
		p.ErrorMessage = "Unexpected error - if this continues please contact an admin."
	}

	return false
}

// Sets p.Session when session cookie exists.
func (p *Page) setSession(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(sessionCookie)
	if errors.Is(err, http.ErrNoCookie) {
		return nil
	} else if err != nil {
		return err
	}

	p.Session, err = db.Session(db.Connection, cookie.Value)
	if db.NotFound(err) {
		// Session in cookie does not exist in DB.
		// Unset it.
		http.SetCookie(w, &http.Cookie{
			Name:   sessionCookie,
			MaxAge: -1,
		})
		return nil
	} else if err != nil {
		return err
	}

	return nil
}

func (p *Page) setLinks() {
	var (
		userLink     Link
		settingsLink Link
	)

	if p.Session != nil {
		username := p.Session.Username
		userLink = newLink("/"+username, username, p.Title)
		settingsLink = newLink("/settings", settingsTitle, p.Title)
	} else {
		userLink = newLink("/login", loginTitle, p.Title)
		settingsLink = newLink("/signup", signupTitle, p.Title)
	}

	p.Links = []Link{
		newLink("/", indexTitle, p.Title),
		userLink,
		newLink("/README.org", aboutTitle, p.Title),
		settingsLink,
	}
}

// Writes content from the templates/ directory.
func (p *Page) writeFromTemplate(w io.Writer) error {
	p.setLinks()

	baseClone, err := baseTemplate.Clone()
	if err != nil {
		return err
	}

	baseClone.Funcs(template.FuncMap{
		"content": func() (template.HTML, error) {
			buf := new(bytes.Buffer)
			err := pageTemplates.ExecuteTemplate(buf, p.templateName, p)
			if err != nil {
				return "", err
			}

			return template.HTML(buf.String()), nil
		},
	})

	return baseClone.Execute(w, p)
}

// Writes content from the html/ directory.
func (p *Page) writeFromHTML(w io.Writer) error {
	p.setLinks()

	baseClone, err := baseTemplate.Clone()
	if err != nil {
		return err
	}

	baseClone.Funcs(template.FuncMap{
		"content": func() (template.HTML, error) {
			html, err := serverContent.ReadFile(filepath.Join("html", p.htmlFile))
			if err != nil {
				return "", err
			}

			return template.HTML(html), err
		},
	})

	return baseClone.Execute(w, p)
}

// Returns a page that will use a go template for content.
func pageFromTemplate(w http.ResponseWriter, r *http.Request, templateName, title string, protected bool) (*Page, error) {
	p := &Page{
		Title:        title,
		Vars:         mux.Vars(r),
		Data:         make(map[string]interface{}),
		templateName: templateName,
		protected:    protected,
	}

	if err := p.setSession(w, r); err != nil {
		return nil, err
	}

	return p, nil
}

// Returns a page that will use a HTML file for content.
func pageFromHTML(w http.ResponseWriter, r *http.Request, title, htmlFile string) (*Page, error) {
	p := &Page{
		Title:    title,
		htmlFile: htmlFile,
	}

	if err := p.setSession(w, r); err != nil {
		return nil, err
	}

	return p, nil
}

func loadTemplates() (err error) {
	defaultContentFunction := template.FuncMap{
		"content": func() (string, error) {
			return "", errors.New("content is not set")
		},
	}

	baseTemplate, err = template.
		New("base").
		Funcs(defaultContentFunction).
		ParseFS(serverContent, "templates/base.tmpl")

	if err != nil {
		return
	}

	pageFunctions := template.FuncMap{
		// Global functions that page templates can call.
		"shortenHash": dotfile.ShortenHash,
	}

	pageTemplates, err = template.
		New("pages").
		Funcs(pageFunctions).
		ParseFS(serverContent, "templates/*/*.tmpl")
	return

}
