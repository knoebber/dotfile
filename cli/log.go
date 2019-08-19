package cli

import (
	"fmt"

	"github.com/knoebber/dotfile/file"
	"gopkg.in/alecthomas/kingpin.v2"
)

type logCommand struct {
	getStorage func() (*file.Storage, error)
	fileName   string
}

func (l *logCommand) run(ctx *kingpin.ParseContext) error {
	_, err := l.getStorage()
	if err != nil {
		return err
	}

	fmt.Printf("TODO: Log %#v\n", l.fileName)
	return nil
}

func addLogSubCommandToApplication(app *kingpin.Application, gs func() (*file.Storage, error)) {
	lc := &logCommand{
		getStorage: gs,
	}
	c := app.Command("log", "shows revision history with commit hashes for a tracked file").Action(lc.run)
	c.Arg("file-name", "tracked file to show history for").Required().StringVar(&lc.fileName)
}
