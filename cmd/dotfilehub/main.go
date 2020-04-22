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

func getConfig() server.Config {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	defaultDBName := filepath.Join(home, defaultDB)

	addr := flag.String("addr", defaultAddress, "HTTP address to listen on")
	dbPath := flag.String("db", defaultDBName, "Name of sqlite database file")
	secure := flag.Bool("secure", false, "Set session cookie to HTTPS only")
	proxyHeaders := flag.Bool("proxyheaders", false, "Set request IP by inspecting reverse proxy headers")
	flag.Parse()

	return server.Config{
		Addr:         *addr,
		DBPath:       *dbPath,
		Secure:       *secure,
		ProxyHeaders: *proxyHeaders,
	}
}

func main() {
	server.Start(getConfig())
}
