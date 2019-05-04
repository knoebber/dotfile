package commands

import (
	"fmt"

	"gopkg.in/alecthomas/kingpin.v2"
)

type pullCommand struct {
	fileName string
}

func (pc *pullCommand) run(ctx *kingpin.ParseContext) error {
	fmt.Printf("TODO: Pull %#v", pc.fileName)
	return nil
}

func addPullSubCommandToApplication(app *kingpin.Application) {
	pc := &pullCommand{}
	p := app.Command("pull", "push changes from central service").Action(pc.run)
	p.Arg("file-name", "the file to pull").Required().StringVar(&pc.fileName)
}
