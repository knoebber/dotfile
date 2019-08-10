package cli

import (
	"fmt"

	"github.com/knoebber/dotfile/file"
	"gopkg.in/alecthomas/kingpin.v2"
)

type pushCommand struct {
	data     *file.Data
	fileName string
}

func (pc *pushCommand) run(ctx *kingpin.ParseContext) error {
	fmt.Printf("TODO: Push %#v", pc.fileName)
	return nil
}

func addPushSubCommandToApplication(app *kingpin.Application, data *file.Data) {
	pc := &pushCommand{
		data: data,
	}
	p := app.Command("push", "push committed changes to central service").Action(pc.run)
	p.Arg("file-name", "the file to push").Required().StringVar(&pc.fileName)
}
