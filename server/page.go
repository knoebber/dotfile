package server

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/dotfile/file"
	"github.com/knoebber/dotfile/usererr"
	"github.com/pkg/errors"
)

// Titles of pages that live on the nav bar.
// Important these stay constant as they are referenced in setLinks()
const (
	indexTitle    = "Dotfilehub"
	aboutTitle    = "About"
	settingsTitle = "Settings"
	signupTitle   = "Signup"
	loginTitle    = "Login"
)

var templates *template.Template

// Page is used for rendering pages and tracking request state.
// Exported fields/methods may be used within templates.
type Page struct {
	Title          string
	SuccessMessage string
	ErrorMessage   string
	Links          []Link
	Vars           map[string]string
	Data           map[string]interface{}

	Session      *db.Session
	templateName string

	// When true restrict page access to logged in page owners.
	protected bool
}

// Owned returns whether the logged in user owns the page.
// Pages without the {username} var are owned by every user.
func (p *Page) Owned() bool {
	pageOwner := p.Vars["username"]

	return pageOwner == "" ||
		(p.Session != nil && strings.EqualFold(p.Session.Username, pageOwner))
}

func (p *Page) flashSuccess(msg string) {
	p.SuccessMessage = msg
}

func (p *Page) setError(w http.ResponseWriter, err error) (done bool) {
	if db.NotFound(err) {
		// TODO create not found template.
		setError(w, err, "Not found", http.StatusNotFound)
		return true
	}

	var uerr *usererr.Error
	if errors.As(err, &uerr) {
		log.Printf("flashing %s error: %s", uerr.Reason, err)
		p.ErrorMessage = uerr.Message
	} else {
		log.Print("flashing fallback from unexpected error: ", err)
		p.ErrorMessage = "Unexpected error - if this continues please contact an admin."
	}

	return false
}

func (p *Page) setSession(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(sessionCookie)
	if errors.Is(err, http.ErrNoCookie) {
		return nil
	} else if err != nil {
		return err
	}

	p.Session, err = db.CheckSession(cookie.Value, r.RemoteAddr)
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
		newLink("/about", aboutTitle, p.Title),
		settingsLink,
	}
}

func (p *Page) render(w http.ResponseWriter) error {
	return templates.ExecuteTemplate(w, p.templateName, p)
}

func newPage(w http.ResponseWriter, r *http.Request, templateName, title string, protected bool) (*Page, error) {
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

	p.setLinks()

	return p, nil
}

func loadTemplates() (err error) {
	templates, err = template.New("dotfilehub").Funcs(template.FuncMap{
		// Global functions that templates can call.
		"shortenEqualText": file.ShortenEqualText,
	}).ParseGlob("tmpl/*.tmpl")

	return
}
