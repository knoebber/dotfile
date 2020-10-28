package server

import (
	"net/http"

	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/dotfile/usererror"
)

func handleEmail(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	if err := db.UpdateEmail(db.Connection, p.userID(), r.Form.Get("email")); err != nil {
		return p.setError(w, err)
	}
	p.Data["email"] = r.Form.Get("email")

	p.flashSuccess("Updated email")
	return
}

func handleTokenForm(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	if token := r.Form.Get("token"); token != "" {
		if err := db.RotateCLIToken(db.Connection, p.Session.UserID, token); err != nil {
			return p.setError(w, err)
		}

		p.flashSuccess("Generated new token")
		return
	}

	password := r.Form.Get("password")
	if err := db.CheckPassword(db.Connection, p.Username(), password); err != nil {
		return p.setError(w, err)
	}

	p.Data["authenticated"] = true
	return
}

func handlePassword(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	currentPass := r.Form.Get("current")

	newPass := r.Form.Get("new")
	confirm := r.Form.Get("confirm")

	if newPass != confirm {
		return p.setError(w, usererror.Invalid("Confirm does not match."))
	}

	if err := db.UpdatePassword(db.Connection, p.Username(), currentPass, newPass); err != nil {
		return p.setError(w, err)
	}

	p.flashSuccess("Updated password")
	return
}

func handleDeleteUser(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	if username != p.Session.Username {
		return p.setError(w, usererror.Invalid("Username does not match"))
	}

	if err := db.DeleteUser(username, password); err != nil {
		return p.setError(w, err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
	return true
}

func handleTheme(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	newTheme := db.UserTheme(r.Form.Get("theme"))

	if err := db.UpdateTheme(db.Connection, p.userID(), newTheme); err != nil {
		return p.setError(w, err)
	}
	p.Session.Theme = newTheme

	return
}

func handleTimezone(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	if err := db.UpdateTimezone(db.Connection, p.Session.UserID, r.Form.Get("timezone")); err != nil {
		return p.setError(w, err)
	}

	p.flashSuccess("Updated timezone")
	return
}

func loadUserCLI(config Config) pageBuilder {
	return func(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
		p.Data["remote"] = config.URL(r)
		return
	}
}

func loadThemes(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	p.Data["themes"] = []db.UserTheme{
		db.UserThemeLight,
		db.UserThemeDark,
	}
	return
}

func loadUserFiles(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	username := p.Vars["username"]
	p.Title = username

	if err := db.ValidateUserExists(db.Connection, username); err != nil {
		return p.setError(w, err)
	}

	files, err := db.FilesByUsername(db.Connection, username, p.Timezone())
	if db.NotFound(err) {
		return
	} else if err != nil {
		return p.setError(w, err)
	}

	p.Data["files"] = files
	return
}

func userHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "user.tmpl",
		loadData:     loadUserFiles,
	})
}

func settingsHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "user_settings.tmpl",
		title:        settingsTitle,
		handleForm:   handleEmail,
		protected:    true,
	})
}

func cliHandler(config Config) http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "cli.tmpl",
		title:        "CLI Setup",
		loadData:     loadUserCLI(config),
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
		handleForm:   handleEmail,
		protected:    true,
	})
}

func timezoneHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "timezone.tmpl",
		title:        "Set Timezone",
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

func deleteUserHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "delete_user.tmpl",
		title:        "Delete Account",
		handleForm:   handleDeleteUser,
		protected:    true,
	})
}
