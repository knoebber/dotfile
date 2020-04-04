package main

import (
	"flag"
	"net/http"

	"log"
)

const defaultAddr = ":3001"

func getAddress() string {
	addr := flag.String("addr", defaultAddr, "HTTP address to listen on")
	flag.Parse()

	return *addr
}

func main() {
	startServer(getAddress())
}

func startServer(addr string) {
	fs := http.FileServer(http.Dir("assets/"))
	http.Handle("/", fs)

	log.Println("serving dotfiles at", addr)
	log.Panicf("starting dotfile server: %v", http.ListenAndServe(addr, nil))
}
