package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/dotfile/server"
)

const defaultAddress = ":3001"

func parseFlags() (string, string) {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	defaultDBName := filepath.Join(home, ".dotfilehub.db")

	addr := flag.String("addr", defaultAddress, "HTTP address to listen on")
	dbPath := flag.String("db", defaultDBName, "Name of sqlite database file")
	flag.Parse()

	return *addr, *dbPath
}

func main() {
	addr, dbPath := parseFlags()

	if err := db.Start(dbPath); err != nil {
		log.Panicf("starting database connection: %v", err)
	}
	defer db.Close()

	server.Start(addr)
}
