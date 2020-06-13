package server

import (
	"fmt"
	"net/http"

	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/dotfile/usererr"
)

func handleEmail(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	if err := db.UpdateEmail(p.Session.UserID, r.Form.Get("email")); err != nil {
		return p.setError(w, err)
	}
	p.Data["email"] = r.Form.Get("email")

	p.flashSuccess("Updated email")
	return
}

func handlePassword(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	currentPass := r.Form.Get("current")

	newPass := r.Form.Get("new")
	confirm := r.Form.Get("confirm")

	if len(newPass) < minPassLength {
		return p.setError(w, usererr.Invalid(fmt.Sprintf("Password must be %d or more characters.", minPassLength)))
	}

	if newPass != confirm {
		return p.setError(w, usererr.Invalid("Confirm does not match."))
	}

	if err := db.UpdatePassword(p.Session.Username, currentPass, newPass); err != nil {
		return p.setError(w, err)
	}

	p.flashSuccess("Updated password")
	return
}

func handleTheme(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	newTheme := db.UserTheme(r.Form.Get("theme"))

	if err := db.UpdateTheme(p.Session.Username, newTheme); err != nil {
		return p.setError(w, err)
	}
	p.Session.Theme = newTheme

	return
}

func loadUserSettings(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	var email string
	username := p.Session.Username

	user, err := db.GetUser(username)

	if err != nil {
		return p.setError(w, err)
	}

	if user.Email != nil {
		email = *user.Email
	}

	p.Data["email"] = email
	p.Data["joined"] = user.JoinDate()
	return
}

func loadThemes(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	p.Data["themes"] = []db.UserTheme{
		db.UserThemeLight,
		db.UserThemeDark,
	}
	return
}

func settingsHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "settings.tmpl",
		title:        settingsTitle,
		loadData:     loadUserSettings,
		handleForm:   handleEmail,
		protected:    true,
	})
}

func themeHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "theme.tmpl",
		title:        "Set Theme",
		loadData:     loadThemes,
		handleForm:   handleTheme,
		protected:    true,
	})
}

func emailHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "email.tmpl",
		title:        "Set Email",
		loadData:     loadUserSettings,
		handleForm:   handleEmail,
		protected:    true,
	})
}

func passwordHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "password.tmpl",
		title:        "Update Password",
		handleForm:   handlePassword,
		protected:    true,
	})
}
