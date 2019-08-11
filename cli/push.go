package cli

import (
	"fmt"

	"github.com/knoebber/dotfile/file"
	"gopkg.in/alecthomas/kingpin.v2"
)

type pushCommand struct {
	storage  *file.Storage
	fileName string
}

func (pc *pushCommand) run(ctx *kingpin.ParseContext) error {
	fmt.Printf("TODO: Push %#v", pc.fileName)
	return nil
}

func addPushSubCommandToApplication(app *kingpin.Application, storage *file.Storage) {
	pc := &pushCommand{
		storage: storage,
	}
	p := app.Command("push", "push committed changes to central service").Action(pc.run)
	p.Arg("file-name", "the file to push").Required().StringVar(&pc.fileName)
}
