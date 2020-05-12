package cli

import (
	"fmt"

	"github.com/knoebber/dotfile/local"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

type pullCommand struct {
	getStorage func() (*local.Storage, error)
	fileName   string
	pullAll    bool
}

func (pc *pullCommand) run(ctx *kingpin.ParseContext) error {
	_, err := loadFile(pc.fileName)
	if err != nil {
		return err
	}

	if pc.pullAll {
		fmt.Println("TODO: Pull all")
	} else if pc.fileName != "" {
		fmt.Printf("TODO: Pull %#v\n", pc.fileName)
	} else {
		return errors.New("neither filename nor --all provided to pull")
	}
	return nil
}

func addPullSubCommandToApplication(app *kingpin.Application) {
	pc := new(pullCommand)

	p := app.Command("pull", "pull changes from central service").Action(pc.run)
	p.Arg("file-name", "the file to pull").StringVar(&pc.fileName)
	p.Flag("all", "pull all tracked files").BoolVar(&pc.pullAll)
}
