package cli

import (
	"time"

	"github.com/knoebber/dotfile/file"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

const defaultMessageTimestampDisplayFormat = "January 02, 2006 3:04 PM -0700"

type commitCommand struct {
	data          *file.Data
	fileName      string
	commitMessage string
}

func (c *commitCommand) run(ctx *kingpin.ParseContext) error {
	_, err := file.Commit(c.data, c.fileName, c.commitMessage)
	if err != nil {
		return errors.Wrapf(err, "failed to commit %#v", c.fileName)
	}

	return nil
}

func addCommitSubCommandToApplication(app *kingpin.Application, data *file.Data) {
	cc := &commitCommand{
		data: data,
	}
	c := app.Command("commit", "commit file to working tree").Action(cc.run)
	c.Arg("file-name", "the file to track").Required().StringVar(&cc.fileName)
	c.Arg("commit-message",
		"a memo to remind yourself what's in this version; defaults to local timestamp").
		Default(time.Now().Format(defaultMessageTimestampDisplayFormat)).
		StringVar(&cc.commitMessage)
}
