package commands

import (
	"fmt"

	"gopkg.in/alecthomas/kingpin.v2"
)

type initCommand struct {
	fileName string
	altName  string
}

func (ic *initCommand) run(ctx *kingpin.ParseContext) error {
	fmt.Printf("TODO: Init %#v (altName: %#v)\n", ic.fileName, ic.altName)
	return nil
}

func addInitSubCommandToApplication(app *kingpin.Application) {
	ic := &initCommand{}
	p := app.Command("init", "begin tracking a file").Action(ic.run)
	p.Arg("file-name", "the file to track").Required().StringVar(&ic.fileName)
	p.Arg("alt-name", "optional friendly name").StringVar(&ic.altName)
}
