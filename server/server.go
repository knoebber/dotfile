package server

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/knoebber/dotfile/db"
)

// Config configures the server.
type Config struct {
	Addr         string // Address to listen at.
	DBPath       string // The path to store the sqlite database file.
	Secure       bool   // Sets session cookie to HTTPS only.
	ProxyHeaders bool   // Sets request IP from reverse proxy headers.
}

const timeout = 10 * time.Second

// Start starts the dotfile web server.
// Expects an assets folder in the same directory from where the binary is ran.
func Start(cfg Config) {
	if err := db.Start(cfg.DBPath); err != nil {
		log.Panicf("starting database connection: %v", err)
	}
	defer db.Close()

	r := mux.NewRouter()

	setupRoutes(r, cfg.Secure)

	s := &http.Server{
		Addr:         cfg.Addr,
		WriteTimeout: timeout,
		ReadTimeout:  timeout,
	}

	if cfg.ProxyHeaders {
		s.Handler = handlers.LoggingHandler(os.Stdout, handlers.ProxyHeaders(r))
	} else {
		s.Handler = handlers.LoggingHandler(os.Stdout, r)
	}

	if err := loadTemplates(); err != nil {
		log.Panic(err)
	}

	log.Printf("using sqlite3 database %s", cfg.DBPath)
	log.Println("serving dotfiles at", cfg.Addr)

	log.Panicf("starting dotfile server: %v", s.ListenAndServe())
}
