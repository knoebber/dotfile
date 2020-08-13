package local

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/knoebber/dotfile/usererror"
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

func configPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dotfileDir := filepath.Join(configDir, "dotfile")
	if !exists(dotfileDir) {
		if err := os.Mkdir(dotfileDir, 0755); err != nil {
			return "", err
		}
	}

	return filepath.Join(dotfileDir, "dotfile.json"), nil
}

// GetUserConfig reads the user config.
// Creates a default file when it doesn't yet exist.
func GetUserConfig() (*UserConfig, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	cfg := new(UserConfig)

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

	path, err := configPath()
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
		return usererror.Invalid(fmt.Sprintf("%#v is not a valid config key", key))
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
