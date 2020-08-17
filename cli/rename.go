package cli

import "gopkg.in/alecthomas/kingpin.v2"

type renameCommand struct {
	alias    string
	newAlias string
}

func (rc *renameCommand) run(ctx *kingpin.ParseContext) error {
	s, err := loadFile(rc.alias)
	if err != nil {
		return err
	}

	return s.Rename(rc.newAlias)
}

func addRenameSubCommandToApplication(app *kingpin.Application) {
	rc := new(renameCommand)

	p := app.Command("rename", "change a files alias").Action(rc.run)
	p.Arg("alias", "the file to rename").Required().StringVar(&rc.alias)
	p.Arg("new alias", "the new name").StringVar(&rc.newAlias)
}
