package server

import (
	"fmt"
	"net/http"

	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/dotfile/usererror"
)

func handleEmail(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	if err := db.UpdateEmail(p.Session.UserID, r.Form.Get("email")); err != nil {
		return p.setError(w, err)
	}
	p.Data["email"] = r.Form.Get("email")

	p.flashSuccess("Updated email")
	return
}

func handleTokenForm(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	if token := r.Form.Get("token"); token != "" {
		if err := db.RotateToken(p.Session.UserID, token); err != nil {
			return p.setError(w, err)
		}

		p.flashSuccess("Generated new token")
		return
	}

	password := r.Form.Get("password")
	if err := db.CheckPassword(p.Session.Username, password); err != nil {
		return p.setError(w, err)
	}

	p.Data["authenticated"] = true
	return
}

func handlePassword(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	currentPass := r.Form.Get("current")

	newPass := r.Form.Get("new")
	confirm := r.Form.Get("confirm")

	if len(newPass) < minPassLength {
		return p.setError(w, usererror.Invalid(fmt.Sprintf("Password must be %d or more characters.", minPassLength)))
	}

	if newPass != confirm {
		return p.setError(w, usererror.Invalid("Confirm does not match."))
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

func handleTimezone(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	if err := db.UpdateTimezone(p.Session.UserID, r.Form.Get("timezone")); err != nil {
		return p.setError(w, err)
	}

	p.flashSuccess("Updated timezone")
	return
}

func createLoadUserCLI(config Config) pageBuilder {
	return func(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
		var remote string

		user, err := db.GetUser(p.Session.Username)
		if err != nil {
			return p.setError(w, err)
		}

		p.Data["token"] = user.CLIToken
		if config.Secure {
			remote = "https://"
		} else {
			remote = "http://"
		}
		if config.Host != "" {
			remote += config.Host
		} else {
			remote += r.Host
		}

		p.Data["remote"] = remote

		return
	}
}

func loadUserSettings(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	username := p.Session.Username

	user, err := db.GetUser(username)

	if err != nil {
		return p.setError(w, err)
	}

	p.Data["user"] = user
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

func cliHandler(config Config) http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "cli.tmpl",
		title:        "CLI Setup",
		loadData:     createLoadUserCLI(config),
		handleForm:   handleTokenForm,
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

func timezoneHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "timezone.tmpl",
		title:        "Set Timezone",
		loadData:     loadUserSettings,
		handleForm:   handleTimezone,
		protected:    true,
	})
}

func passwordHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "password.tmpl",
		title:        "Set Password",
		handleForm:   handlePassword,
		protected:    true,
	})
}
