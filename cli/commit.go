package cli

import (
	"github.com/knoebber/dotfile/file"
	"github.com/knoebber/dotfile/local"
	"gopkg.in/alecthomas/kingpin.v2"
)

type commitCommand struct {
	getStorage    func() (*local.Storage, error)
	fileName      string
	commitMessage string
}

func (c *commitCommand) run(ctx *kingpin.ParseContext) error {
	s, err := c.getStorage()
	if err != nil {
		return err
	}

	if err := file.NewCommit(s, c.fileName, c.commitMessage); err != nil {
		return err
	}

	return nil
}

func addCommitSubCommandToApplication(app *kingpin.Application, gs func() (*local.Storage, error)) {
	cc := &commitCommand{
		getStorage: gs,
	}
	c := app.Command("commit", "commit file to working tree").Action(cc.run)
	c.Arg("file-name", "name of file to save new revision of").Required().StringVar(&cc.fileName)
	c.Arg("commit-message",
		"a memo to remind yourself what's in this version").
		StringVar(&cc.commitMessage)
}
