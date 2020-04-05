package server

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const timeoutSeconds = 10

// Start starts the dotfile web server.
// Expects an assets folder in the same directory from where its ran.
func Start(addr string) {
	r := mux.NewRouter()
	log.Println("serving dotfiles at", addr)

	setupRoutes(r)

	s := &http.Server{
		Handler:      handlers.LoggingHandler(os.Stdout, r),
		Addr:         addr,
		WriteTimeout: timeoutSeconds * time.Second,
		ReadTimeout:  timeoutSeconds * time.Second,
	}

	log.Panicf("starting dotfile server: %v", s.ListenAndServe())
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
