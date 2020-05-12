package file

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"regexp"

	"github.com/pkg/errors"
)

var pathToAliasRegex = regexp.MustCompile(`(\w+)(\.\w+)?$`)

type NotTrackedError struct {
	alias string
}

func (e *NotTrackedError) Error() string {
	return fmt.Sprintf("%#v is not tracked", e.alias)
}

// NotTracked returns a new NotTrackedError
func ErrNotTracked(alias string) error {
	return &NotTrackedError{alias}
}

// PathToAlias creates an alias from the path of the file.
// Works by removing leading dots and file extensions.
// Examples: ~/.vimrc: vimrc
//           ~/.config/i3/config: config
//           ~/.config/alacritty/alacritty.yml: alacritty
func PathToAlias(path string) (string, error) {
	matches := pathToAliasRegex.FindStringSubmatch(path)
	if len(matches) < 2 {
		return "", fmt.Errorf("creating alias for %#v", path)
	}
	return matches[1], nil
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

func uncompress(compressed []byte) (*bytes.Buffer, error) {
	uncompressed := new(bytes.Buffer)

	r, err := zlib.NewReader(bytes.NewBuffer(compressed))
	if err != nil {
		return nil, errors.Wrap(err, "uncompressing commit revision")
	}
	defer r.Close()

	if _, err = io.Copy(uncompressed, r); err != nil {
		return nil, errors.Wrap(err, "copying uncompressed data")
	}

	return uncompressed, nil
}
