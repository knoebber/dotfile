package cli

import (
	"bytes"
	"encoding/json"
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

func (sc *showCommand) run(*kingpin.ParseContext) error {
	var (
		content []byte
		err     error
	)

	if sc.remote || sc.username != "" {
		content, err = sc.showRemote()
	} else {
		content, err = sc.showLocal()
	}

	if err != nil {
		return err
	}

	fmt.Print(string(content))
	return nil
}

func (sc *showCommand) showLocal() ([]byte, error) {
	storage := &local.Storage{Dir: flags.storageDir, Alias: sc.alias}
	if err := storage.SetTrackingData(); err != nil {
		return nil, err
	}

	if !sc.data {
		return storage.DirtyContent()
	}

	return storage.JSON()
}

func (sc *showCommand) showRemote() ([]byte, error) {
	var buff bytes.Buffer

	client, err := newDotfileClient(false)
	if err != nil {
		return nil, err
	}
	if sc.username != "" {
		client.Username = sc.username
	}

	if !sc.data {
		return client.Content(sc.alias)
	}

	content, err := client.TrackingDataBytes(sc.alias)
	if err != nil {
		return nil, err
	}

	if err := json.Indent(&buff, content, "", "  "); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func addShowSubCommandToApplication(app *kingpin.Application) {
	sc := new(showCommand)
	c := app.Command("show", "show the file").Action(sc.run)
	c.Arg("alias", "the file to show").Required().StringVar(&sc.alias)
	c.Flag("data", "show the file data in json format").Short('d').BoolVar(&sc.data)
	c.Flag("remote", "show the file on remote").Short('r').BoolVar(&sc.remote)
	c.Flag("username", "show the file owned by username on remote").Short('u').StringVar(&sc.username)
}
