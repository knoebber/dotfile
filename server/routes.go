package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

func setupRoutes(r *mux.Router) {
	assetRoutes(r)
}

func assetRoutes(r *mux.Router) {
	assets := http.FileSystem(http.Dir("assets/"))
	r.Path("/").Handler(http.FileServer(assets))
	r.Path("/style.css").Handler(http.FileServer(assets))
	r.Path("/favicon.ico").Handler(http.FileServer(assets))
}
