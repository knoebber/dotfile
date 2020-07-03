package cli

import (
	"fmt"

	"github.com/knoebber/dotfile/local"
	"gopkg.in/alecthomas/kingpin.v2"
)

type configCommand struct {
	key   string
	value string
}

func (cc *configCommand) run(ctx *kingpin.ParseContext) error {
	if cc.key == "" {
		fmt.Println(config.user)
		return nil
	}

	return local.SetUserConfig(config.home, cc.key, cc.value)
}

func addConfigSubCommandToApplication(app *kingpin.Application) {
	cc := new(configCommand)

	p := app.Command("config", "set dotfile configurations").Action(cc.run)
	p.Arg("key", "the config key to change - <remote/username/token>").EnumVar(&cc.key,
		"remote",
		"username",
		"token",
	)

	p.Arg("value", "the new value").StringVar(&cc.value)
}
