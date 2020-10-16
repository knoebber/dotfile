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

// Config contains local user settings for dotfile.
type Config struct {
	Remote   string `json:"remote"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

func (c *Config) String() string {
	return fmt.Sprintf("remote: %q\nusername: %q\ntoken: %q",
		c.Remote,
		c.Username,
		c.Token,
	)
}

func createDefaultConfig(path string) ([]byte, error) {
	var newCfg Config

	bytes, err := json.MarshalIndent(newCfg, "", jsonIndent)
	if err != nil {
		return nil, errors.Wrap(err, "marshalling new user config file")
	}

	if err = ioutil.WriteFile(path, bytes, 0644); err != nil {
		return nil, errors.Wrap(err, "saving new user config file")
	}

	return bytes, nil
}

func configBytes(path string) ([]byte, error) {
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
		if err := os.MkdirAll(dotfileDir, 0755); err != nil {
			return "", err
		}
	}

	return filepath.Join(dotfileDir, "dotfile.json"), nil
}

// ReadConfig reads the user's config.
// Creates a default file when it doesn't yet exist.
func ReadConfig() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	cfg := new(Config)

	bytes, err := configBytes(path)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(bytes, cfg); err != nil {
		return nil, errors.Wrapf(err, "unmarshaling user config to struct")
	}

	return cfg, nil
}

// SetConfig sets a value in the dotfile config json file.
func SetConfig(key string, value string) error {
	cfg := make(map[string]*string)

	path, err := configPath()
	if err != nil {
		return err
	}

	bytes, err := configBytes(path)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(bytes, &cfg); err != nil {
		return errors.Wrapf(err, "unmarshaling user config to map")
	}

	if _, ok := cfg[key]; !ok {
		return usererror.Invalid(fmt.Sprintf("%q is not a valid config key", key))
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
