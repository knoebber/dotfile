package commands

import (
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

type editCommand struct {
	fileName string
}

func (e *editCommand) run(ctx *kingpin.ParseContext) error {
	editor := os.Getenv("EDITOR")
	editorPath, err := exec.LookPath(editor)
	if err != nil {
		return errors.Wrapf(err, "error getting path for editor %#v", editor)
	}
	cmd := exec.Command(editorPath, e.fileName)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func addEditSubCommandToApplication(app *kingpin.Application) {
	ec := &editCommand{}
	c := app.Command("edit", "open a tracked file in $EDITOR").Action(ec.run)
	c.Arg("file-name", "the file to track").Required().StringVar(&ec.fileName)
}
