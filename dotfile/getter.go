package dotfile

import (
	"bytes"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
)

// Getter is an interface that wraps methods for reading tracked files.
type Getter interface {
	DirtyContent() (contents []byte, err error)        // Uncommitted changes.
	Revision(hash string) (revision []byte, err error) // Revision at hash.
}

// UncompressRevision reads a revision and uncompresses it.
// Returns the uncompressed bytes of alias at hash.
func UncompressRevision(g Getter, hash string) (*bytes.Buffer, error) {
	contents, err := g.Revision(hash)
	if err != nil {
		return nil, err
	}

	return Uncompress(contents)
}

// IsClean returns whether the dirty content matches the hash.
// Returns true when there is no dirty content.
func IsClean(g Getter, hash string) (bool, error) {
	contents, err := g.DirtyContent()
	if err != nil {
		return false, err
	}
	if contents == nil {
		return true, nil
	}

	return hash == hashContent(contents), nil
}

// Runs a diff on the revision at hash1 against the revision at hash2.
// If hash2 is empty, compares the dirty content of the file.
// Returns an usererror when there is no difference.
func Diff(g Getter, hash1, hash2 string) (*gotextdiff.Unified, error) {
	var text1, text2 string

	revision1, err := UncompressRevision(g, hash1)
	if err != nil {
		return nil, err
	}

	text1 = revision1.String()

	if hash2 == "" {
		contents, err := g.DirtyContent()
		if err != nil {
			return nil, err
		}
		text2 = string(contents)
	} else {
		revision2, err := UncompressRevision(g, hash2)
		if err != nil {
			return nil, err
		}
		text2 = revision2.String()
	}

	// package gotextdiff is a copy from the internal gopls implementation, so it's not a perfect fit for this usecase.
	edits := myers.ComputeEdits("", text1, text2)
	diff := gotextdiff.ToUnified("", "", text1, edits)
	if len(diff.Hunks) == 0 {
		return nil, ErrNoChanges
	}
	return &diff, nil
}
