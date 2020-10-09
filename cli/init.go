package cli

import (
	"fmt"

	"github.com/knoebber/dotfile/dotfile"
	"github.com/knoebber/dotfile/local"
	"gopkg.in/alecthomas/kingpin.v2"
)

type initCommand struct {
	path  string
	alias string
}

func (ic *initCommand) run(*kingpin.ParseContext) error {
	alias, err := dotfile.Alias(ic.alias, ic.path)
	if err != nil {
		return err
	}

	storage := &local.Storage{Dir: flags.storageDir, Alias: alias}
	if err = storage.InitFile(ic.path); err != nil {
		return err
	}

	fmt.Printf("Initialized as %q\n", alias)
	return nil
}

func addInitSubCommandToApplication(app *kingpin.Application) {
	ic := new(initCommand)

	p := app.Command("init", "begin tracking a file").Action(ic.run)
	p.Arg("path", "the file to track").Required().ExistingFileVar(&ic.path)
	p.Arg("alias", "optional friendly name").StringVar(&ic.alias)
}
