package server

import "net/http"

func aboutHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "about.tmpl",
		title:        aboutTitle,
	})
}

func acknowledgmentsHander() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "acknowledgments.tmpl",
		title:        "Acknowledgments",
	})
}
