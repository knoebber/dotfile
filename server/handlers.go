package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/dotfile/usererr"
	"github.com/pkg/errors"
)

const (
	minPassLength = 8
	sessionCookie = "dotfilehub-session"
)

// If form handler returns an error then the same form will be rendered again with a flash error.
// When there is no error it's assumed that the response is set.
type formHandler func(w http.ResponseWriter, r *http.Request) error

func createStaticHandler(templateName, title string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, err := newPage(r, templateName, title)
		if err != nil {
			pageError(w, title, err)
		}

		if err := page.render(w); err != nil {
			pageError(w, title, err)
		}
	}
}

func createFormHandler(handleForm formHandler, templateName, title string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, err := newPage(r, templateName, title)
		if err != nil {
			pageError(w, title, err)
		}

		if r.Method == "PUT" || r.Method == "POST" {
			if err := r.ParseForm(); err != nil {
				badRequest(w, errors.Wrapf(err, "parsing form %#v", title))
				return
			}

			// Set a flash error.
			page.ErrorMessage = checkErr(handleForm(w, r))
		}

		if err := page.render(w); err != nil {
			pageError(w, title, err)
		}
	}
}

func createHandleLogin(secure bool) formHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		s, err := db.UserLogin(r.Form.Get("username"), r.Form.Get("password"))

		if err != nil {
			// Print the real error and show the user a generic catch all.
			log.Print(err)
			return usererr.Invalid("Username or password is incorrect.")
		}
		http.SetCookie(w, &http.Cookie{
			Name:     sessionCookie,
			Value:    s.Session,
			Secure:   secure,
			HttpOnly: true,
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}
}

func handleSignup(w http.ResponseWriter, r *http.Request) error {
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	confirm := r.Form.Get("confirm")

	if len(password) < minPassLength {
		return usererr.Invalid(fmt.Sprintf("Password must be at least %d characters.", minPassLength))
	}

	if password != confirm {
		return usererr.Invalid("Passwords do not match")
	}

	_, err := db.CreateUser(username, password, nil)
	if err != nil {
		return err
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
	return nil
}

// Returns a friendly error message when error is not nil.
func checkErr(err error) string {
	if err == nil {
		return ""
	}

	log.Print(err)

	if uerr, ok := err.(usererr.Messager); ok {
		return uerr.Message()
	}

	return "Unexpected error - if this continues please contact an admin."
}

func pageError(w http.ResponseWriter, title string, err error) {
	setError(
		w,
		errors.Wrapf(err, "rendering page %#v", title),
		"Failed to render template",
		http.StatusInternalServerError,
	)
}

func badRequest(w http.ResponseWriter, err error) {
	setError(w, err, "Request is invalid", http.StatusBadRequest)
}

func setError(w http.ResponseWriter, err error, errMsg string, status int) {
	log.Print(err)
	http.Error(w, errMsg, status)
}
