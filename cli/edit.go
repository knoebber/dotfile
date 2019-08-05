package cli

import (
	"os"
	"os/exec"

	"github.com/knoebber/dotfile/file"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

type editCommand struct {
	commonFlags
	fileName string
}

var (
	execCommand = exec.Command

	ErrEditorEnvVarNotSet = errors.New("EDITOR environment variable must be set")
)

func (e *editCommand) run(ctx *kingpin.ParseContext) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return ErrEditorEnvVarNotSet
	}

	path, err := file.GetPath(e.fileName)
	if err != nil {
		return errors.Wrapf(err, "error getting path for filename: %#v", e.fileName)
	}

	cmd := execCommand(editor, path)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func addEditSubCommandToApplication(app *kingpin.Application) {
	ec := &editCommand{}
	c := app.Command("edit", "open a tracked file in $EDITOR").Action(ec.run)
	c.Arg("file-name", "the file to edit").Required().StringVar(&ec.fileName)
}
