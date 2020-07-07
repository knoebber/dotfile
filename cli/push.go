package cli

import (
	"github.com/knoebber/dotfile/local"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

type pushCommand struct {
	fileName string
}

func (pc *pushCommand) run(ctx *kingpin.ParseContext) error {
	storage, err := local.NewStorage(config.home, config.storageDir)
	if err != nil {
		return errors.Wrap(err, "getting storage")
	}

	local.Push(storage, config.user, pc.fileName)
	return nil
}

func addPushSubCommandToApplication(app *kingpin.Application) {
	pc := new(pushCommand)

	p := app.Command("push", "push committed changes to central service").Action(pc.run)
	p.Arg("file-name", "the file to push").Required().StringVar(&pc.fileName)
}
