// Package cli creates a command line interface for interacting with local dotfiles.
package cli

import (
	"github.com/knoebber/dotfile/dotfileclient"
	"github.com/knoebber/dotfile/local"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Flags that multiple cli commands share.
type globalFlags struct {
	storageDir string
}

var flags globalFlags

func newDotfileClient() (*dotfileclient.Client, error) {
	user, err := local.GetUserConfig()
	if err != nil {
		return nil, err
	}

	return dotfileclient.New(user.Remote, user.Username, user.Token), nil
}

func loadFile(alias string) (*local.Storage, error) {
	storage := &local.Storage{
		Dir:   flags.storageDir,
		Alias: alias,
	}

	if err := storage.SetTrackingData(); err != nil {
		return nil, errors.Wrapf(err, "loading %q", alias)
	}

	return storage, nil
}

func setConfig(app *kingpin.Application) error {
	defaultStorageDir, err := local.DefaultStorageDir()
	if err != nil {
		return err
	}

	app.Version("0.9.0")

	app.Flag("storage-dir", "The directory where dotfile data is stored").
		Default(defaultStorageDir).
		ExistingDirVar(&flags.storageDir)

	return nil
}

// AddCommandsToApplication initializes the cli.
func AddCommandsToApplication(app *kingpin.Application) error {
	if err := setConfig(app); err != nil {
		return err
	}

	addInitSubCommandToApplication(app)
	addShowSubCommandToApplication(app)
	addListSubCommandToApplication(app)
	addEditSubCommandToApplication(app)
	addDiffSubCommandToApplication(app)
	addLogSubCommandToApplication(app)
	addCheckoutSubCommandToApplication(app)
	addCommitSubCommandToApplication(app)
	addPushSubCommandToApplication(app)
	addPullSubCommandToApplication(app)
	addConfigSubCommandToApplication(app)
	addMoveSubCommandToApplication(app)
	addRenameSubCommandToApplication(app)
	addForgetSubCommandToApplication(app)

	return nil
}
