package cli

import (
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Allows for easy unit tests.
var execCommand = exec.Command

type editCommand struct {
	alias string
}

var errEditorEnvVarNotSet = errors.New("EDITOR environment variable must be set")

func (e *editCommand) run(*kingpin.ParseContext) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return errEditorEnvVarNotSet
	}

	s, err := loadFile(e.alias)
	if err != nil {
		return err
	}

	path, err := s.Path()
	if err != nil {
		return err
	}

	cmd := execCommand(editor, path)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func addEditSubCommandToApplication(app *kingpin.Application) {
	ec := new(editCommand)
	c := app.Command("edit", "open a tracked file in $EDITOR").Action(ec.run)
	c.Arg("alias", "the file to edit").HintAction(flags.defaultAliasList).Required().StringVar(&ec.alias)
}
