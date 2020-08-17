package cli

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

type moveCommand struct {
	alias      string
	newPath    string
	createDirs bool
}

func (mc *moveCommand) run(ctx *kingpin.ParseContext) error {
	s, err := loadFile(mc.alias)
	if err != nil {
		return err
	}

	return s.Move(mc.newPath, mc.createDirs)
}

func addMVSubCommandToApplication(app *kingpin.Application) {
	mc := new(moveCommand)

	p := app.Command("mv", "move a file").Action(mc.run)
	p.Arg("alias", "the file to push").Required().StringVar(&mc.alias)
	p.Arg("new path", "the path to the new destination").StringVar(&mc.newPath)
	p.Flag("create-dirs", "create directories that do not exist").Short('c').BoolVar(&mc.createDirs)
}
