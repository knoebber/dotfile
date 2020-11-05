package main

import (
	"fmt"
	"os"

	"github.com/knoebber/dotfile/cli"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	app := kingpin.New("dotfile", "version control optimized for single files")
	if err := cli.AddCommandsToApplication(app); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
	}
	kingpin.MustParse(app.Parse(os.Args[1:]))
}
