package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

func setupRoutes(r *mux.Router, secure bool) {
	assetRoutes(r)
	staticRoutes(r)
	authRoutes(r, secure)
}

func assetRoutes(r *mux.Router) {
	assets := http.FileSystem(http.Dir("assets/"))
	r.Path("/style.css").Handler(http.FileServer(assets))
	r.Path("/favicon.ico").Handler(http.FileServer(assets))
}

func staticRoutes(r *mux.Router) {
	r.HandleFunc("/", createStaticHandler("index.tmpl", indexTitle))
	r.HandleFunc("/about", createStaticHandler("about.tmpl", aboutTitle))
}

func authRoutes(r *mux.Router, secure bool) {
	r.HandleFunc("/login", createFormHandler(createHandleLogin(secure), "login.tmpl", loginTitle))
	r.HandleFunc("/signup", createFormHandler(handleSignup, "signup.tmpl", signupTitle))
}
