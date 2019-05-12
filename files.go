package dotfile

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

const (
	dotfileDir = ".dotfile"
	dotfile    = "files.json"
)

type trackedFile struct {
	Path    string   `json:"path"`
	Commits []string `json:"commits"`
}

// Init sets up a file for dotfile to track.
func Init(path string, fileName string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("%s not found", path)
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	if fileName == "" {
		var err error
		fileName, err = pathToName(path)
		if err != nil {
			return err
		}
	}

	// Replace the full path with a relative path
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	re := regexp.MustCompile(home)
	path = re.ReplaceAllString(path, "~")

	d, err := getData(home)
	if err != nil {
		return err
	}

	d[fileName] = trackedFile{
		Path:    path,
		Commits: []string{},
	}

	if err := writeData(d, home); err != nil {
		return err
	}

	fmt.Printf("Initialized %s as %s\n", path, fileName)
	return nil
}

// GetPath gets the full path for a tracked file.
func GetPath(fileName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	d, err := getData(home)
	if err != nil {
		return "", err
	}

	// In case the user passes a path or includes the files extension.
	name, err := pathToName(fileName)
	if err != nil {
		return "", err
	}

	file, ok := d[name]
	if !ok {
		return "", fmt.Errorf("%s is not tracked", name)
	}

	// Replace the relative path with the full path.
	re := regexp.MustCompile("~")
	path := re.ReplaceAllString(file.Path, home)

	return path, nil
}

func getData(home string) (map[string]trackedFile, error) {
	var d map[string]trackedFile
	d = make(map[string]trackedFile)
	// Create the directory if it doesn't exist.
	dir := fmt.Sprintf("%s/%s/", home, dotfileDir)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
		fmt.Printf("Created %s\n", dir)
	}

	// Create the data file if it doesn't exist.
	path := fmt.Sprintf("%s/%s", dir, dotfile)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		f, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		f.Close()
		return d, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Read the entire file into bytes so it can be unmarshalled.
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &d); err != nil {
		return nil, err
	}

	return d, nil
}

func writeData(d map[string]trackedFile, home string) error {
	path := fmt.Sprintf("%s/%s/%s", home, dotfileDir, dotfile)
	json, err := json.MarshalIndent(d, "", " ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(path, json, 0644); err != nil {
		return err
	}
	return nil
}

func pathToName(path string) (string, error) {
	// Create a name from the path of the file.
	// Examples: ~/.vimrc: vimrc
	//           ~/.config/i3/config: config
	//           ~/.config/alacritty/alacritty.yml: alacritty
	re := regexp.MustCompile(`(\w+)(\.\w+)?$`)
	matches := re.FindStringSubmatch(path)
	if len(matches) < 1 {
		return "", fmt.Errorf("failed to get name from %s", path)
	}
	return matches[1], nil
}
