package commands

import (
	"fmt"

	"gopkg.in/alecthomas/kingpin.v2"
)

type checkoutCommand struct {
	fileName   string
	commitHash string
}

func (c *checkoutCommand) run(ctx *kingpin.ParseContext) error {
	fmt.Printf("TODO: Checkout %#v commitHash: %#v\n", c.fileName, c.commitHash)
	return nil
}

func addCheckoutSubCommandToApplication(app *kingpin.Application) {
	cc := &checkoutCommand{}
	c := app.Command("checkout", "revert a file to a previously committed state").Action(cc.run)
	c.Arg("file-name", "file to revert changes in").Required().StringVar(&cc.fileName)
	c.Arg("commit-hash", "the revision to revert to").StringVar(&cc.commitHash)
}
