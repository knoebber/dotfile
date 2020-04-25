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
}

func createHandler(desc *pageDescription) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, err := newPage(w, r, desc.templateName, desc.title)
		if err != nil {
			pageError(w, desc.title, err)
			return
		}

		if desc.loadData != nil && desc.loadData(w, r, page) {
			return
		}

		if r.Method == "POST" && desc.handleForm != nil {
			if err := r.ParseForm(); err != nil {
				badRequest(w, errors.Wrapf(err, "parsing form %#v", page.Title))
				return
			}

			if desc.handleForm(w, r, page) {
				return
			}
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
