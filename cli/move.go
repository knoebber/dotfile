package cli

import "gopkg.in/alecthomas/kingpin.v2"

type moveCommand struct {
	alias      string
	newPath    string
	parentDirs bool
}

func (mc *moveCommand) run(*kingpin.ParseContext) error {
	s, err := loadFile(mc.alias)
	if err != nil {
		return err
	}

	return s.Move(mc.newPath, mc.parentDirs)
}

func addMoveSubCommandToApplication(app *kingpin.Application) {
	mc := new(moveCommand)

	p := app.Command("mv", "move a file").Action(mc.run)
	p.Arg("alias", "the file to move").Required().StringVar(&mc.alias)
	p.Arg("new path", "the path to the new destination").StringVar(&mc.newPath)
	p.Flag("parent-dirs", "create parent directories that do not exist").Short('p').BoolVar(&mc.parentDirs)
}
