package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

func setupRoutes(r *mux.Router) {
	assetRoutes(r)
	staticRoutes(r)
	formRoutes(r)
}

func assetRoutes(r *mux.Router) {
	assets := http.FileSystem(http.Dir("assets/"))
	r.Path("/style.css").Handler(http.FileServer(assets))
	r.Path("/favicon.ico").Handler(http.FileServer(assets))
}

func staticRoutes(r *mux.Router) {
	r.HandleFunc("/", createStaticHandler("index.tmpl", indexTitle))
	r.HandleFunc("/about", createStaticHandler("about.tmpl", aboutTitle))
	r.HandleFunc("/login", createStaticHandler("login.tmpl", loginTitle)).Methods("GET")
	r.HandleFunc("/signup", createStaticHandler("signup.tmpl", "Signup")).Methods("GET")
}

func formRoutes(r *mux.Router) {
	r.HandleFunc("/login", handleLogin).Methods("POST")
	r.HandleFunc("/signup", handleSignup).Methods("POST")
}
