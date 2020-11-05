package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/knoebber/dotfile/server"
)

const (
	defaultAddress = ":3000"
	defaultDBName  = ".dotfilehub.db"
)

func config() server.Config {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	defaultDBPath := filepath.Join(home, defaultDBName)

	addr := flag.String("addr", defaultAddress, "HTTP address to listen on")
	dbPath := flag.String("db", defaultDBPath, "Path to sqlite3 database file")
	host := flag.String("host", "", "Override the host header for remote name display")
	secure := flag.Bool("secure", false, "Set session cookie to HTTPS only")
	proxyHeaders := flag.Bool("proxyheaders", false, "Set request IP by inspecting reverse proxy headers")
	smtpConfigPath := flag.String("smtp-config-path", "", "Sets up a SMTP client for account recovery")
	flag.Parse()

	return server.Config{
		Addr:           *addr,
		DBPath:         *dbPath,
		Secure:         *secure,
		ProxyHeaders:   *proxyHeaders,
		Host:           *host,
		SMTPConfigPath: *smtpConfigPath,
	}
}

func main() {
	c := config()

	s, err := server.New(c)
	if err != nil {
		fmt.Println("failed to start dotfilehub server:", err)
		os.Exit(1)
	}

	log.Printf("using sqlite3 database %s", c.DBPath)
	log.Println("serving dotfiles at", c.Addr)

	log.Panicf("dotfilehub listen and serve: %v", s.ListenAndServe())
}
