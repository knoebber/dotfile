package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

func setupRoutes(r *mux.Router) {
	assetRoutes(r)
	staticRoutes(r)
}

func assetRoutes(r *mux.Router) {
	assets := http.FileSystem(http.Dir("assets/"))
	r.Path("/style.css").Handler(http.FileServer(assets))
	r.Path("/favicon.ico").Handler(http.FileServer(assets))
}

func staticRoutes(r *mux.Router) {
	r.HandleFunc("/", createStaticHandler("index.tmpl", indexTitle))
	r.HandleFunc("/about.html", createStaticHandler("about.tmpl", aboutTitle))
	r.HandleFunc("/login.html", createStaticHandler("login.tmpl", loginTitle))
}
