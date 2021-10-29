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
	storageDir       string
	configPath       string
	defaultAliasList func() []string
}

var flags globalFlags

func newDotfileClient(tokenRequired bool) (*dotfileclient.Client, error) {
	config, err := local.ReadConfig(flags.configPath)
	if err != nil {
		return nil, err
	}
	if config.Remote == "" {
		return nil, errors.New("config value for \"remote\" must be set")
	}
	if config.Username == "" {
		return nil, errors.New("config value for \"username\" must be set")
	}
	if tokenRequired && config.Token == "" {
		return nil, errors.New("config value for \"token\" must be set")
	}

	return dotfileclient.New(config.Remote, config.Username, config.Token), nil
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

	defaultConfigPath, err := local.DefaultConfigPath()
	if err != nil {
		return err
	}

	// Used for tab completion in commands that have an alias argument.
	flags.defaultAliasList = local.ListAliases(defaultStorageDir)

	app.Version("1.0.5")

	app.Flag("storage-dir", "The directory where dotfile data is stored").
		Default(defaultStorageDir).
		ExistingDirVar(&flags.storageDir)
	app.Flag("config-file", "The json file to use for configuration").
		Default(defaultConfigPath).
		StringVar(&flags.configPath)
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
	addRemoveSubCommandToApplication(app)

	return nil
}
