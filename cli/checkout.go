package main

import (
	"github.com/knoebber/dotfile/file"
	"github.com/knoebber/dotfile/local"
	"gopkg.in/alecthomas/kingpin.v2"
)

type checkoutCommand struct {
	getStorage func() (*local.Storage, error)
	fileName   string
	commitHash string
}

func (c *checkoutCommand) run(ctx *kingpin.ParseContext) error {
	s, err := c.getStorage()
	if err != nil {
		return err
	}

	if err := file.Checkout(s, c.fileName, c.commitHash); err != nil {
		return err
	}

	return nil
}

func addCheckoutSubCommandToApplication(app *kingpin.Application, gs func() (*local.Storage, error)) {
	cc := &checkoutCommand{
		getStorage: gs,
	}
	c := app.Command("checkout", "revert a file to a previously committed state").Action(cc.run)
	c.Arg("file-name", "name of file to revert changes in").Required().StringVar(&cc.fileName)
	c.Arg("commit-hash", "the revision to revert to").StringVar(&cc.commitHash)
}
