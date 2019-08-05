package cli

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	defaultConfigDir  string = ".dotfile"
	defaultConfigName string = "files.json"
)

type commonFlags struct {
	configDir  string
	configName string
}

func AddCommandsToApplication(app *kingpin.Application) {
	addInitSubCommandToApplication(app)
	addEditSubCommandToApplication(app)
	addDiffSubCommandToApplication(app)
	addLogSubCommandToApplication(app)
	addCheckoutSubCommandToApplication(app)
	addCommitSubCommandToApplication(app)
	addPushSubCommandToApplication(app)
	addPullSubCommandToApplication(app)
}

// Flags that all dotfile commands share
func addCommonFlags(app *kingpin.Application, configDir *string, configName *string) {
	app.Flag("config-dir", "The directory where version control data is stored").
		Default(defaultConfigDir).
		StringVar(configDir)
	app.Flag("config-name", "The main json file that tracks checked in files").
		Default(defaultConfigName).
		StringVar(configName)
}
