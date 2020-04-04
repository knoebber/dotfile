package main

import (
	"flag"

	"github.com/knoebber/dotfile/server"
)

const defaultAddress = ":3001"

func getAddress() string {
	addr := flag.String("addr", defaultAddress, "HTTP address to listen on")
	flag.Parse()

	return *addr
}

func main() {
	server.Start(getAddress())
}
