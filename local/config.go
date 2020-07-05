package local

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/knoebber/dotfile/usererr"
	"github.com/pkg/errors"
)

const defaultRemote = "https://dotfilehub.com"

// UserConfig contains local user settings for dotfile.
type UserConfig struct {
	Remote   string `json:"remote"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

func (uc *UserConfig) String() string {
	return fmt.Sprintf("remote: %#v\nusername: %#v\ntoken: %#v",
		uc.Remote,
		uc.Username,
		uc.Token,
	)
}

func getConfigPath(home string) (string, error) {
	var configDir string

	config := filepath.Join(home, ".config/")

	if exists(config) {
		// Priority one : ~/.config/dotfile
		configDir = filepath.Join(config, "dotfile/")
	} else {
		// Priority two: ~/.dotfile/
		configDir = filepath.Join(home, ".dotfile/")
	}

	if !exists(configDir) {
		if err := os.Mkdir(configDir, 0755); err != nil {
			return "", errors.Wrap(err, "creating config directory")
		}
	}

	return filepath.Join(configDir, "config.json"), nil
}

func createDefaultConfig(path string) ([]byte, error) {
	newCfg := UserConfig{Remote: defaultRemote}

	bytes, err := json.MarshalIndent(newCfg, "", jsonIndent)
	if err != nil {
		return nil, errors.Wrap(err, "marshalling new user config file")
	}

	if err = ioutil.WriteFile(path, bytes, 0644); err != nil {
		return nil, errors.Wrap(err, "saving new user config file")
	}

	return bytes, nil
}

func getConfigBytes(path string) ([]byte, error) {
	if !exists(path) {
		return createDefaultConfig(path)
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "reading config directory")
	}

	return bytes, nil
}

// GetUserConfig reads the user config.
// Creates a default file when it doesn't yet exist.
func GetUserConfig(home string) (*UserConfig, error) {
	cfg := new(UserConfig)

	path, err := getConfigPath(home)
	if err != nil {
		return nil, err
	}

	bytes, err := getConfigBytes(path)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(bytes, cfg); err != nil {
		return nil, errors.Wrapf(err, "unmarshaling user config to struct")
	}

	return cfg, nil
}

// SetUserConfig sets a value in the dotfile config json file.
func SetUserConfig(home string, key string, value string) error {
	cfg := make(map[string]*string)

	path, err := getConfigPath(home)
	if err != nil {
		return err
	}

	bytes, err := getConfigBytes(path)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(bytes, &cfg); err != nil {
		return errors.Wrapf(err, "unmarshaling user config to map")
	}

	if _, ok := cfg[key]; !ok {
		return usererr.Invalid(fmt.Sprintf("%#v is not a valid config key", key))
	}

	cfg[key] = &value

	bytes, err = json.MarshalIndent(cfg, "", jsonIndent)
	if err != nil {
		return errors.Wrap(err, "marshalling updated config map")
	}

	if err = ioutil.WriteFile(path, bytes, 0644); err != nil {
		return errors.Wrap(err, "saving updated config file")
	}

	return nil
}
