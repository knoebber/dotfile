package cli

import "gopkg.in/alecthomas/kingpin.v2"

type removeCommand struct {
	alias string
}

func (rc *removeCommand) run(*kingpin.ParseContext) error {
	s, err := loadFile(rc.alias)
	if err != nil {
		return err
	}

	return s.Remove()
}

func addRemoveSubCommandToApplication(app *kingpin.Application) {
	rc := new(removeCommand)

	p := app.Command("rm", "remove the tracked file and all its data").Action(rc.run)
	p.Arg("alias", "the file to remove").Required().StringVar(&rc.alias)
}
