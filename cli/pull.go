package cli

import (
	"fmt"

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
	username := config.user.Username

	if pc.username != "" {
		username = pc.username
	}

	if username == "" {
		return errors.New("must set config username or use --username flag")
	}

	if pc.pullAll {
		return pullAll()
	} else if pc.fileName != "" {
		return pullFile(username, pc.fileName)
	} else {
		return errors.New("neither filename nor --all provided to pull")
	}

	return nil
}

func pullAll() error {
	fmt.Println("TODO pull all")
	return nil
}

func pullFile(username, alias string) error {
	return local.Pull(config.user.Remote, username, alias)
}

func addPullSubCommandToApplication(app *kingpin.Application) {
	pc := new(pullCommand)

	p := app.Command("pull", "pull changes from central service").Action(pc.run)
	p.Arg("file-name", "the file to pull").StringVar(&pc.fileName)
	p.Flag("username", "override config username").StringVar(&pc.username)
	p.Flag("all", "pull all tracked files").BoolVar(&pc.pullAll)
}
