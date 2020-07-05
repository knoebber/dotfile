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
	storage, err := local.NewStorage(config.home, config.storageDir)
	if err != nil {
		return errors.Wrap(err, "getting storage")
	}

	// Create a new config so that the username can be overridden when flag is provided.
	// Dereference the pointer to avoid mutating the global config variable.
	cfg := *config.user

	if pc.username != "" {
		cfg.Username = pc.username
	}

	if cfg.Username == "" {
		return errors.New("must set config username or use --username flag")
	}

	if pc.pullAll {
		return pullAll()
	} else if pc.fileName != "" {
		return pullFile(storage, &cfg, pc.fileName)
	} else {
		return errors.New("neither filename nor --all provided to pull")
	}

	return nil
}

func pullAll() error {
	fmt.Println("TODO pull all")
	return nil
}

func pullFile(s *local.Storage, cfg *local.UserConfig, alias string) error {
	return local.Pull(s, cfg, alias)
}

func addPullSubCommandToApplication(app *kingpin.Application) {
	pc := new(pullCommand)

	p := app.Command("pull", "pull changes from central service").Action(pc.run)
	p.Arg("file-name", "the file to pull").StringVar(&pc.fileName)
	p.Flag("username", "override config username").StringVar(&pc.username)
	p.Flag("all", "pull all tracked files").BoolVar(&pc.pullAll)
}
