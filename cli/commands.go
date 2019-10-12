package cli

import (
	"github.com/knoebber/dotfile/local"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"

	"fmt"
	"os"
)

const (
	defaultStorageDir  string = ".dotfile/"
	defaultStorageName string = "files.json"
)

// Dotfile depends on the operating system having the concept of a home directory.
func getHome() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "failed to get user home directory")
	}
	return home, nil
}

// Returns a function that initializes dotfile storage.
// The result function must be ran at the time of a command being run so that
// the user can override default storage configuration with --storage-dir or --storage-name.
func getStorageClosure(home string, dir, name *string) func() (*local.Storage, error) {
	return func() (*local.Storage, error) {
		storage, err := local.NewStorage(home, *dir, *name)

		if err != nil {
			return nil, errors.Wrap(err, "getting local storage")
		}

		return storage, nil
	}
}

// AddCommandsToApplication initializes the cli.
func AddCommandsToApplication(app *kingpin.Application) error {
	var (
		storageDirectory string
		storageName      string
	)

	home, err := getHome()
	if err != nil {
		return err
	}

	gs := getStorageClosure(home, &storageDirectory, &storageName)

	app.Flag("storage-dir", "The directory where version control data is stored").
		Default(fmt.Sprintf("%s/%s", home, defaultStorageDir)).
		StringVar(&storageDirectory)
	app.Flag("storage-name", "The main json file that tracks checked in files").
		Default(defaultStorageName).
		StringVar(&storageName)

	addInitSubCommandToApplication(app, gs)
	addEditSubCommandToApplication(app, gs)
	addDiffSubCommandToApplication(app, gs)
	addLogSubCommandToApplication(app, gs)
	addCheckoutSubCommandToApplication(app, gs)
	addCommitSubCommandToApplication(app, gs)
	addPushSubCommandToApplication(app, gs)
	addPullSubCommandToApplication(app, gs)

	return nil
}
