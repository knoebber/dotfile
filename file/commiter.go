package file

import (
	"bytes"
	"fmt"
	"time"

	"github.com/knoebber/dotfile/usererror"
)

// TODO change init to take an initial commit message.
// If the file is created on server, make message like
// "Initial commit on https://dotfilehub.com"
// Currently there is ambiguity when pulling a file that has two initial commits.
const initialCommitMessage = "Initial commit"

// Commiter is the interace that wraps methods needed for saving commits.
type Commiter interface {
	Getter
	SaveCommit(buff *bytes.Buffer, c *Commit) error
}

// Init creates a new commit with the initial commit message.
func Init(c Commiter, path, alias string) error {
	if err := CheckPath(path); err != nil {
		return err
	}

	if err := CheckAlias(alias); err != nil {
		return err
	}

	return NewCommit(c, initialCommitMessage)
}

// NewCommit saves a revision of the file at its current state.
func NewCommit(c Commiter, message string) error {
	contents, err := c.Content()
	if err != nil {
		return err
	}

	compressed, hash, err := hashAndCompress(contents)
	if err != nil {
		return err
	}

	exists, err := c.HasCommit(hash)
	if err != nil {
		return err
	}
	if exists {
		return usererror.Invalid(fmt.Sprintf("Commit %#v already exists", hash))
	}

	newCommit := &Commit{
		Hash:      hash,
		Message:   message,
		Timestamp: time.Now().Unix(),
	}

	if err := c.SaveCommit(compressed, newCommit); err != nil {
		return err
	}

	return nil
}
