package cli

import "gopkg.in/alecthomas/kingpin.v2"

type pushCommand struct {
	alias string
}

func (pc *pushCommand) run(ctx *kingpin.ParseContext) error {
	s, err := loadFile(pc.alias)
	if err != nil {
		return err
	}

	client, err := newDotfileClient()
	if err != nil {
		return err
	}

	return s.Push(client)
}

func addPushSubCommandToApplication(app *kingpin.Application) {
	pc := new(pushCommand)

	p := app.Command("push", "push committed changes to a dotfile server").Action(pc.run)
	p.Arg("alias", "the file to push").Required().StringVar(&pc.alias)
}
