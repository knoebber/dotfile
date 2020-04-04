package cli

import (
	"fmt"

	"github.com/knoebber/dotfile/local"
	"gopkg.in/alecthomas/kingpin.v2"
)

type pushCommand struct {
	getStorage func() (*local.Storage, error)
	fileName   string
}

func (pc *pushCommand) run(ctx *kingpin.ParseContext) error {
	_, err := pc.getStorage()
	if err != nil {
		return err
	}

	fmt.Printf("TODO: Push %#v", pc.fileName)
	return nil
}

func addPushSubCommandToApplication(app *kingpin.Application, gs func() (*local.Storage, error)) {
	pc := &pushCommand{
		getStorage: gs,
	}
	p := app.Command("push", "push committed changes to central service").Action(pc.run)
	p.Arg("file-name", "the file to push").Required().StringVar(&pc.fileName)
}
