package cli

import (
	"fmt"

	"github.com/knoebber/dotfile/local"
	"gopkg.in/alecthomas/kingpin.v2"
)

type diffCommand struct {
	getStorage func() (*local.Storage, error)
	fileName   string
}

func (d *diffCommand) run(ctx *kingpin.ParseContext) error {
	_, err := d.getStorage()
	if err != nil {
		return err
	}

	fmt.Printf("TODO: Diff %#v\n", d.fileName)
	return nil
}

func addDiffSubCommandToApplication(app *kingpin.Application, gs func() (*local.Storage, error)) {
	dc := &diffCommand{
		getStorage: gs,
	}
	c := app.Command("diff", "check changes to tracked file").Action(dc.run)
	c.Arg("file-name", "file to check changes in").Required().StringVar(&dc.fileName)
}
