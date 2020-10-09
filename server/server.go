// Package server serves a web interface for interacting with dotfiles stored in a database.
package server

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/knoebber/dotfile/db"
	"github.com/pkg/errors"
)

// Config configures the server.
type Config struct {
	Addr           string      // Address to listen at.
	DBPath         string      // The path to store the sqlite database file.
	Secure         bool        // Tell the server code that the host is using https.
	ProxyHeaders   bool        // Sets request IP from reverse proxy headers.
	Host           string      // Overrides http.Request.Host when not empty.
	SMTP           *SMTPConfig // Sets up a SMTP Client
	SMTPConfigPath string      // Sets SMTP from this file's JSON when not empty.
}

// URL returns the configured url.
// If c.Host is not set it will use the requests host header.
func (c Config) URL(r *http.Request) string {
	protocol := "http://"
	if c.Secure {
		protocol = "https://"
	}

	if c.Host == "" {
		return protocol + r.Host
	}

	return protocol + c.Host
}

const timeout = 10 * time.Second

// Start starts the dotfile web server.
// Expects an assets folder in the same directory from where the binary is ran.
func Start(config Config) error {
	var err error

	if err = db.Start(config.DBPath); err != nil {
		return errors.Wrapf(err, "starting database")
	}
	defer db.Close()

	if config.SMTPConfigPath != "" {
		config.SMTP, err = smtpConfig(config.SMTPConfigPath)
		if err != nil {
			return err
		}
	}

	r := mux.NewRouter()

	if err := setupRoutes(r, config); err != nil {
		return err
	}

	s := &http.Server{
		Addr:         config.Addr,
		WriteTimeout: timeout,
		ReadTimeout:  timeout,
	}

	if config.ProxyHeaders {
		s.Handler = handlers.LoggingHandler(os.Stdout, handlers.ProxyHeaders(r))
	} else {
		s.Handler = handlers.LoggingHandler(os.Stdout, r)
	}

	if err := loadTemplates(); err != nil {
		return err
	}

	log.Printf("using sqlite3 database %s", config.DBPath)
	log.Println("serving dotfiles at", config.Addr)

	log.Panicf("dotfilehub listen and serve: %v", s.ListenAndServe())
	return nil
}
