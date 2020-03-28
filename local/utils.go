package local

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// RelativePath converts path into a relative path.
// Returns an error when path does not exist.
func RelativePath(path, home string) (string, error) {
	_, fileErr := os.Stat(path)
	if os.IsNotExist(fileErr) {
		return "", fmt.Errorf("%#v not found", path)
	}

	fullPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	relativePath := strings.Replace(fullPath, home, "~", 1)

	if !strings.Contains(relativePath, "~") {
		return "", fmt.Errorf("%#v is not in home directory", path)
	}

	return relativePath, nil
}

func fullPath(relativePath, home string) string {
	return strings.Replace(relativePath, "~", home, 1)
}

// Creates a directory and a file.
// Returns true if any files were created.
func createIfNotExist(dir, fileName string) (bool, error) {
	if !exists(dir) {
		if createErr := os.Mkdir(dir, 0755); createErr != nil {
			return false, createErr
		}
	}

	if exists(fileName) {
		return false, nil
	}

	f, createErr := os.Create(fileName)

	if createErr != nil {
		return false, createErr
	}

	if err := f.Close(); err != nil {
		return false, errors.Wrapf(err, "closing %#v", fileName)
	}

	return true, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}
