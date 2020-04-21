package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

func setupRoutes(r *mux.Router, secure bool) {
	assetRoutes(r)
	staticRoutes(r)
	userRoutes(r, secure)
}

func assetRoutes(r *mux.Router) {
	assets := http.FileSystem(http.Dir("assets/"))
	r.Path("/style.css").Handler(http.FileServer(assets))
	r.Path("/favicon.ico").Handler(http.FileServer(assets))
}

func staticRoutes(r *mux.Router) {
	r.HandleFunc("/", getIndexHandler())
	r.HandleFunc("/about", getAboutHandler())
}

func userRoutes(r *mux.Router, secure bool) {
	r.HandleFunc("/login", getLoginHandler(secure))
	r.HandleFunc("/signup", getSignupHandler())
	r.HandleFunc("/logout", getLogoutHandler())

	sub := r.PathPrefix("/{username}").Subrouter()
	sub.HandleFunc("", getUserHandler())
	sub.HandleFunc("/email", getEmailHandler())
	sub.HandleFunc("/password", getPasswordHandler())
}
