package file

import (
	"bytes"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// Getter is an interface that wraps methods for reading tracked files.
type Getter interface {
	GetContents() (contents []byte, err error)
	GetRevision(hash string) (revision []byte, err error)
	HasCommit(hash string) (exists bool, err error)
}

// UncompressRevision reads a revision and uncompresses it.
// Returns the uncompressed bytes of alias at hash.
func UncompressRevision(g Getter, hash string) (*bytes.Buffer, error) {
	contents, err := g.GetRevision(hash)
	if err != nil {
		return nil, err
	}

	uncompressed, err := Uncompress(contents)
	if err != nil {
		return nil, err
	}

	return uncompressed, nil
}

// IsClean returns whether the contents of g matches hash.
func IsClean(g Getter, hash string) (bool, error) {
	contents, err := g.GetContents()
	if err != nil {
		return false, err
	}

	return hash == hashContent(contents), nil
}

// Diff runs a diff on the revision at hash1 against the revision at hash2.
// If hash2 is empty, compares the current contents of the file.
// Returns an usererror when there is no difference.
func Diff(g Getter, hash1, hash2 string) ([]diffmatchpatch.Diff, error) {
	var text1, text2 string

	revision1, err := UncompressRevision(g, hash1)
	if err != nil {
		return nil, err
	}

	text1 = revision1.String()

	if hash2 == "" {
		contents, err := g.GetContents()
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

	dmp := diffmatchpatch.New()

	diffs := dmp.DiffCleanupSemantic(dmp.DiffMain(text1, text2, false))

	for _, diff := range diffs {
		if diff.Type == diffmatchpatch.DiffInsert ||
			diff.Type == diffmatchpatch.DiffDelete {
			return diffs, nil
		}
	}

	return nil, ErrNoChanges
}
