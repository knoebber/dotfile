package commands

import (
	"github.com/knoebber/dotfile"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
	"time"
)

const defaultMessageTimestampDisplayFormat = "January 02, 2006 3:04 PM -0700"

type commitCommand struct {
	fileName      string
	commitMessage string
}

func (c *commitCommand) run(ctx *kingpin.ParseContext) error {
	if err := dotfile.Commit(c.fileName, c.commitMessage); err != nil {
		return errors.Wrapf(err, "failed to commit %#v", c.fileName)
	}
	return nil
}

func addCommitSubCommandToApplication(app *kingpin.Application) {
	cc := &commitCommand{}
	c := app.Command("commit", "commit file to working tree").Action(cc.run)
	c.Arg("file-name", "the file to track").Required().StringVar(&cc.fileName)
	c.Arg("commit-message",
		"a memo to remind yourself what's in this version; defaults to local timestamp").
		Default(time.Now().Format(defaultMessageTimestampDisplayFormat)).
		StringVar(&cc.commitMessage)
}
