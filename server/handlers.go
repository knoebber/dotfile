package server

import "net/http"

func handleIndex(w http.ResponseWriter, r *http.Request) {
	renderStatic(w, newStatic(indexTitle, "Welcome to dotfilehub"))
}

func handleAbout(w http.ResponseWriter, r *http.Request) {
	renderStatic(w, newStatic(aboutTitle, "About this website :)"))
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	renderStatic(w, newStatic(loginTitle, "Login.. um TODO ??"))
}
