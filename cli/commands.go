package cli

import (
	"os"

	"github.com/knoebber/dotfile/local"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

type cliConfig struct {
	storageDir string
	configDir  string
	home       string
	user       *local.UserConfig
}

var config cliConfig

func getHome() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "getting user home directory")
	}
	return home, nil
}

func loadFile(alias string) (*local.Storage, error) {
	storage, err := local.LoadFile(config.home, config.storageDir, alias)

	if err != nil {
		return nil, errors.Wrapf(err, "loading file %#v", alias)
	}

	return storage, nil
}

func setConfig(app *kingpin.Application) error {
	home, err := getHome()
	if err != nil {
		return err
	}
	config.home = home

	defaultStorageDir, err := local.GetDefaultStorageDir(config.home)
	if err != nil {
		return err
	}

	config.user, err = local.GetUserConfig(config.home)
	if err != nil {
		return err
	}

	app.Version("0.9.0")

	app.Flag("storage-dir", "The directory where dotfile data is stored").
		Default(defaultStorageDir).
		ExistingDirVar(&config.storageDir)

	return nil
}

// AddCommandsToApplication initializes the cli.
func AddCommandsToApplication(app *kingpin.Application) error {
	if err := setConfig(app); err != nil {
		return err
	}

	addInitSubCommandToApplication(app)
	addEditSubCommandToApplication(app)
	addDiffSubCommandToApplication(app)
	addLogSubCommandToApplication(app)
	addCheckoutSubCommandToApplication(app)
	addCommitSubCommandToApplication(app)
	addPushSubCommandToApplication(app)
	addPullSubCommandToApplication(app)
	addConfigSubCommandToApplication(app)

	return nil
}
