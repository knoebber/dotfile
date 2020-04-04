package main

import (
	"fmt"
	"os"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	app := kingpin.New("dotfile", "version control optimized for single files")
	if err := addCommandsToApplication(app); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	kingpin.MustParse(app.Parse(os.Args[1:]))
}
