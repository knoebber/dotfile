package cli

import (
	"fmt"
	"github.com/hexops/gotextdiff"
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

	to := d.commitHash
	if to == "" {
		to = "*"
	}
	fmt.Printf("\033[1mdiff %s %s\033[0m\n", s.FileData.Path, to)

	unified, err := dotfile.Diff(s, d.commitHash, "")
	if err != nil {
		return err
	}

	for _, hunk := range unified.Hunks {
		if len(unified.Hunks) > 1 {
			fmt.Println("\033[1m==HUNK==\033[0m")
		}
		for _, line := range hunk.Lines {
			text := line.Content

			switch line.Kind {
			case gotextdiff.Insert:
				fmt.Print("\x1b[32m")
				fmt.Print(text)
				fmt.Print("\x1b[0m")
			case gotextdiff.Delete:
				fmt.Print("\x1b[31m")
				fmt.Print(text)
				fmt.Print("\x1b[0m")
			case gotextdiff.Equal:
				fmt.Print(text)
			}
		}
	}

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
