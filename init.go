package dotfile

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

func Init(path string, altName string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("%s not found", path)
	}
	if altName == "" {
		// Create a name from the path of the file.
		// Examples: ~/.vimrc: vimrc
		//           ~/.config/i3/config: config
		//           ~/.config/alacritty/alacritty.yml: alacritty
		re := regexp.MustCompile(`(\w+)(\.\w+)?$`)
		altName = re.FindStringSubmatch(path)[1]
	}

	path = filepath.Dir(path)

	fmt.Printf("Initialized %s as %s\n", path, altName)
	return nil
}
