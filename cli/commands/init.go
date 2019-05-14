package commands

import (
	"github.com/knoebber/dotfile"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

type initCommand struct {
	fileName string
	altName  string
}

func (ic *initCommand) run(ctx *kingpin.ParseContext) error {
	if err := dotfile.Init(ic.fileName, ic.altName); err != nil {
		return errors.Wrapf(err, "error initializing: %#v", ic.fileName)
	}
	return nil
}

func addInitSubCommandToApplication(app *kingpin.Application) {
	ic := &initCommand{}
	p := app.Command("init", "begin tracking a file").Action(ic.run)
	p.Arg("file-name", "the file to track").Required().StringVar(&ic.fileName)
	p.Arg("alt-name", "optional friendly name").StringVar(&ic.altName)
}
