package cli

import (
	"github.com/knoebber/dotfile/file"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

type initCommand struct {
	getStorage func() (*file.Storage, error)
	fileName   string
	altName    string
}

func (ic *initCommand) run(ctx *kingpin.ParseContext) error {
	s, err := ic.getStorage()
	if err != nil {
		return err
	}

	if err := file.Init(s, ic.fileName, ic.altName); err != nil {
		return errors.Wrapf(err, "failed to initialize %#v", ic.fileName)
	}

	return nil
}

func addInitSubCommandToApplication(app *kingpin.Application, gs func() (*file.Storage, error)) {
	ic := &initCommand{
		getStorage: gs,
	}
	p := app.Command("init", "begin tracking a file").Action(ic.run)
	p.Arg("file-name", "the file to track").Required().StringVar(&ic.fileName)
	p.Arg("alt-name", "optional friendly name").StringVar(&ic.altName)
}
