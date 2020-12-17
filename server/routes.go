package server

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/knoebber/dotfile/db"
	"github.com/pkg/errors"
)

func setupRoutes(r *mux.Router, config Config) error {
	staticRoutes(r)
	assetRoutes(r)
	apiRoutes(r)
	dotfileRoutes(r, config)
	return createReservedUsernames(r)
}

// Pages that get their content from the html/ directory.
func staticRoutes(r *mux.Router) {
	r.HandleFunc("/README.org", createStaticHandler(aboutTitle, "README.html"))
	r.HandleFunc("/terms", createStaticHandler("Terms of Use", "terms.html"))
	r.HandleFunc("/docs/cli.org", createStaticHandler("CLI Documentation", "cli.html"))
	r.HandleFunc("/docs/web.org", createStaticHandler("Web Documentation", "web.html"))
	r.HandleFunc("/docs/acknowledgments.org", createStaticHandler("Acknowledgments", "acknowledgments.html"))
	r.NotFoundHandler = createStaticHandler("Not Found", "404.html")
}

func assetRoutes(r *mux.Router) {
	assets := http.FileSystem(http.Dir("assets/"))
	r.Path("/style.css").Handler(http.FileServer(assets))
	r.Path("/favicon.ico").Handler(http.FileServer(assets))
	r.Path("/robots.txt").Handler(http.FileServer(assets))
}

func apiRoutes(r *mux.Router) {
	r.HandleFunc("/api/v1/user/{username}", handleFileListJSON)
	r.HandleFunc("/api/v1/user/{username}/{alias}", handleFileJSON).Methods("GET")
	r.HandleFunc("/api/v1/user/{username}/{alias}", handlePush).Methods("POST")
	r.HandleFunc("/api/v1/user/{username}/{alias}/raw", handleRawFile)
	r.HandleFunc("/api/v1/user/{username}/{alias}/{hash}", handleRawCompressedCommit)
}

func dotfileRoutes(r *mux.Router, config Config) {
	r.HandleFunc("/", indexHandler())
	r.HandleFunc("/feed.rss", createRSSFeed(config))
	r.HandleFunc("/signup", signupHandler(config.Secure))
	r.HandleFunc("/login", loginHandler(config.Secure))
	r.HandleFunc("/account_recovery", accountRecoveryHandler(config))
	r.HandleFunc("/reset_password", resetPasswordHandler())
	r.HandleFunc("/logout", logoutHandler())
	r.HandleFunc("/new_file", newFileHandler())
	r.HandleFunc("/settings", settingsHandler())
	r.HandleFunc("/settings/email", emailHandler())
	r.HandleFunc("/settings/timezone", timezoneHandler())
	r.HandleFunc("/settings/password", passwordHandler())
	r.HandleFunc("/settings/theme", themeHandler())
	r.HandleFunc("/settings/cli", cliHandler(config))
	r.HandleFunc("/settings/delete", deleteUserHandler())
	r.HandleFunc("/{username}", userHandler())
	r.HandleFunc("/{username}/{alias}", fileHandler())
	r.HandleFunc("/{username}/{alias}/raw", handleRawFile)
	r.HandleFunc("/{username}/{alias}/commits", commitsHandler())
	r.HandleFunc("/{username}/{alias}/edit", editFileHandler())
	r.HandleFunc("/{username}/{alias}/diff", diffHandler())
	r.HandleFunc("/{username}/{alias}/init", confirmNewFileHandler())
	r.HandleFunc("/{username}/{alias}/commit", confirmEditHandler())
	r.HandleFunc("/{username}/{alias}/settings", fileSettingsHandler())
	r.HandleFunc("/{username}/{alias}/settings/update", updateFileHandler())
	r.HandleFunc("/{username}/{alias}/settings/delete", deleteFileHandler())
	r.HandleFunc("/{username}/{alias}/settings/clear", clearFileHandler())
	r.HandleFunc("/{username}/{alias}/{hash}", commitHandler())
	r.HandleFunc("/{username}/{alias}/{hash}/raw", handleRawUncompressedCommit)
}

// Prevent users for registering any username that conflicts with an existing route.
// For example a user named "login" wouldn't be able to see their files.
func createReservedUsernames(r *mux.Router) error {
	var reserved []interface{}
	err := r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err != nil {
			return err
		}

		split := strings.Split(pathTemplate, "/")
		if split[1] == "{username}" {
			return nil
		}

		reserved = append(reserved, split[1])
		return nil
	})

	if err != nil {
		return errors.Wrapf(err, "walking routes")
	}

	if err := db.SeedReservedUsernames(db.Connection, reserved); err != nil {
		return err
	}

	return nil
}
