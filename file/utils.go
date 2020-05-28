package file

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"regexp"

	"github.com/knoebber/dotfile/usererr"
	"github.com/pkg/errors"
)

var pathToAliasRegex = regexp.MustCompile(`(\w+)(\.\w+)?$`)
var validAliasRegex = regexp.MustCompile(`^\w+$`)

// NotTrackedError is returned when a file is not tracked.
// TODO moved to usererr
type NotTrackedError struct {
	alias string
}

func (e *NotTrackedError) Error() string {
	return fmt.Sprintf("%#v is not tracked", e.alias)
}

// ErrNotTracked returns a new NotTrackedError
func ErrNotTracked(alias string) error {
	return &NotTrackedError{alias}
}

// GetAlias creates an alias when the passed in alias is empty.
// It works by removing leading dots and file extensions from the path.
// Examples: ~/.vimrc: vimrc
//           ~/.config/i3/config: config
//           ~/.config/alacritty/alacritty.yml: alacritty
func GetAlias(alias, path string) (string, error) {
	if alias != "" {
		return alias, nil
	}

	matches := pathToAliasRegex.FindStringSubmatch(path)
	if len(matches) < 2 {
		return "", fmt.Errorf("creating alias for %#v", path)
	}
	return matches[1], nil
}

// CheckAlias checks whether the alias format is allowed.
func CheckAlias(alias string) error {
	if !validAliasRegex.Match([]byte(alias)) {
		return usererr.Invalid(fmt.Sprintf("%#v has non word characters", alias))
	}

	return nil
}

func hashAndCompress(contents []byte) (*bytes.Buffer, string, error) {
	hash := fmt.Sprintf("%x", sha1.Sum(contents))

	compressed := new(bytes.Buffer)
	w := zlib.NewWriter(compressed)
	defer w.Close()

	if _, err := w.Write(contents); err != nil {
		return nil, "", errors.Wrap(err, "compressing file for commit")
	}

	return compressed, hash, nil
}

// Uncompress uncompresses bytes.
// Bytes are expected to be zlib compressed.
func Uncompress(compressed []byte) (*bytes.Buffer, error) {
	uncompressed := new(bytes.Buffer)

	r, err := zlib.NewReader(bytes.NewBuffer(compressed))
	if err != nil {
		return nil, errors.Wrap(err, "uncompressing commit revision")
	}
	defer r.Close()

	if _, err = io.Copy(uncompressed, r); err != nil {
		return nil, errors.Wrap(err, "copying commits uncompressed data")
	}

	return uncompressed, nil
}
