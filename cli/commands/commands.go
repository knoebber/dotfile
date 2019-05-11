package commands

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

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
