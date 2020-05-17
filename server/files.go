package server

import "net/http"

func getNewFileHandler() http.HandlerFunc {
	return createHandler(&pageDescription{
		templateName: "new_file.tmpl",
		title:        newFileTitle,
	})
}
