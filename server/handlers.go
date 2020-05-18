package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pkg/errors"
)

const (
	sessionCookie = "dotfilehub-session"
)

type pageBuilder func(w http.ResponseWriter, r *http.Request, p *Page) (done bool)

type pageDescription struct {
	templateName string
	title        string

	loadData   pageBuilder
	handleForm pageBuilder

	protected bool
}

// Closure that creates http.HandlerFunc's.
func createHandler(desc *pageDescription) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, err := newPage(w, r, desc.templateName, desc.title, desc.protected)
		if err != nil {
			pageError(w, desc.title, err)
			return
		}

		if page.protected && page.Session == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Optionally handle a form.
		if r.Method == "POST" && desc.handleForm != nil {
			if err := r.ParseForm(); err != nil {
				badRequest(w, errors.Wrapf(err, "parsing form %#v", page.Title))
				return
			}

			if desc.handleForm(w, r, page) {
				// Returns true when the form handler wrote the response writer.
				// Common case is the form set a redirect.
				// No need to load more data or render the template in this case.
				return
			}
		}

		// Optionally call a function to load data into page.Data.
		// This is for templates to use when rendering a view.
		if desc.loadData != nil && desc.loadData(w, r, page) {
			return
		}

		if err := page.render(w); err != nil {
			templateError(w, page.Title, err)
		}
	}
}

func pageError(w http.ResponseWriter, title string, err error) {
	setError(
		w,
		errors.Wrapf(err, "creating page %#v", title),
		fmt.Sprintf("Failed to create page %#v", title),
		http.StatusInternalServerError,
	)
}

func templateError(w http.ResponseWriter, title string, err error) {
	log.Print(errors.Wrapf(err, "template error: rendering %#v", title))
}

func badRequest(w http.ResponseWriter, err error) {
	setError(w, err, "Request is invalid", http.StatusBadRequest)
}

func setError(w http.ResponseWriter, err error, errMsg string, status int) {
	log.Print(err)
	http.Error(w, errMsg, status)
}
