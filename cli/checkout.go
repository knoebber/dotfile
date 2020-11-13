package cli

import (
	"fmt"
	"github.com/knoebber/dotfile/dotfile"
	"github.com/knoebber/dotfile/usererror"
	"gopkg.in/alecthomas/kingpin.v2"
)

type checkoutCommand struct {
	alias      string
	commitHash string
	force      bool
}

func (c *checkoutCommand) run(*kingpin.ParseContext) error {
	s, err := loadFile(c.alias)
	if err != nil {
		return err
	}
	if c.commitHash == "" {
		c.commitHash = s.FileData.Revision
	}

	if !c.force {
		clean, err := dotfile.IsClean(s, c.commitHash)
		if err != nil {
			return err
		}
		if !clean {
			return usererror.Invalid(fmt.Sprintf(`"%s" has uncommitted changes, use -f to override`, c.alias))
		}
	}

	if err := dotfile.Checkout(s, c.commitHash); err != nil {
		return err
	}

	return nil
}

func addCheckoutSubCommandToApplication(app *kingpin.Application) {
	cc := new(checkoutCommand)

	c := app.Command("checkout", "revert a file to a previously committed state").Action(cc.run)
	c.Arg("alias", "name of file to revert changes in").Required().StringVar(&cc.alias)
	c.Arg("commit-hash", "the revision to revert to").StringVar(&cc.commitHash)
	c.Flag("force", "revert a file with uncommitted changes").Short('f').BoolVar(&cc.force)
}
