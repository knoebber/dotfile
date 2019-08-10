package cli

import (
	"fmt"

	"github.com/knoebber/dotfile/file"
	"gopkg.in/alecthomas/kingpin.v2"
)

type logCommand struct {
	data     *file.Data
	fileName string
}

func (l *logCommand) run(ctx *kingpin.ParseContext) error {
	fmt.Printf("TODO: Log %#v\n", l.fileName)
	return nil
}

func addLogSubCommandToApplication(app *kingpin.Application, data *file.Data) {
	lc := &logCommand{
		data: data,
	}
	c := app.Command("log", "shows revision history with commit hashes for a tracked file").Action(lc.run)
	c.Arg("file-name", "tracked file to show history for").Required().StringVar(&lc.fileName)
}
