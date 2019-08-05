package main

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/knoebber/dotfile/cli"
)

func main() {
	app := kingpin.New("dotfile", "version control optimized for single files")
	cli.AddCommandsToApplication(app)
	kingpin.MustParse(app.Parse(os.Args[1:]))
}
