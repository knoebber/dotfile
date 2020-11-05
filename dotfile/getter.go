package dotfile

import (
	"bytes"
	"html"
	"html/template"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
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

	uncompressed, err := Uncompress(contents)
	if err != nil {
		return nil, err
	}

	return uncompressed, nil
}

// IsClean returns whether the dirty content matches the hash.
func IsClean(g Getter, hash string) (bool, error) {
	contents, err := g.DirtyContent()
	if err != nil {
		return false, err
	}

	return hash == hashContent(contents), nil
}

// Diff runs a diff on the revision at hash1 against the revision at hash2.
// If hash2 is empty, compares the dirty content of the file.
// Returns an usererror when there is no difference.
func Diff(g Getter, hash1, hash2 string) ([]diffmatchpatch.Diff, error) {
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

// DiffPrettyText is based on diffmatchpatch.DiffPrettyText.
// It returns colorized text.
func DiffPrettyText(g Getter, hash1, hash2 string) (string, error) {
	var buff strings.Builder

	diffs, err := Diff(g, hash1, hash2)
	if err != nil {
		return "", err
	}

	for _, diff := range diffs {
		text := diff.Text

		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			_, _ = buff.WriteString("\x1b[32m")
			_, _ = buff.WriteString(text)
			_, _ = buff.WriteString("\x1b[0m")
		case diffmatchpatch.DiffDelete:
			_, _ = buff.WriteString("\x1b[31m")
			_, _ = buff.WriteString(text)
			_, _ = buff.WriteString("\x1b[0m")
		case diffmatchpatch.DiffEqual:
			_, _ = buff.WriteString(text)
		}
	}

	return buff.String(), nil
}

// DiffPrettyHTML is based on diffmatchpatch.DiffPrettyHTML.
// It returns HTML that is ready to be added to a template.
func DiffPrettyHTML(g Getter, hash1, hash2 string) (template.HTML, error) {
	var buff strings.Builder

	diffs, err := Diff(g, hash1, hash2)
	if err != nil {
		return "", err
	}

	for _, diff := range diffs {

		text := html.EscapeString(diff.Text)
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			_, _ = buff.WriteString("<ins>")
			_, _ = buff.WriteString(text)
			_, _ = buff.WriteString("</ins>")
		case diffmatchpatch.DiffDelete:
			_, _ = buff.WriteString("<del>")
			_, _ = buff.WriteString(text)
			_, _ = buff.WriteString("</del>")
		case diffmatchpatch.DiffEqual:
			_, _ = buff.WriteString("<span>")
			_, _ = buff.WriteString(text)
			_, _ = buff.WriteString("</span>")
		}
	}
	return template.HTML(buff.String()), nil
}
