package file

import (
	"bytes"
	"fmt"
	"io"

	"github.com/knoebber/dotfile/usererror"
)

// Reverter is the interface that wraps methods needed for reverting a tracked file.
type Reverter interface {
	io.Closer
	Getter
	Revert(buff *bytes.Buffer, hash string) (err error)
}

// Checkout reverts a tracked file to its state at hash.
// Closes r on success.
func Checkout(r Reverter, hash string) error {
	exists, err := r.HasCommit(hash)
	if err != nil {
		return err
	}
	if !exists {
		return usererror.Invalid(fmt.Sprintf("Revision %#v not found", hash))
	}

	uncompressed, err := UncompressRevision(r, hash)
	if err != nil {
		return err
	}

	if err := r.Revert(uncompressed, hash); err != nil {
		return err
	}

	return r.Close()
}
