package cli

import (
	"github.com/knoebber/dotfile/file"
	"gopkg.in/alecthomas/kingpin.v2"
)

type checkoutCommand struct {
	fileName   string
	commitHash string
}

func (c *checkoutCommand) run(ctx *kingpin.ParseContext) error {
	s, err := loadFile(c.fileName)
	if err != nil {
		return err
	}
	if c.commitHash == "" {
		c.commitHash = s.Tracking.Revision
	}

	if err := file.Checkout(s, c.commitHash); err != nil {
		return err
	}

	return nil
}

func addCheckoutSubCommandToApplication(app *kingpin.Application) {
	cc := new(checkoutCommand)

	c := app.Command("checkout", "revert a file to a previously committed state").Action(cc.run)
	c.Arg("file-name", "name of file to revert changes in").Required().StringVar(&cc.fileName)
	c.Arg("commit-hash", "the revision to revert to").StringVar(&cc.commitHash)
}
