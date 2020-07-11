package cli

import (
	"fmt"

	"github.com/knoebber/dotfile/file"
	"gopkg.in/alecthomas/kingpin.v2"
)

type initCommand struct {
	path  string
	alias string
}

func (ic *initCommand) run(ctx *kingpin.ParseContext) error {
	alias, err := file.GetAlias(ic.alias, ic.path)
	if err != nil {
		return err
	}

	s, err := loadFileStorage(alias)
	if err != nil {
		return err
	}

	if s.HasFile {
		return fmt.Errorf("%#v is already tracked", ic.alias)
	}

	if err = s.InitFile(ic.path); err != nil {
		return err
	}

	fmt.Printf("Initialized as %#v\n", alias)
	return nil
}

func addInitSubCommandToApplication(app *kingpin.Application) {
	ic := new(initCommand)

	p := app.Command("init", "begin tracking a file").Action(ic.run)
	p.Arg("path", "the file to track").Required().ExistingFileVar(&ic.path)
	p.Arg("alias", "optional friendly name").StringVar(&ic.alias)
}
