package cli

import (
	"fmt"

	"github.com/knoebber/dotfile/file"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

type pullCommand struct {
	storage  *file.Storage
	fileName string
	pullAll  bool
}

func (pc *pullCommand) run(ctx *kingpin.ParseContext) error {
	if pc.pullAll {
		fmt.Println("TODO: Pull all")
	} else if pc.fileName != "" {
		fmt.Printf("TODO: Pull %#v\n", pc.fileName)
	} else {
		return errors.New("neither filename nor --all provided to pull")
	}
	return nil
}

func addPullSubCommandToApplication(app *kingpin.Application, storage *file.Storage) {
	pc := &pullCommand{
		storage: storage,
	}
	p := app.Command("pull", "pull changes from central service").Action(pc.run)
	p.Arg("file-name", "the file to pull").StringVar(&pc.fileName)
	p.Flag("all", "pull all tracked files").BoolVar(&pc.pullAll)
}
