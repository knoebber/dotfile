package cli

import (
	"fmt"

	"github.com/knoebber/dotfile/dotfileclient"
	"gopkg.in/alecthomas/kingpin.v2"
)

type showCommand struct {
	fileName string
	data     bool
	remote   bool
	username string
}

func (sc *showCommand) run(ctx *kingpin.ParseContext) error {
	var (
		content []byte
		err     error
	)

	storage, err := loadFile(sc.fileName)
	if err != nil {
		return err
	}

	client := dotfileclient.New(storage.User.Remote, storage.User.Username, storage.User.Token)

	if sc.username != "" {
		sc.remote = true
		client.Username = sc.username
	}

	if sc.data {
		if !sc.remote {
			content, err = storage.GetJSON()
		} else {
			content, err = client.GetTrackingDataBytes(sc.fileName)
			// TODO this isn't a super helpful option - the json isn't formatted.
		}
	} else {
		if !sc.remote {
			content, err = storage.GetContents()
		} else {
			content, err = client.GetContents(sc.fileName)
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
	c.Arg("file-name", "the file to show").Required().StringVar(&sc.fileName)
	c.Flag("data", "show the file data in json format").BoolVar(&sc.data)
	c.Flag("remote", "show the file on remote").BoolVar(&sc.remote)
	c.Flag("username", "show the file owned by username on remote").StringVar(&sc.username)
}
