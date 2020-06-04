package cli

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/knoebber/dotfile/file"
	"github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/alecthomas/kingpin.v2"
)

type diffCommand struct {
	fileName   string
	commitHash string
}

func (d *diffCommand) run(ctx *kingpin.ParseContext) error {
	var buff bytes.Buffer

	s, err := loadFile(d.fileName)
	if err != nil {
		return err
	}

	if d.commitHash == "" {
		d.commitHash = s.Tracking.Revision
	}

	diffs, err := file.Diff(s, d.commitHash, "")
	if err != nil {
		return err
	}

	for _, diff := range diffs {
		text := diff.Text

		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			_, _ = buff.WriteString("\x1b[32m")
			_, _ = buff.WriteString(text)
			_, _ = buff.WriteString("\x1b[0m")
		case diffmatchpatch.DiffDelete:
			_, _ = buff.WriteString("\x1b[31m")
			_, _ = buff.WriteString(text)
			_, _ = buff.WriteString("\x1b[0m")
		case diffmatchpatch.DiffEqual:
			_, _ = buff.WriteString(shortenEqualText(text))
		}
	}

	fmt.Println(buff.String())
	return nil
}

func shortenEqualText(text string) string {
	lines := strings.Split(text, "\n")
	if len(lines) <= 3 {
		return text
	}

	// There are atleast 4 lines.
	// Take the first and last two lines; discard the rest.
	return strings.Join(lines[:2], "\n") + "\n" + strings.Join(lines[len(lines)-2:], "\n")
}

func addDiffSubCommandToApplication(app *kingpin.Application) {
	dc := new(diffCommand)
	c := app.Command("diff", "check changes to tracked file").Action(dc.run)
	c.Arg("file-name", "file to check changes in").Required().StringVar(&dc.fileName)
	c.Arg("commit-hash",
		"the revision to diff against; default current").
		StringVar(&dc.commitHash)

}
