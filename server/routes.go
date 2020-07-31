package server

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/knoebber/dotfile/db"
)

func setupRoutes(r *mux.Router, config Config) {
	staticRoutes(r, config)
	assetRoutes(r)
	apiRoutes(r)
	// Important to register these last so non dynamic routes take precedence.
	dynamicRoutes(r)

	createReservedUsernames(r)
}

// Pages that do not have a path variables.
func staticRoutes(r *mux.Router, config Config) {
	r.HandleFunc("/", indexHandler())
	r.HandleFunc("/about", aboutHandler())
	r.HandleFunc("/acknowledgments", acknowledgmentsHander())
	r.HandleFunc("/signup", signupHandler(config.Secure))
	r.HandleFunc("/login", loginHandler(config.Secure))
	r.HandleFunc("/logout", logoutHandler())
	r.HandleFunc("/new_file", newFileHandler())
	r.HandleFunc("/settings", settingsHandler())
	r.HandleFunc("/settings/email", emailHandler())
	r.HandleFunc("/settings/timezone", timezoneHandler())
	r.HandleFunc("/settings/password", passwordHandler())
	r.HandleFunc("/settings/theme", themeHandler())
	r.HandleFunc("/settings/cli", cliHandler(config))
}

func assetRoutes(r *mux.Router) {
	assets := http.FileSystem(http.Dir("assets/"))
	r.Path("/style.css").Handler(http.FileServer(assets))
	r.Path("/favicon.ico").Handler(http.FileServer(assets))
}

func apiRoutes(r *mux.Router) {
	r.HandleFunc("/api/{username}/{alias}", handleFileJSON).Methods("GET")
	r.HandleFunc("/api/{username}/{alias}", handlePush).Methods("POST")
	r.HandleFunc("/api/{username}/{alias}/{hash}", handleRawCompressedCommit)
}

func dynamicRoutes(r *mux.Router) {
	r.HandleFunc("/temp_file/{alias}/create", confirmNewFileHandler())
	r.HandleFunc("/temp_file/{alias}/commit", confirmEditHandler())
	r.HandleFunc("/{username}", userHandler())
	r.HandleFunc("/{username}/{alias}", fileHandler())
	r.HandleFunc("/{username}/{alias}/raw", handleRawFile)
	r.HandleFunc("/{username}/{alias}/commits", commitsHandler())
	r.HandleFunc("/{username}/{alias}/edit", editFileHandler())
	r.HandleFunc("/{username}/{alias}/diff", diffHandler())
	r.HandleFunc("/{username}/{alias}/{hash}", commitHandler())
	r.HandleFunc("/{username}/{alias}/{hash}/raw", handleRawUncompressedCommit)
}

// Prevent users for registering any username that conflicts with an existing route.
// For example a user named "about" wouldn't be able to see their files.
func createReservedUsernames(r *mux.Router) {
	reserved := []interface{}{}
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
		log.Fatalf("walking routes: %s", err)
	}

	if err := db.SeedReservedUsernames(reserved); err != nil {
		log.Fatal(err)
	}
}
