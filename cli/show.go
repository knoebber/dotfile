package cli

import (
	"fmt"

	"github.com/knoebber/dotfile/local"
	"gopkg.in/alecthomas/kingpin.v2"
)

type showCommand struct {
	alias    string
	data     bool
	remote   bool
	username string
}

func (sc *showCommand) run(ctx *kingpin.ParseContext) error {
	var (
		content []byte
		err     error
	)

	storage := &local.Storage{Dir: flags.storageDir, Alias: sc.alias}
	if !sc.remote {
		if err := storage.SetTrackingData(); err != nil {
			return err
		}
	}

	client, err := newDotfileClient()
	if err != nil {
		return err
	}

	if sc.username != "" {
		sc.remote = true
		client.Username = sc.username
	}

	if sc.data {
		if !sc.remote {
			content, err = storage.JSON()
		} else {
			content, err = client.TrackingDataBytes(sc.alias)
			// TODO this isn't a super helpful option - the json isn't formatted.
		}
	} else {
		if !sc.remote {
			content, err = storage.Content()
		} else {
			content, err = client.Content(sc.alias)
		}
	}

	if err != nil {
		return err
	}

	fmt.Print(string(content))
	return nil
}

func addShowSubCommandToApplication(app *kingpin.Application) {
	sc := new(showCommand)
	c := app.Command("show", "show the file").Action(sc.run)
	c.Arg("alias", "the file to show").Required().StringVar(&sc.alias)
	c.Flag("data", "show the file data in json format").Short('d').BoolVar(&sc.data)
	c.Flag("remote", "show the file on remote").Short('r').BoolVar(&sc.remote)
	c.Flag("username", "show the file owned by username on remote").Short('u').StringVar(&sc.username)
}
