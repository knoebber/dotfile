package commands

import (
	"fmt"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
)

const defaultMessageTimestampDisplayFormat = "January 02, 2006 3:04 PM -0700"

type commitCommand struct {
	fileName      string
	commitMessage string
}

func (c *commitCommand) run(ctx *kingpin.ParseContext) error {
	fmt.Printf("TODO: Commit %#v with message %#v\n", c.fileName, c.commitMessage)
	return nil
}

func addCommitSubCommandToApplication(app *kingpin.Application) {
	cc := &commitCommand{}
	c := app.Command("commit", "commit file to working tree").Action(cc.run)
	c.Arg("file-name", "the file to track").Required().StringVar(&cc.fileName)
	c.Arg("commit-message", "a memo to remind yourself what's in this version; defaults to local timestamp").Default(time.Now().Format(defaultMessageTimestampDisplayFormat)).StringVar(&cc.commitMessage)

}
