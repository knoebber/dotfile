package file

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/knoebber/dotfile/usererr"
	"github.com/pkg/errors"
)

var (
	pathToAliasRegex = regexp.MustCompile(`(\w+)(\.\w+)?$`)
	validAliasRegex  = regexp.MustCompile(`^\w+$`)
	// Must start in ~/ or /, cannot end in /
	validPathRegex = regexp.MustCompile(`^~?/.+[^/]$`)

	// ErrNoChanges is returned when a diff operation finds no changes.
	ErrNoChanges = usererr.Invalid("No changes")
)

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

// CheckPath checks whether the alias format is allowed.
func CheckPath(path string) error {
	if !validPathRegex.Match([]byte(path)) {
		return usererr.Invalid(fmt.Sprintf("%#v is not a valid file path", path))
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

// ShortenEqualText splits text into newlines and discards everything in the middle.
// Used for removing equal text in a diff that is not near any changes.
//
// When there are less than 4 lines returns text unchanged.
// Otherwise takes the first/last two lines and discards the rest.
func ShortenEqualText(text string) string {
	lines := strings.Split(text, "\n")
	if len(lines) <= 3 {
		return text
	}

	return strings.Join(lines[:2], "\n") + "\n" + strings.Join(lines[len(lines)-2:], "\n")
}

// ShortenHash shortens a hash to a more friendly size.
func ShortenHash(hash string) string {
	return hash[0:7]
}
