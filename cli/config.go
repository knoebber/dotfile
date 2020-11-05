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

func (cc *configCommand) run(*kingpin.ParseContext) error {
	if cc.key != "" && cc.value != "" {
		return local.SetConfig(flags.configPath, cc.key, cc.value)
	}

	config, err := local.ReadConfig(flags.configPath)
	if err != nil {
		return err
	}

	if cc.key == "remote" {
		fmt.Println(config.Remote)
	} else if cc.key == "username" {
		fmt.Println(config.Username)
	} else if cc.key == "token" {
		fmt.Println(config.Token)
	} else {
		fmt.Println(config)
	}
	return nil
}

func addConfigSubCommandToApplication(app *kingpin.Application) {
	cc := new(configCommand)

	p := app.Command("config", "set or print dotfile configurations").Action(cc.run)
	p.Arg("key", "the config key to change or print - <remote/username/token>").EnumVar(&cc.key,
		"remote",
		"username",
		"token",
	)

	p.Arg("value", "the new value").StringVar(&cc.value)
}
