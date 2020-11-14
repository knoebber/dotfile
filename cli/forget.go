package cli

import "gopkg.in/alecthomas/kingpin.v2"

type forgetCommand struct {
	alias   string
	commits bool
}

func (fc *forgetCommand) run(*kingpin.ParseContext) error {
	s, err := loadFile(fc.alias)
	if err != nil {
		return err
	}

	if fc.commits {
		return s.RemoveCommits()
	}

	return s.Forget()
}

func addForgetSubCommandToApplication(app *kingpin.Application) {
	fc := new(forgetCommand)

	c := app.Command("forget", "untrack a file - removes all tracking data").Action(fc.run)
	c.Arg("alias", "the file to forget").HintAction(flags.defaultAliasList).Required().StringVar(&fc.alias)
	c.Flag("commits", "remove all commits except the current").Short('c').BoolVar(&fc.commits)
}
