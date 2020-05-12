package cli

import (
	"fmt"

	"github.com/knoebber/dotfile/local"
	"gopkg.in/alecthomas/kingpin.v2"
)

type initCommand struct {
	path  string
	alias string
}

func (ic *initCommand) run(ctx *kingpin.ParseContext) error {
	alias, err := local.InitFile(config.home, config.storageDir, ic.path, ic.alias)
	if err != nil {
		return err
	}

	fmt.Printf("Initialized %s as %#v\n", ic.path, alias)
	return nil
}

func addInitSubCommandToApplication(app *kingpin.Application) {
	ic := new(initCommand)

	p := app.Command("init", "begin tracking a file").Action(ic.run)
	p.Arg("path", "the file to track").Required().ExistingFileVar(&ic.path)
	p.Arg("alias", "optional friendly name").StringVar(&ic.alias)
}
