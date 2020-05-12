package cli

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/knoebber/dotfile/local"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

type cliConfig struct {
	storageDir string
	home       string
}

var (
	config cliConfig

	// Used for edit and diff.
	// Allows for easy unit tests.
	execCommand = exec.Command
)

func getHome() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "getting user home directory")
	}
	return home, nil
}

// Gets the default location for storing dotfile information.
func getDefaultStorageDir(home string) (storageDir string, err error) {
	configDir := filepath.Join(home, ".local/share/")
	if local.Exists(configDir) {
		// Priority one : ~/.local/share/dotfile
		storageDir = filepath.Join(configDir, "dotfile/")
	} else {
		// Priority two: ~/.dotfile/
		storageDir = filepath.Join(home, ".dotfile/")
	}

	if !local.Exists(storageDir) {
		err = os.Mkdir(storageDir, 0755)
	}
	return
}

func loadFile(alias string) (*local.Storage, error) {
	storage, err := local.LoadFile(config.home, config.storageDir, alias)

	if err != nil {
		return nil, errors.Wrapf(err, "loading file %#v", alias)
	}

	return storage, nil
}

func addFlagsToApplication(app *kingpin.Application) (err error) {
	config.home, err = getHome()
	if err != nil {
		return err
	}
	defaultStorageDir, err := getDefaultStorageDir(config.home)
	if err != nil {
		return err
	}

	app.Version("0.9.0")

	app.Flag("storage-dir", "The directory where dotfile data is stored").
		Default(defaultStorageDir).
		ExistingDirVar(&config.storageDir)

	return
}

// AddCommandsToApplication initializes the cli.
func AddCommandsToApplication(app *kingpin.Application) error {
	if err := addFlagsToApplication(app); err != nil {
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

	return nil
}
