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
	var err error

	s, err := loadFileStorage(pc.fileName)
	if err = local.AssertClean(s); err != nil {
		return err
	}

	if pc.username != "" {
		s.User.Username = pc.username
	}

	if s.User.Username == "" {
		return errors.New("must set config username or use --username flag")
	}

	if pc.pullAll {
		return pullAll()
	} else if pc.fileName != "" {
		return s.Pull()
	} else {
		return errors.New("neither filename nor --all provided to pull")
	}
}

func pullAll() error {
	fmt.Println("TODO pull all")
	return nil
}

func addPullSubCommandToApplication(app *kingpin.Application) {
	pc := new(pullCommand)

	p := app.Command("pull", "pull changes from central service").Action(pc.run)
	p.Arg("file-name", "the file to pull").StringVar(&pc.fileName)
	p.Flag("username", "override config username").StringVar(&pc.username)
	p.Flag("all", "pull all tracked files").BoolVar(&pc.pullAll)
}
