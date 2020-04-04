package server

import (
	"log"
	"net/http"
)

// Start starts the dotfile web server.
// Expects an assets folder in the same directory from where its ran.
func Start(addr string) {

	fs := http.FileServer(http.Dir("assets/"))
	http.Handle("/", fs)

	log.Println("serving dotfiles at", addr)
	log.Panicf("starting dotfile server: %v", http.ListenAndServe(addr, nil))
}
