package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/dotfile/usererror"
)

const minPassLength = 8

// Redirects to index when session exists.
func checkSession(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	if p.Session != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return true
	}
	return false
}

func createHandleLogin(secure bool) pageBuilder {
	return func(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
		return login(w, r, p, secure)
	}
}

func login(w http.ResponseWriter, r *http.Request, p *Page, secure bool) (done bool) {
	s, err := db.UserLogin(r.Form.Get("username"), r.Form.Get("password"))
	if err != nil {
		// Print the real error and show the user a generic catch all.
		log.Print(err)
		return p.setError(w, usererror.Invalid("Username or password is incorrect."))
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    s.Session,
		Secure:   secure,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
	return true
}

func createHandleSignup(secure bool) pageBuilder {
	return func(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
		username := r.Form.Get("username")
		password := r.Form.Get("password")
		confirm := r.Form.Get("confirm")

		if len(password) < minPassLength {
			msg := fmt.Sprintf("Password must be at least %d characters.", minPassLength)
			return p.setError(w, usererror.Invalid(msg))
		}

		if password != confirm {
			p.setError(w, usererror.Invalid("Passwords do not match."))
			return false
		}

		_, err := db.CreateUser(username, password)
		if err != nil {
			return p.setError(w, err)
		}

		return login(w, r, p, secure)
	}
}

func handleLogout(w http.ResponseWriter, r *http.Request, p *Page) (done bool) {
	if err := db.Logout(p.Session.ID); err != nil {
		log.Print(err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:   sessionCookie,
		MaxAge: -1,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
	return true
}

func loginHandler(secure bool) http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "login.tmpl",
		title:        loginTitle,
		loadData:     checkSession,
		handleForm:   createHandleLogin(secure),
	})
}

func signupHandler(secure bool) http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "signup.tmpl",
		title:        "Signup",
		loadData:     checkSession,
		handleForm:   createHandleSignup(secure),
	})
}

func logoutHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		handleForm: handleLogout,
		protected:  true,
	})
}
