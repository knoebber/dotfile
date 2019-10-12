package cli

import (
	"github.com/knoebber/dotfile/file"
	"github.com/knoebber/dotfile/local"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

type logCommand struct {
	getStorage func() (*local.Storage, error)
	fileName   string
}

func (l *logCommand) run(ctx *kingpin.ParseContext) error {
	s, err := l.getStorage()
	if err != nil {
		return err
	}

	if err := file.Log(s, l.fileName); err != nil {
		return errors.Wrapf(err, "failed to get log for %#v", l.fileName)
	}
	return nil
}

func addLogSubCommandToApplication(app *kingpin.Application, gs func() (*local.Storage, error)) {
	lc := &logCommand{
		getStorage: gs,
	}
	c := app.Command("log", "shows revision history with commit hashes for a tracked file").Action(lc.run)
	c.Arg("file-name", "tracked file to show history for").Required().StringVar(&lc.fileName)
}
