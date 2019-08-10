package cli

import (
	"github.com/knoebber/dotfile/file"
	"gopkg.in/alecthomas/kingpin.v2"

	"fmt"
	"os"
)

const (
	defaultDataDir  string = ".dotfile/"
	defaultDataName string = "files.json"
)

// Dotfile depends on the system having the concept of a home directory.
func getHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return home
}

func AddCommandsToApplication(app *kingpin.Application) {
	data := &file.Data{
		Home: getHome(),
	}

	app.Flag("data-dir", "The directory where version control data is stored").
		Default(fmt.Sprintf("%s/%s", data.Home, defaultDataDir)).
		StringVar(&data.Dir)
	app.Flag("data-name", "The main json file that tracks checked in files").
		Default(defaultDataName).
		StringVar(&data.Name)

	addInitSubCommandToApplication(app, data)
	addEditSubCommandToApplication(app, data)
	addDiffSubCommandToApplication(app, data)
	addLogSubCommandToApplication(app, data)
	addCheckoutSubCommandToApplication(app, data)
	addCommitSubCommandToApplication(app, data)
	addPushSubCommandToApplication(app, data)
	addPullSubCommandToApplication(app, data)
}
