package cli

import (
	"fmt"

	"github.com/knoebber/dotfile/file"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

type initCommand struct {
	storage  *file.Storage
	fileName string
	altName  string
}

func (ic *initCommand) run(ctx *kingpin.ParseContext) error {
	alias, err := file.Init(ic.storage, ic.fileName, ic.altName)
	if err != nil {
		return errors.Wrapf(err, "failed to initialize %#v", ic.fileName)
	}

	fmt.Printf("Initialized %s as %#v\n", ic.fileName, alias)
	return nil
}

func addInitSubCommandToApplication(app *kingpin.Application, storage *file.Storage) {
	ic := &initCommand{
		storage: storage,
	}
	p := app.Command("init", "begin tracking a file").Action(ic.run)
	p.Arg("file-name", "the file to track").Required().StringVar(&ic.fileName)
	p.Arg("alt-name", "optional friendly name").StringVar(&ic.altName)
}
