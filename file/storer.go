package file

import (
	"bytes"
	"fmt"
	"regexp"
)

const initialCommitMessage = "Initial commit"

var pathToAliasRegex = regexp.MustCompile(`(\w+)(\.\w+)?$`)

// Storer is an interface that encapsulates the I/O that is required for dotfile.
type Storer interface {
	GetContents(string) ([]byte, error)
	GetTracked(string) (*Tracked, error)
	SaveTracked(*Tracked) error
	GetRevision(string, string) ([]byte, error)
	SaveRevision(*Tracked, *Commit) error
	Revert([]byte, string) error
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
	var (
		tf     *Tracked
		commit *Commit
	)

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

	contents, err := s.GetContents(relativePath)
	if err != nil {
		return
	}

	commit, err = newCommit(contents, initialCommitMessage)
	if err != nil {
		return
	}

	tf = &Tracked{
		RelativePath: relativePath,
		Revision:     commit.Hash,
		Commits:      []Commit{*commit},
		Alias:        alias,
	}

	if err = s.SaveRevision(tf, commit); err != nil {
		return
	}

	return
}

// NewCommit saves a revision of the file at its current state.
func NewCommit(s Storer, alias, message string) error {
	tf, err := MustGetTracked(s, alias)
	if err != nil {
		return err
	}

	contents, err := s.GetContents(tf.RelativePath)
	if err != nil {
		return err
	}

	commit, err := newCommit(contents, message)
	if err != nil {
		return err
	}

	tf.Commits = append(tf.Commits, *commit)
	tf.Revision = commit.Hash

	return s.SaveRevision(tf, commit)
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

	if err := s.Revert(uncompressed.Bytes(), tf.RelativePath); err != nil {
		return err
	}

	tf.Revision = hash

	return s.SaveTracked(tf)
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
