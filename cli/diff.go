package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/knoebber/dotfile/file"
	"github.com/knoebber/dotfile/local"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	diffCmd     = "diff"
	diffType    = "-u" // unified view
	colorOption = "--color"
)

type diffCommand struct {
	fileName   string
	commitHash string
}

func (d *diffCommand) run(ctx *kingpin.ParseContext) error {
	s, err := loadFile(d.fileName)
	if err != nil {
		return err
	}

	if d.commitHash == "" {
		d.commitHash = s.Tracking.Revision
	}

	if _, err := exec.LookPath(diffCmd); err != nil {
		return fmt.Errorf("%s not found in $PATH", diffCmd)
	}

	return diff(s, d.fileName, d.commitHash)
}

// Color is supported by GNU diff utilities 3.4 and greater.
// https://savannah.gnu.org/forum/forum.php?forum_id=8639
func diffSupportsColor() bool {
	return execCommand(diffCmd, colorOption, "/dev/null", "/dev/null").Run() == nil
}

func diff(s *local.Storage, fileName, hash string) error {
	var cmd *exec.Cmd

	path, err := s.GetPath()
	if err != nil {
		return err
	}

	lastRevision, err := file.UncompressRevision(s, hash)
	if err != nil {
		return err
	}

	if diffSupportsColor() {
		cmd = execCommand(diffCmd, diffType, colorOption, "-", path)
	} else {
		cmd = execCommand(diffCmd, diffType, "-", path)
	}

	cmd.Stdin = lastRevision
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	if err == nil {
		fmt.Println("No changes")
		return nil
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		return err
	}

	exitCode := exitErr.ExitCode()

	// diff returns 1 when file has changes.
	if exitCode == 1 {
		return nil
	} else if exitCode > 1 {
		return err
	}

	return nil
}

func addDiffSubCommandToApplication(app *kingpin.Application) {
	dc := new(diffCommand)
	c := app.Command("diff", "check changes to tracked file").Action(dc.run)
	c.Arg("file-name", "file to check changes in").Required().StringVar(&dc.fileName)
	c.Arg("commit-hash",
		"the revision to diff against; default current").
		StringVar(&dc.commitHash)

}
