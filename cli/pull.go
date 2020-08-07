package cli

import (
	"fmt"

	"github.com/knoebber/dotfile/dotfileclient"
	"github.com/knoebber/dotfile/local"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

type pullCommand struct {
	fileName string
	username string
	pullAll  bool
}

func (pc *pullCommand) run(ctx *kingpin.ParseContext) error {
	var err error

	storage, err := loadStorage()
	if err != nil {
		return err
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	if pc.username != "" {
		client.Username = pc.username
	}

	if pc.pullAll {
		return pullAll(storage, client)
	} else if pc.fileName != "" {
		if err := storage.SetTrackingData(pc.fileName); err != nil {
			return err
		}

		return storage.Pull(client)
	} else {
		return errors.New("neither filename nor --all provided to pull")
	}
}

func pullAll(storage *local.Storage, client *dotfileclient.Client) error {
	files, err := client.GetFileList()
	if err != nil {
		return err
	}

	for _, alias := range files {
		fmt.Println("pulling", alias)
		if err := storage.SetTrackingData(alias); err != nil {
			return err
		}

		if err := storage.Pull(client); err != nil {
			fmt.Printf("failed to pull %q: %v\n", alias, err)
		}
	}
	return nil
}

func addPullSubCommandToApplication(app *kingpin.Application) {
	pc := new(pullCommand)

	p := app.Command("pull", "pull changes from central service").Action(pc.run)
	p.Arg("file-name", "the file to pull").StringVar(&pc.fileName)
	p.Flag("username", "override config username").StringVar(&pc.username)
	p.Flag("all", "pull all tracked files").BoolVar(&pc.pullAll)
}
