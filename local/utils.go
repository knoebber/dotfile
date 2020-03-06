package local

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RelativePath converts a file path into a path relative to the users home directory.
func RelativePath(path, home string) (relativePath string, err error) {
	var fullPath string

	_, fileErr := os.Stat(path)
	if os.IsNotExist(fileErr) {
		err = fmt.Errorf("%#v not found", path)
		return
	}

	fullPath, err = filepath.Abs(path)
	if err != nil {
		return
	}

	relativePath = strings.Replace(fullPath, home, "~", 1)

	return
}

func fullPath(relativePath, home string) string {
	return strings.Replace(relativePath, "~", home, 1)
}

// Creates a directory and a file.
// Returns true if any files were created.
func createIfNotExist(dir, fileName string) (created bool, err error) {
	_, err = os.Stat(dir)
	if os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
		created = true
	} else if err != nil {
		return false, err
	}

	_, err = os.Stat(fileName)
	if os.IsNotExist(err) {
		f, createErr := os.Create(fileName)
		if createErr != nil {
			return false, createErr
		}
		defer f.Close()

		created = true
	} else if err != nil {
		return
	}
	err = nil
	return
}
