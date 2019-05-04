package commands

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

func AddCommandsToApplication(app *kingpin.Application) {
	addInitSubCommandToApplication(app)
	addCommitSubCommandToApplication(app)
	addPushSubCommandToApplication(app)
	addPullSubCommandToApplication(app)
}
