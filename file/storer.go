package file

import (
	"bytes"
	"fmt"
	"time"
)

const initialCommitMessage = "Initial commit"

// Storer is an interface that encapsulates the I/O that is required for dotfile.
type Storer interface {
	HasCommit(hash string) (exists bool, err error)
	GetContents() (contents []byte, err error)
	GetRevision(hash string) (revision []byte, err error)
	SaveCommit(buff *bytes.Buffer, hash, message string, timestamp time.Time) error
	Revert(buff *bytes.Buffer, hash string) (err error)
}

// UncompressRevision reads a revision and uncompresses it.
// Returns the uncompressed bytes of alias at hash.
func UncompressRevision(s Storer, hash string) (*bytes.Buffer, error) {
	contents, err := s.GetRevision(hash)
	if err != nil {
		return nil, err
	}

	uncompressed, err := uncompress(contents)
	if err != nil {
		return nil, err
	}

	return uncompressed, nil
}

// Init creates a new commit with the initial commit message.
func Init(s Storer, alias string) error {
	if err := CheckAlias(alias); err != nil {
		return err
	}

	return NewCommit(s, initialCommitMessage)
}

// NewCommit saves a revision of the file at its current state.
func NewCommit(s Storer, message string) error {
	contents, err := s.GetContents()
	if err != nil {
		return err
	}

	compressed, hash, err := hashAndCompress(contents)
	if err != nil {
		return err
	}

	exists, err := s.HasCommit(hash)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("commit %#v already exists", hash)
	}

	return s.SaveCommit(compressed, hash, message, time.Now())
}

// Checkout reverts a tracked file to its state at hash.
func Checkout(s Storer, hash string) error {
	exists, err := s.HasCommit(hash)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("revision %#v not found", hash)
	}

	uncompressed, err := UncompressRevision(s, hash)
	if err != nil {
		return err
	}

	if err := s.Revert(uncompressed, hash); err != nil {
		return err
	}

	return nil
}
