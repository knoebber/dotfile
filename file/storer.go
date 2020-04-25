package file

import (
	"bytes"
	"fmt"
	"regexp"
	"time"
)

const initialCommitMessage = "Initial commit"

var pathToAliasRegex = regexp.MustCompile(`(\w+)(\.\w+)?$`)

// Storer is an interface that encapsulates the I/O that is required for dotfile.
type Storer interface {
	GetContents(path string) (contents []byte, err error)
	GetTracked(alias string) (tf *Tracked, err error)
	SaveTracked(tf *Tracked) (err error)
	GetRevision(alias, hash string) (contents []byte, err error)
	SaveRevision(tf *Tracked, buff *bytes.Buffer, hash string) (err error)
	Revert(contents []byte, path string) (err error)
}

// Creates an alias from the path of the file.
// Works by removing leading dots and file extensions.
// Examples: ~/.vimrc: vimrc
//           ~/.config/i3/config: config
//           ~/.config/alacritty/alacritty.yml: alacritty
func pathToAlias(path string) (string, error) {
	matches := pathToAliasRegex.FindStringSubmatch(path)
	if len(matches) < 2 {
		return "", fmt.Errorf("failed to get alias from %#v", path)
	}
	return matches[1], nil
}

func newRevision(s Storer, tf *Tracked, message string) error {
	contents, err := s.GetContents(tf.RelativePath)
	if err != nil {
		return err
	}

	compressed, hash, err := hashAndCompress(contents)
	if err != nil {
		return err
	}

	tf.Revision = hash
	tf.Commits = append(tf.Commits, Commit{
		Hash:      hash,
		Message:   message,
		Timestamp: time.Now().Unix(),
	})

	return s.SaveRevision(tf, compressed, hash)
}

// MustGetTracked attempts to find alias.
// Returns an error when it doesn't exist.
func MustGetTracked(s Storer, alias string) (*Tracked, error) {
	tf, err := s.GetTracked(alias)
	if err != nil {
		return nil, err
	}

	if tf == nil {
		return nil, fmt.Errorf("%#v is not tracked", alias)
	}

	return tf, nil
}

// UncompressRevision gets and uncompressions a revision.
// Returns the uncompressed bytes of alias at hash.
func UncompressRevision(s Storer, alias, hash string) (*bytes.Buffer, error) {
	contents, err := s.GetRevision(alias, hash)
	if err != nil {
		return nil, err
	}

	uncompressed, err := uncompress(contents)
	if err != nil {
		return nil, err
	}

	return uncompressed, nil
}

// Init initializes a file for dotfile to track.
// It creates a tracked file with an initial commit.
func Init(s Storer, relativePath, fileName string) (alias string, err error) {
	var tf *Tracked

	if fileName == "" {
		alias, err = pathToAlias(relativePath)
		if err != nil {
			return
		}
	} else {
		alias = fileName
	}

	tf, err = s.GetTracked(alias)
	if err != nil {
		return
	}

	if tf != nil {
		err = fmt.Errorf("%#v is tracking %s", alias, tf.RelativePath)
		return
	}

	tf = &Tracked{
		RelativePath: relativePath,
		Alias:        alias,
		Commits:      []Commit{},
	}

	if err := newRevision(s, tf, initialCommitMessage); err != nil {
		return "", err
	}

	return alias, nil
}

// NewCommit saves a revision of the file at its current state.
func NewCommit(s Storer, alias, message string) error {
	tf, err := MustGetTracked(s, alias)
	if err != nil {
		return err
	}

	return newRevision(s, tf, message)
}

// Checkout reverts a tracked file to its state at hash.
func Checkout(s Storer, alias, hash string) error {
	tf, err := MustGetTracked(s, alias)
	if err != nil {
		return err
	}

	// Checkout to latest version by default.
	if hash == "" {
		hash = tf.Revision
	}

	found := false
	for _, c := range tf.Commits {
		if c.Hash == hash {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("revision %#v not found", hash)
	}

	uncompressed, err := UncompressRevision(s, alias, hash)
	if err != nil {
		return err
	}

	if err := s.Revert(uncompressed.Bytes(), tf.RelativePath); err != nil {
		return err
	}

	tf.Revision = hash

	return s.SaveTracked(tf)
}
