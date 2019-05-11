package commands

import (
	"fmt"

	"gopkg.in/alecthomas/kingpin.v2"
)

type diffCommand struct {
	fileName string
}

func (d *diffCommand) run(ctx *kingpin.ParseContext) error {
	fmt.Printf("TODO: Diff %#v\n", d.fileName)
	return nil
}

func addDiffSubCommandToApplication(app *kingpin.Application) {
	dc := &diffCommand{}
	c := app.Command("diff", "check changes to tracked file").Action(dc.run)
	c.Arg("file-name", "file to check changes in").Required().StringVar(&dc.fileName)
}
