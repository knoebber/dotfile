package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

func setupRoutes(r *mux.Router) {
	publicRoutes(r)
	assetRoutes(r)
}

func publicRoutes(r *mux.Router) {
	r.HandleFunc("/", handleIndex)
	r.HandleFunc("/about.html", handleAbout)
	r.HandleFunc("/login.html", handleLogin)
}

func assetRoutes(r *mux.Router) {
	assets := http.FileSystem(http.Dir("assets/"))
	r.Path("/style.css").Handler(http.FileServer(assets))
	r.Path("/favicon.ico").Handler(http.FileServer(assets))
}
