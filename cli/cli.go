package main

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/knoebber/dotfile/cli/commands"
)

var (
	verbose = kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()
	name    = kingpin.Arg("name", "Name of user.").Required().String()
)

func main() {
	app := kingpin.New("dotfile", "Better version control for configuration files")
	commands.AddCommandsToApplication(app)
	kingpin.MustParse(app.Parse(os.Args[1:]))
}
