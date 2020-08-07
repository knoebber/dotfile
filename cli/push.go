package cli

import "gopkg.in/alecthomas/kingpin.v2"

type pushCommand struct {
	fileName string
}

func (pc *pushCommand) run(ctx *kingpin.ParseContext) error {
	s, err := loadFile(pc.fileName)
	if err != nil {
		return err
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	return s.Push(client)
}

func addPushSubCommandToApplication(app *kingpin.Application) {
	pc := new(pushCommand)

	p := app.Command("push", "push committed changes to a dotfile server").Action(pc.run)
	p.Arg("file-name", "the file to push").Required().StringVar(&pc.fileName)
}
