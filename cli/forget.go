package cli

import "gopkg.in/alecthomas/kingpin.v2"

type forgetCommand struct {
	alias string
}

func (fc *forgetCommand) run(ctx *kingpin.ParseContext) error {
	s, err := loadFile(fc.alias)
	if err != nil {
		return err
	}

	return s.Forget()
}

func addForgetSubCommandToApplication(app *kingpin.Application) {
	fc := new(forgetCommand)

	p := app.Command("forget", "untrack a file - removes all tracking data").Action(fc.run)
	p.Arg("alias", "the file to forget").Required().StringVar(&fc.alias)
}
