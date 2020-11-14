package cli

import "gopkg.in/alecthomas/kingpin.v2"

type renameCommand struct {
	alias    string
	newAlias string
}

func (rc *renameCommand) run(*kingpin.ParseContext) error {
	s, err := loadFile(rc.alias)
	if err != nil {
		return err
	}

	return s.Rename(rc.newAlias)
}

func addRenameSubCommandToApplication(app *kingpin.Application) {
	rc := new(renameCommand)

	p := app.Command("rename", "change a files alias").Action(rc.run)
	p.Arg("alias", "the file to rename").HintAction(flags.defaultAliasList).Required().StringVar(&rc.alias)
	p.Arg("new alias", "the new name").Required().StringVar(&rc.newAlias)
}
