package file

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// Init sets up a file for dotfile to track.
// Returns the alias for the newly tracked file.
func Init(d *Data, filePath, altName string) (string, error) {
	var (
		alias string
		err   error
	)

	if _, err = os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("\"%#v\" not found", filePath)
	}

	// Get the full path so that it can later turn it into a relative path.
	fullPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}

	alias = altName
	if altName == "" {
		alias, err = pathToAlias(fullPath)
		if err != nil {
			return "", err
		}
	}

	if err = d.setup(); err != nil {
		return "", err
	}

	// Replace the full path with a relative path.
	re := regexp.MustCompile(d.Home)
	relativePath := re.ReplaceAllString(fullPath, "~")

	if err = d.save(alias, &trackedFile{
		Path: relativePath,
	}); err != nil {
		return "", err
	}
	return alias, nil
}

// Commit hashes and saves the current state of a tracked file.
func Commit(d *Data, alias, message string) (string, error) {
	file, err := d.getTrackedFile(alias)
	if err != nil {
		return "", err
	}

	f, err := os.Open(file.getFullPath(d.Home))
	if err != nil {
		return "", err
	}
	defer f.Close()

	fileBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	hash := fmt.Sprintf("%x", sha1.Sum(fileBytes))

	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	w.Write(fileBytes)
	w.Close()

	file.Current = hash
	c := &commit{
		Hash:      hash,
		Message:   message,
		Timestamp: time.Now().Unix(),
	}

	file.Commits = append(file.Commits, c)

	return hash, d.saveCommit(c, alias, file, compressed.Bytes())
}

// GetPath gets the full path for a tracked file.
func GetPath(d *Data, alias string) (string, error) {
	file, err := d.getTrackedFile(alias)
	if err != nil {
		return "", err
	}

	return file.getFullPath(d.Home), nil
}

// Creates a alias from the path of the file.
// Does this by stripping leading dots and file extensions.
// Examples: ~/.vimrc: vimrc
//           ~/.config/i3/config: config
//           ~/.config/alacritty/alacritty.yml: alacritty
func pathToAlias(path string) (string, error) {
	re := regexp.MustCompile(`(\w+)(\.\w+)?$`)
	matches := re.FindStringSubmatch(path)
	if len(matches) < 1 {
		return "", fmt.Errorf("failed to get name from %#v", path)
	}
	return matches[1], nil
}
