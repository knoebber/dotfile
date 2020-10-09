package cli

import (
	"fmt"

	"github.com/knoebber/dotfile/local"
	"gopkg.in/alecthomas/kingpin.v2"
)

type listCommand struct {
	path     bool
	remote   bool
	username string
}

func (lc *listCommand) run(*kingpin.ParseContext) (err error) {
	var result []string

	client, err := newDotfileClient()
	if err != nil {
		return err
	}

	if lc.username != "" {
		lc.remote = true
		client.Username = lc.username
	}

	if lc.remote {
		result, err = client.List(lc.path)
	} else {
		result, err = local.List(flags.storageDir, lc.path)
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
	c.Flag("path", "include path in list").Short('p').BoolVar(&lc.path)
	c.Flag("remote", "read file list from remote").Short('r').BoolVar(&lc.remote)
	c.Flag("username", "read files owned by username on remote").Short('u').StringVar(&lc.username)
}
