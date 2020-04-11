package server

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const timeout = 10 * time.Second

// Start starts the dotfile web server.
// Expects an assets folder in the same directory from where the binary is ran.
func Start(addr string) {
	r := mux.NewRouter()
	log.Println("serving dotfiles at", addr)

	r.Use(handlers.ProxyHeaders)
	setupRoutes(r)

	s := &http.Server{
		Handler:      handlers.LoggingHandler(os.Stdout, r),
		Addr:         addr,
		WriteTimeout: timeout,
		ReadTimeout:  timeout,
	}

	if err := loadTemplates(); err != nil {
		log.Panic(err)
	}

	log.Panicf("starting dotfile server: %v", s.ListenAndServe())
}
