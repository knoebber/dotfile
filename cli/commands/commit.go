package commands

import (
	"fmt"

	"gopkg.in/alecthomas/kingpin.v2"
)

type commitCommand struct {
	fileName string
}

func (c *commitCommand) run(ctx *kingpin.ParseContext) error {
	fmt.Printf("TODO: Commit %#v\n", c.fileName)
	return nil
}

func addCommitSubCommandToApplication(app *kingpin.Application) {
	cc := &commitCommand{}
	c := app.Command("commit", "commit file to working tree").Action(cc.run)
	c.Arg("file-name", "the file to track").Required().StringVar(&cc.fileName)

}
