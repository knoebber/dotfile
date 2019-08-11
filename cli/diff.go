package cli

import (
	"fmt"

	"github.com/knoebber/dotfile/file"
	"gopkg.in/alecthomas/kingpin.v2"
)

type diffCommand struct {
	storage  *file.Storage
	fileName string
}

func (d *diffCommand) run(ctx *kingpin.ParseContext) error {
	fmt.Printf("TODO: Diff %#v\n", d.fileName)
	return nil
}

func addDiffSubCommandToApplication(app *kingpin.Application, storage *file.Storage) {
	dc := &diffCommand{
		storage: storage,
	}
	c := app.Command("diff", "check changes to tracked file").Action(dc.run)
	c.Arg("file-name", "file to check changes in").Required().StringVar(&dc.fileName)
}
