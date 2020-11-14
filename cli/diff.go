package cli

import (
	"fmt"
	"github.com/knoebber/dotfile/dotfile"
	"gopkg.in/alecthomas/kingpin.v2"
)

// TODO commitHash should match on first 7 characters as well.
// Same for checkout
type diffCommand struct {
	alias      string
	commitHash string
}

func (d *diffCommand) run(*kingpin.ParseContext) error {
	s, err := loadFile(d.alias)
	if err != nil {
		return err
	}

	if d.commitHash == "" {
		d.commitHash = s.FileData.Revision
	}

	diff, err := dotfile.DiffPrettyText(s, d.commitHash, "")
	if err != nil {
		return err
	}

	fmt.Println(diff)
	return nil
}

func addDiffSubCommandToApplication(app *kingpin.Application) {
	dc := new(diffCommand)
	c := app.Command("diff", "check changes to tracked file").Action(dc.run)
	c.Arg("alias", "file to check changes in").
		HintAction(flags.defaultAliasList).
		Required().
		StringVar(&dc.alias)
	c.Arg("commit-hash",
		"the revision to diff against; default current").
		StringVar(&dc.commitHash)

}
