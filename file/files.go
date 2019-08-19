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
	"strings"
	"time"

	"github.com/pkg/errors"
)

var pathToAliasRegex = regexp.MustCompile(`(\w+)(\.\w+)?$`)

// Init sets up a file for dotfile to track.
// Returns the alias for the newly tracked file.
func Init(s *Storage, filePath, altName string) (string, error) {
	var (
		alias string
		err   error
	)

	if _, err = os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("%#v not found", filePath)
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

	// Replace the full path with a relative path.
	relativePath := strings.Replace(fullPath, s.GetHome(), "~", 1)

	if err = s.save(alias, &trackedFile{
		Path: relativePath,
	}); err != nil {
		return "", err
	}
	return alias, nil
}

// Commit hashes and saves the current state of a tracked file.
func Commit(s *Storage, alias, message string) (string, error) {
	file, err := s.getTrackedFile(alias)
	if err != nil {
		return "", err
	}

	path := file.getFullPath(s.GetHome())
	f, err := os.Open(path)
	if err != nil {
		return "", errors.Wrapf(err, "failed to open %s", path)
	}
	defer f.Close()

	fileBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read %s", path)
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

	return hash, s.saveCommit(c, alias, file, compressed.Bytes())
}

// GetPath gets the full path for a tracked file.
func GetPath(s *Storage, alias string) (string, error) {
	file, err := s.getTrackedFile(alias)
	if err != nil {
		return "", err
	}

	return file.getFullPath(s.GetHome()), nil
}

// Creates a alias from the path of the file.
// Does this by stripping leading dots and file extensions.
// Examples: ~/.vimrc: vimrc
//           ~/.config/i3/config: config
//           ~/.config/alacritty/alacritty.yml: alacritty
func pathToAlias(path string) (string, error) {
	matches := pathToAliasRegex.FindStringSubmatch(path)
	if len(matches) < 2 {
		return "", fmt.Errorf("failed to get name from %#v", path)
	}
	return matches[1], nil
}
