package dotfile

import (
	"bytes"
	"time"

	"github.com/knoebber/usererror"
)

const initialCommitMessage = "Initial commit"

// Committer is an interface for saving saving commits.
type Committer interface {
	Getter
	HasCommit(hash string) (exists bool, err error)
	SaveCommit(buff *bytes.Buffer, c *Commit) error
}

// Init creates a new commit with the initial commit message.
func Init(c Committer, path, alias string) error {
	if err := CheckPath(path); err != nil {
		return err
	}

	if err := CheckAlias(alias); err != nil {
		return err
	}

	return NewCommit(c, initialCommitMessage)
}

// NewCommit saves a revision of the file at its current state.
func NewCommit(c Committer, message string) error {
	contents, err := c.DirtyContent()
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
		return usererror.Format("Commit %q already exists", hash)
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
