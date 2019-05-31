package dotfile

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

var (
	notTrackedErr = errors.New("file not tracked, use 'dot init <file>' first")
)

// Init sets up a file for dotfile to track.
func Init(path string, fileName string) error {
	d := &data{}

	// Get the full path to the file. Return an error if it doesn't exist.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Errorf("%#v not found", path)
	}
	fullPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// If the user didn't supply a name to init, then generate a name.
	if fileName == "" {
		name, err := pathToName(fullPath)
		if err != nil {
			return err
		}
		fileName = name
	}

	if err := d.setup(); err != nil {
		return err
	}
	if err := d.get(); err != nil {
		return err
	}

	// Replace the full path with a relative path.
	re := regexp.MustCompile(d.home)
	path = re.ReplaceAllString(fullPath, "~")

	d.files[fileName] = trackedFile{
		Path: path,
	}

	if err := d.save(); err != nil {
		return err
	}
	fmt.Printf("Initialized %s as %s\n", path, fileName)
	return nil
}

// Commit hashes and saves the current state of a tracked file.
func Commit(fileName string, message string) error {
	d := &data{}

	if err := d.get(); err != nil {
		return err
	}
	file, ok := d.files[fileName]
	if !ok {
		return notTrackedErr
	}

	f, err := os.Open(file.getFullPath(d.home))
	if err != nil {
		return err
	}
	defer f.Close()

	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	hash := fmt.Sprintf("%x", sha1.Sum(bytes))
	fmt.Println(hash)

	return nil
}

// GetPath gets the full path for a tracked file.
func GetPath(fileName string) (string, error) {
	d := &data{}
	if err := d.get(); err != nil {
		return "", err
	}

	file, ok := d.files[fileName]
	if !ok {
		return "", notTrackedErr
	}
	return file.getFullPath(d.home), nil
}

// Creates a name from the path of the file.
// Does this by stripping leading dots and file extensions.
// Examples: ~/.vimrc: vimrc
//           ~/.config/i3/config: config
//           ~/.config/alacritty/alacritty.yml: alacritty
func pathToName(path string) (string, error) {
	re := regexp.MustCompile(`(\w+)(\.\w+)?$`)
	matches := re.FindStringSubmatch(path)
	if len(matches) < 1 {
		return "", fmt.Errorf("failed to get name from %#v", path)
	}
	return matches[1], nil
}
