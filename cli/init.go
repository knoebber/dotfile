package main

import (
	"fmt"

	"github.com/knoebber/dotfile/file"
	"github.com/knoebber/dotfile/local"
	"gopkg.in/alecthomas/kingpin.v2"
)

type initCommand struct {
	getStorage func() (*local.Storage, error)
	fileName   string
	alias      string
}

func (ic *initCommand) run(ctx *kingpin.ParseContext) error {
	s, err := ic.getStorage()
	if err != nil {
		return err
	}

	relativePath, err := local.RelativePath(ic.fileName, s.Home)
	if err != nil {
		return err
	}

	alias, err := file.Init(s, relativePath, ic.alias)
	if err != nil {
		return err
	}

	fmt.Printf("Initialized %s as %#v\n", ic.fileName, alias)
	return nil
}

func addInitSubCommandToApplication(app *kingpin.Application, gs func() (*local.Storage, error)) {
	ic := &initCommand{
		getStorage: gs,
	}
	p := app.Command("init", "begin tracking a file").Action(ic.run)
	p.Arg("file-name", "the file to track").Required().StringVar(&ic.fileName)
	p.Arg("alias", "optional friendly name").StringVar(&ic.alias)
}
