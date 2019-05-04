package commands

import (
	"fmt"

	"gopkg.in/alecthomas/kingpin.v2"
)

type pushCommand struct {
	fileName string
}

func (pc *pushCommand) run(ctx *kingpin.ParseContext) error {
	fmt.Printf("TODO: Push %#v", pc.fileName)
	return nil
}

func addPushSubCommandToApplication(app *kingpin.Application) {
	pc := &pushCommand{}
	p := app.Command("push", "push committed changes to central service").Action(pc.run)
	p.Arg("file-name", "the file to push").Required().StringVar(&pc.fileName)
}
