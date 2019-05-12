package dotfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

func Init(path string, altName string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("%s not found", path)
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if altName == "" {
		// Create a name from the path of the file.
		// Examples: ~/.vimrc: vimrc
		//           ~/.config/i3/config: config
		//           ~/.config/alacritty/alacritty.yml: alacritty
		re := regexp.MustCompile(`(\w+)(\.\w+)?$`)
		altName = re.FindStringSubmatch(path)[1]
	}

	// Replace the full path with a relative path
	home := os.Getenv("HOME")
	if home == "" {
		return errors.New("HOME environment variable must be set")
	}
	re := regexp.MustCompile(home)
	path = re.ReplaceAllString(path, "~")

	fmt.Printf("Initialized %s as %s\n", path, altName)
	return nil
}
