package server

import "net/http"

func createStaticHandler(templateName, title string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderStatic(w, templateName, title)
	}
}
