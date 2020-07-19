// Package cli creates a command line interface for interacting with local dotfiles.
package cli

import (
	"fmt"
	"os"

	"github.com/knoebber/dotfile/local"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

type cliConfig struct {
	storageDir string
	home       string
}

var config cliConfig

func getHome() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "getting user home directory")
	}
	return home, nil
}

func loadFileStorage(alias string) (*local.Storage, error) {
	configPath, err := local.GetConfigPath(config.home)
	if err != nil {
		return nil, err
	}

	storage, err := local.NewStorage(config.home, config.storageDir, configPath)
	if err != nil {
		return nil, errors.Wrap(err, "creating local storage")
	}

	if err := storage.SetTrackingData(alias); err != nil {
		return nil, errors.Wrapf(err, "loading %#v", alias)
	}
	return storage, nil
}

func loadFile(alias string) (*local.Storage, error) {
	storage, err := loadFileStorage(alias)
	if err != nil {
		return nil, err
	}

	if !storage.HasFile {
		return nil, fmt.Errorf("%#v is not tracked", alias)
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
	addShowSubCommandToApplication(app)
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
