package file

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

const (
	timestampDisplayFormat = "January 02, 2006 3:04 PM -0700"
	initialCommitMessage   = "Initial commit"
)

// TODO GetTracked: return an error or not? make consistent, see checkout vs init.

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

// Init initializes a file for dotfile to track.
// It creates a tracked file with an initial commit.
func Init(s Storer, relativePath, alias string) (err error) {
	if alias == "" {
		alias, err = pathToAlias(relativePath)
		if err != nil {
			return err
		}
	}

	tf, err := s.GetTracked(alias)
	if err != nil {
		return err
	}

	if tf != nil {
		err = fmt.Errorf("%#v is tracking %s", alias, tf.RelativePath)
		return
	}

	contents, err := s.GetContents(relativePath)
	if err != nil {
		return
	}

	commit, err := newCommit(contents, initialCommitMessage)
	if err != nil {
		return err
	}

	tf = &Tracked{
		RelativePath: relativePath,
		Revision:     commit.Hash,
		Commits:      []Commit{*commit},
		Alias:        alias,
	}

	if err = s.SaveRevision(tf, commit); err != nil {
		return err
	}

	log.Printf("Initialized %s as %#v", relativePath, alias)
	return nil
}

func NewCommit(s Storer, alias, message string) error {
	tf, err := s.GetTracked(alias)
	if err != nil {
		return err
	}

	if tf == nil {
		return fmt.Errorf("%#v is not tracked", alias)
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
	return s.SaveRevision(tf, commit)
}

func Checkout(s Storer, alias, hash string) error {
	tf, err := s.GetTracked(alias)
	if err != nil {
		return err
	}

	if tf == nil {
		return fmt.Errorf("%#v is not tracked", alias)
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

	contents, err := s.GetRevision(alias, hash)
	if err != nil {
		return err
	}

	uncompressed, err := uncompress(contents)
	if err != nil {
		return err
	}

	if err := s.Revert(uncompressed.Bytes(), tf.RelativePath); err != nil {
		return err
	}

	tf.Revision = hash

	return s.SaveTracked(tf)
}

func Log(s Storer, alias string) error {
	tf, err := s.GetTracked(alias)
	if err != nil {
		return err
	}

	delim := strings.Repeat("=", len(tf.Revision))

	for _, commit := range tf.Commits {
		timeStamp := time.Unix(commit.Timestamp, 0).Format(timestampDisplayFormat)
		log.Printf("\n%s", delim)
		log.Print(timeStamp)

		if commit.Message != "" {
			log.Print(commit.Message)
		}

		log.Print(commit.Hash)
		log.Print(delim)
	}
	return nil
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
