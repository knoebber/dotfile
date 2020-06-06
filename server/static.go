package server

import "net/http"

func getIndexHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "index.tmpl",
		title:        indexTitle,
	})
}

func getAboutHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "about.tmpl",
		title:        aboutTitle,
	})
}

func getAcknowledgmentsHander() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "acknowledgments.tmpl",
		title:        "Acknowledgments",
	})
}
