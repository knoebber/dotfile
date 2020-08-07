package cli

import (
	"fmt"

	"gopkg.in/alecthomas/kingpin.v2"
)

type listCommand struct {
	remote   bool
	username string
}

func (lc *listCommand) run(ctx *kingpin.ParseContext) (err error) {
	var result []string

	storage, err := loadStorage()
	if err != nil {
		return err
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	if lc.username != "" {
		lc.remote = true
		client.Username = lc.username
	}

	if lc.remote {
		result, err = client.GetFileList()
	} else {
		result, err = storage.GetLocalFileList()
	}
	if err != nil {
		return
	}

	for _, f := range result {
		fmt.Println(f)
	}

	return nil
}

func addListSubCommandToApplication(app *kingpin.Application) {
	lc := new(listCommand)
	c := app.Command("ls", "list all tracked files, an asterisks signifies uncommited changes").Action(lc.run)
	c.Flag("remote", "read file list from remote").BoolVar(&lc.remote)
	c.Flag("username", "read files owned by username on remote").StringVar(&lc.username)
}
