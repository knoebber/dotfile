package commands

import (
	"fmt"

	"gopkg.in/alecthomas/kingpin.v2"
)

type logCommand struct {
	fileName string
}

func (l *logCommand) run(ctx *kingpin.ParseContext) error {
	fmt.Printf("TODO: Log %#v\n", l.fileName)
	return nil
}

func addLogSubCommandToApplication(app *kingpin.Application) {
	lc := &logCommand{}
	c := app.Command("log", "shows revision history with commit hashes for a tracked file").Action(lc.run)
	c.Arg("file-name", "tracked file to show history for").Required().StringVar(&lc.fileName)
}
