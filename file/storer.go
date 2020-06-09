package file

import (
	"bytes"
	"fmt"
	"time"

	"github.com/knoebber/dotfile/usererr"
	"github.com/sergi/go-diff/diffmatchpatch"
)

const (
	diffCmd              = "diff"
	diffType             = "-u" // unified view
	colorOption          = "--color"
	initialCommitMessage = "Initial commit"
)

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

	uncompressed, err := Uncompress(contents)
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
		return usererr.Invalid(fmt.Sprintf("Commit %#v already exists", hash))
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

// Diff runs a diff on the revision at hash1 against the revision at hash2.
// If hash2 is empty, compares the current contents of the file.
func Diff(s Storer, hash1, hash2 string) ([]diffmatchpatch.Diff, error) {
	var text1, text2 string

	revision1, err := UncompressRevision(s, hash1)
	if err != nil {
		return nil, err
	}

	text1 = revision1.String()

	if hash2 == "" {
		contents, err := s.GetContents()
		if err != nil {
			return nil, err
		}
		text2 = string(contents)
	} else {
		revision2, err := UncompressRevision(s, hash2)
		if err != nil {
			return nil, err
		}
		text2 = revision2.String()
	}

	diffs := diffmatchpatch.New().DiffMain(text1, text2, false)

	for _, diff := range diffs {
		if diff.Type == diffmatchpatch.DiffInsert ||
			diff.Type == diffmatchpatch.DiffDelete {
			return diffs, nil
		}
	}

	return nil, usererr.Invalid("No changes")
}
