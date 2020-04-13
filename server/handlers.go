package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/knoebber/dotfile/db"
	"github.com/pkg/errors"
)

const minPasswordLength = 8

func createStaticHandler(templateName, title string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := renderStatic(w, templateName, title); err != nil {
			internalError(w, errors.Wrapf(err, "rendering static page %#v", title))
		}
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		badRequest(w, errors.Wrap(err, "parsing login form"))
		return
	}

	// TODO set cookie with session id.
	_, err := db.UserLogin(r.Form.Get("username"), r.Form.Get("password"))
	if err != nil {
		setError(w, err, "Invalid username/password", http.StatusUnauthorized)
		return
	}
}

func handleSignup(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		badRequest(w, errors.Wrap(err, "parsing signup form"))
		return
	}

	username := r.Form.Get("username")
	password := r.Form.Get("password")

	if len(password) < minPasswordLength {
		badRequest(w, fmt.Errorf("password must be at least %d characters", minPasswordLength))
		return
	}

	_, err := db.CreateUser(username, password, nil)
	if err != nil {
		internalError(w, err)
		return
	}
}

func internalError(w http.ResponseWriter, err error) {
	setError(w, err, "Unexpected error :(", http.StatusInternalServerError)
}

func badRequest(w http.ResponseWriter, err error) {
	setError(w, err, "Request is invalid", http.StatusBadRequest)
}

func setError(w http.ResponseWriter, err error, errMsg string, status int) {
	log.Print(err)
	http.Error(w, errMsg, status)
}
