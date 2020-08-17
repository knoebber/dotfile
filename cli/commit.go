package cli

import (
	"github.com/knoebber/dotfile/file"
	"gopkg.in/alecthomas/kingpin.v2"
)

type commitCommand struct {
	alias         string
	commitMessage string
}

func (c *commitCommand) run(ctx *kingpin.ParseContext) error {
	s, err := loadFile(c.alias)
	if err != nil {
		return err
	}

	if err := file.NewCommit(s, c.commitMessage); err != nil {
		return err
	}

	return nil
}

func addCommitSubCommandToApplication(app *kingpin.Application) {
	cc := new(commitCommand)

	c := app.Command("commit", "save a revision of file").Action(cc.run)
	c.Arg("alias", "name of file to save new revision of").Required().StringVar(&cc.alias)
	c.Arg("commit-message",
		"a memo to remind yourself what's in this version").
		StringVar(&cc.commitMessage)
}
