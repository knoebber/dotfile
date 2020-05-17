package server

import (
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/dotfile/usererr"
	"github.com/pkg/errors"
)

// Titles of static pages.
const (
	indexTitle    = "Dotfilehub"
	aboutTitle    = "About"
	loginTitle    = "Login"
	signupTitle   = "Signup"
	exploreTitle  = "Explore"
	emailTitle    = "Update Email"
	passwordTitle = "Update Password"
)

var templates *template.Template

// Page is used for rendering pages and tracking request state.
// Exported fields/methods are called within templates.
type Page struct {
	Title          string
	SuccessMessage string
	ErrorMessage   string
	Links          []Link
	Vars           map[string]string
	Data           map[string]string

	Session      *db.Session
	templateName string
}

// Owned returns whether the current logged in user owns the page.
func (p *Page) Owned() bool {
	return p.Session != nil && p.Session.Username == p.Vars["username"]
}

func (p *Page) flashSuccess(msg string) {
	p.SuccessMessage = msg
}

func (p *Page) setError(w http.ResponseWriter, err error) (done bool) {
	if db.NotFound(err) {
		setError(w, err, "Not found", http.StatusNotFound)
		return true
	}

	if uerr, ok := err.(usererr.Messager); ok {
		log.Printf("flashing %s", err)
		p.ErrorMessage = uerr.Message()
	} else {
		log.Print("flashing fallback from unexpected error: ", err)
		p.ErrorMessage = "Unexpected error - if this continues please contact an admin."
	}

	return false
}

func (p *Page) setSession(w http.ResponseWriter, r *http.Request) error {
	if p.Session != nil {
		return nil
	}

	cookie, err := r.Cookie(sessionCookie)
	if errors.Is(err, http.ErrNoCookie) {
		return nil
	} else if err != nil {
		return err
	}

	p.Session, err = db.CheckSession(cookie.Value, r.RemoteAddr)
	if db.NotFound(err) {
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
	var userLink Link

	if p.Session != nil {
		username := p.Session.Username
		userLink = newLink("/"+username, "Profile", p.Title)
	} else {
		userLink = newLink("/login", loginTitle, p.Title)
	}

	p.Links = []Link{
		newLink("/", indexTitle, p.Title),
		newLink("/explore", exploreTitle, p.Title),
		newLink("/about", aboutTitle, p.Title),
		userLink,
	}
}

func (p *Page) render(w http.ResponseWriter) error {
	return templates.ExecuteTemplate(w, p.templateName, p)
}

func newPage(w http.ResponseWriter, r *http.Request, templateName, title string) (*Page, error) {
	p := &Page{
		Title:        title,
		Vars:         mux.Vars(r),
		Data:         make(map[string]string),
		templateName: templateName,
	}

	if err := p.setSession(w, r); err != nil {
		return nil, err
	}

	p.setLinks()

	return p, nil
}

func loadTemplates() (err error) {
	templates, err = template.ParseGlob("tmpl/*.tmpl")
	return
}