package cli

import (
	"fmt"

	"github.com/knoebber/dotfile/dotfileclient"
	"github.com/knoebber/dotfile/local"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

type pullCommand struct {
	alias      string
	username   string
	pullAll    bool
	createDirs bool
}

func (pc *pullCommand) run(ctx *kingpin.ParseContext) error {
	var err error

	storage := &local.Storage{Dir: flags.storageDir, Alias: pc.alias}

	client, err := newDotfileClient()
	if err != nil {
		return err
	}

	if pc.username != "" {
		client.Username = pc.username
	}
	if pc.pullAll {
		return pullAll(storage, client, pc.createDirs)
	} else if pc.alias != "" {
		return storage.Pull(client, pc.createDirs)
	} else {
		return errors.New("neither alias nor --all provided to pull")
	}
}

func pullAll(storage *local.Storage, client *dotfileclient.Client, createDirs bool) error {
	files, err := client.List(false)
	if err != nil {
		return err
	}

	for _, alias := range files {
		storage.Alias = alias
		if err := storage.Pull(client, createDirs); err != nil {
			fmt.Printf("failed to pull %q: %v\n", alias, err)
		} else {
			fmt.Println("pulled", alias)
		}
	}
	return nil
}

func addPullSubCommandToApplication(app *kingpin.Application) {
	pc := new(pullCommand)

	p := app.Command("pull", "pull changes from central service").Action(pc.run)
	p.Arg("alias", "the file to pull").StringVar(&pc.alias)
	p.Flag("username", "override config username").Short('u').StringVar(&pc.username)
	p.Flag("all", "pull all tracked files").Short('a').BoolVar(&pc.pullAll)
	p.Flag("create-dirs", "create directories that do not exist").Short('c').BoolVar(&pc.createDirs)
}
