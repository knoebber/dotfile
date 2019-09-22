package cli

import (
	"time"

	"github.com/knoebber/dotfile/file"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

const defaultMessageTimestampDisplayFormat = "January 02, 2006 3:04 PM -0700"

type commitCommand struct {
	getStorage    func() (*file.Storage, error)
	fileName      string
	commitMessage string
}

func (c *commitCommand) run(ctx *kingpin.ParseContext) error {
	s, err := c.getStorage()
	if err != nil {
		return err
	}

	if _, err := file.Commit(s, c.fileName, c.commitMessage); err != nil {
		return errors.Wrapf(err, "failed to commit %#v", c.fileName)
	}

	return nil
}

func addCommitSubCommandToApplication(app *kingpin.Application, gs func() (*file.Storage, error)) {
	cc := &commitCommand{
		getStorage: gs,
	}
	c := app.Command("commit", "commit file to working tree").Action(cc.run)
	c.Arg("file-name", "the file to track").Required().StringVar(&cc.fileName)
	c.Arg("commit-message",
		"a memo to remind yourself what's in this version; defaults to local timestamp").
		Default(time.Now().Format(defaultMessageTimestampDisplayFormat)).
		StringVar(&cc.commitMessage)
}
