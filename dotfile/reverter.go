package dotfile

import (
	"bytes"

	"github.com/knoebber/usererror"
)

// Reverter is the interface that wraps methods needed for reverting a tracked file.
type Reverter interface {
	Getter
	HasCommit(hash string) (exists bool, err error)
	Revert(buff *bytes.Buffer, hash string) (err error)
}

// Checkout reverts a tracked file to its state at hash.
func Checkout(r Reverter, hash string) error {
	exists, err := r.HasCommit(hash)
	if err != nil {
		return err
	}
	if !exists {
		return usererror.Format("Revision %q not found", hash)
	}

	uncompressed, err := UncompressRevision(r, hash)
	if err != nil {
		return err
	}

	if err := r.Revert(uncompressed, hash); err != nil {
		return err
	}

	return nil
}
