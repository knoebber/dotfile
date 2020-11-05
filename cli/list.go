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

	if lc.remote || lc.username != "" {
		result, err = lc.listRemote()
	} else {
		result, err = local.List(flags.storageDir, lc.path)
	}
	if err != nil {
		return err
	}

	for _, f := range result {
		fmt.Println(f)
	}

	return nil
}

func (lc *listCommand) listRemote() ([]string, error) {
	client, err := newDotfileClient(false)
	if err != nil {
		return nil, err
	}
	if lc.username != "" {
		client.Username = lc.username
	}

	return client.List(lc.path)
}

func addListSubCommandToApplication(app *kingpin.Application) {
	lc := new(listCommand)
	c := app.Command("ls", "list all tracked files, an asterisks signifies uncommitted changes").Action(lc.run)
	c.Flag("path", "include path in list").Short('p').BoolVar(&lc.path)
	c.Flag("remote", "read file list from remote").Short('r').BoolVar(&lc.remote)
	c.Flag("username", "read files owned by username on remote").Short('u').StringVar(&lc.username)
}
