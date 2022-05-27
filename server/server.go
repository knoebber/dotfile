// Package server serves a web interface for interacting with dotfiles stored in a database.
package server

import (
	"embed"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/knoebber/dotfile/db"
	"github.com/pkg/errors"
)

const timeout = 10 * time.Second

//go:embed assets
//go:embed html
//go:embed templates/base.tmpl templates/auth/* templates/file/* templates/user/*
var serverContent embed.FS

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

// New returns a dotfilehub web server.
func New(config Config) (*http.Server, error) {
	var err error

	if err = db.Start(config.DBPath); err != nil {
		return nil, errors.Wrapf(err, "starting database")
	}

	if config.SMTPConfigPath != "" {
		config.SMTP, err = smtpConfig(config.SMTPConfigPath)
		if err != nil {
			return nil, err
		}
	}

	r := mux.NewRouter()

	if err := setupRoutes(r, config); err != nil {
		return nil, err
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
		return nil, err
	}

	return s, nil
}
