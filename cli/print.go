package cli

import (
	"fmt"

	"gopkg.in/alecthomas/kingpin.v2"
)

type printCommand struct {
	json     bool
	fileName string
}

func (pc *printCommand) run(ctx *kingpin.ParseContext) error {
	var (
		content []byte
		err     error
	)

	s, err := loadFile(pc.fileName)
	if err != nil {
		return err
	}

	if pc.json {
		content, err = s.GetJSON()
	} else {
		content, err = s.GetContents()
	}

	if err != nil {
		return err
	}

	fmt.Print(string(content))
	return nil
}

func addPrintSubCommandToApplication(app *kingpin.Application) {
	pc := new(printCommand)
	c := app.Command("print", "print the file").Action(pc.run)
	c.Arg("file-name", "the file to print").Required().StringVar(&pc.fileName)
	c.Flag("json", "print the file json schema").BoolVar(&pc.json)
}
