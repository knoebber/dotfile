package cli

import (
	"fmt"

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

	if sc.username != "" {
		sc.remote = true
		storage.User.Username = sc.username
	}

	if sc.data {
		if !sc.remote {
			content, err = storage.GetJSON()
		} else {
			content, err = storage.GetRemoteJSON()
		}
	} else {
		if !sc.remote {
			content, err = storage.GetContents()
		} else {
			content, err = storage.GetRemoteContents()
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
