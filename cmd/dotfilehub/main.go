package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/knoebber/dotfile/server"
)

const (
	defaultAddress = ":3000"
	defaultDB      = ".dotfilehub.db"
)

func config() server.Config {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	defaultDBName := filepath.Join(home, defaultDB)

	addr := flag.String("addr", defaultAddress, "HTTP address to listen on")
	dbPath := flag.String("db", defaultDBName, "Name of sqlite database file")
	host := flag.String("host", "", "Override the host header for remote name display")
	secure := flag.Bool("secure", false, "Set session cookie to HTTPS only")
	proxyHeaders := flag.Bool("proxyheaders", false, "Set request IP by inspecting reverse proxy headers")
	smtpConfigPath := flag.String("smtp-config-path", "", "Sets up a SMTP client for account recovery.")
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
	server.Start(config())
}
