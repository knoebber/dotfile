package cli

import (
	"fmt"

	"gopkg.in/alecthomas/kingpin.v2"
)

type showCommand struct {
	json     bool
	fileName string
}

func (pc *showCommand) run(ctx *kingpin.ParseContext) error {
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

func addShowSubCommandToApplication(app *kingpin.Application) {
	pc := new(showCommand)
	c := app.Command("show", "show the file").Action(pc.run)
	c.Arg("file-name", "the file to show").Required().StringVar(&pc.fileName)
	c.Flag("json", "show the file json schema").BoolVar(&pc.json)
}
