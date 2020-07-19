// Package file provides functions and intefaces for dotfile operations.
package file

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/knoebber/dotfile/usererror"
	"github.com/pkg/errors"
)

var (
	// Captures the last part of the path without its file ending.
	pathToAliasRegex = regexp.MustCompile(`(\w+)(\.\w+)?$`)

	// Alias must only be words.
	validAliasRegex = regexp.MustCompile(`^\w+$`)

	// Must start in ~/ or /, cannot end in /
	validPathRegex = regexp.MustCompile(`^~?/.+[^/]$`)

	// ErrNoChanges is returned when a diff operation finds no changes.
	ErrNoChanges = usererror.Invalid("No changes")
)

// TrackingData is the data that dotfile uses to track files.
type TrackingData struct {
	Path     string   `json:"path"`
	Revision string   `json:"revision"`
	Commits  []Commit `json:"commits"`
}

// Commit represents a file revision.
type Commit struct {
	Hash      string `json:"hash"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"` // Unix timestamp.
}

// MapCommits maps hashes to commits.
func (td *TrackingData) MapCommits() map[string]*Commit {
	result := make(map[string]*Commit)
	for _, commit := range td.Commits {
		c := commit // commit always has the same address.
		if _, ok := result[c.Hash]; ok {
			continue
		}

		result[c.Hash] = &c
	}

	return result
}

// MergeTrackingData merges the new data into old.
// Returns the merged data and a slice of the hashes that are new.
func MergeTrackingData(old, new *TrackingData) (merged *TrackingData, newHashes []string, err error) {
	if old.Path != new.Path && old.Path != "" {
		err = fmt.Errorf("merging tracking data: old path %#v does not match new %#v", old.Path, new.Path)
		return
	}

	merged = &TrackingData{
		Path:     new.Path,
		Revision: new.Revision,
		Commits:  old.Commits,
	}

	newHashes = []string{}

	oldMap := make(map[string]bool)
	for _, c := range old.Commits {
		oldMap[c.Hash] = true
	}

	for _, r := range new.Commits {
		if _, ok := oldMap[r.Hash]; ok {
			// Old already has the new hash.
			continue
		}

		// Add the new hash.
		newHashes = append(newHashes, r.Hash)
		merged.Commits = append(merged.Commits, r)
	}

	sort.Slice(merged.Commits, func(i, j int) bool {
		return merged.Commits[i].Timestamp < merged.Commits[j].Timestamp
	})
	return
}

// GetAlias creates an alias when the passed in alias is empty.
// It works by removing leading dots and file extensions from the path.
// Examples: ~/.vimrc: vimrc
//           ~/.config/i3/config: config
//           ~/.config/alacritty/alacritty.yml: alacritty
func GetAlias(alias, path string) (string, error) {
	if alias != "" {
		return alias, nil
	}

	matches := pathToAliasRegex.FindStringSubmatch(path)
	if len(matches) < 2 {
		return "", fmt.Errorf("creating alias for %#v", path)
	}
	return matches[1], nil
}

// CheckAlias checks whether the alias format is allowed.
func CheckAlias(alias string) error {
	if !validAliasRegex.Match([]byte(alias)) {
		return usererror.Invalid(fmt.Sprintf("%#v has non word characters", alias))
	}

	return nil
}

// CheckPath checks whether the alias format is allowed.
func CheckPath(path string) error {
	if !validPathRegex.Match([]byte(path)) {
		return usererror.Invalid(fmt.Sprintf("%#v is not a valid file path", path))
	}

	return nil
}

func hashAndCompress(contents []byte) (*bytes.Buffer, string, error) {
	hash := fmt.Sprintf("%x", sha1.Sum(contents))
	compressed, err := Compress(contents)

	if err != nil {
		return nil, "", err
	}

	return compressed, hash, nil
}

// Compress compresses bytes with zlib.
func Compress(uncompressed []byte) (*bytes.Buffer, error) {
	compressed := new(bytes.Buffer)
	w := zlib.NewWriter(compressed)
	defer w.Close()

	if _, err := w.Write(uncompressed); err != nil {
		return nil, errors.Wrap(err, "compressing content")
	}

	return compressed, nil
}

// Uncompress uncompresses bytes.
// Bytes are expected to be zlib compressed.
func Uncompress(compressed []byte) (*bytes.Buffer, error) {
	uncompressed := new(bytes.Buffer)

	r, err := zlib.NewReader(bytes.NewBuffer(compressed))
	if err != nil {
		return nil, errors.Wrap(err, "uncompressing commit revision")
	}
	defer r.Close()

	if _, err = io.Copy(uncompressed, r); err != nil {
		return nil, errors.Wrap(err, "copying commits uncompressed data")
	}

	return uncompressed, nil
}

// ShortenEqualText splits text into newlines and discards everything in the middle.
// Used for removing equal text in a diff that is not near any changes.
//
// When there are less than 4 lines returns text unchanged.
// Otherwise takes the first/last two lines and discards the rest.
func ShortenEqualText(text string) string {
	lines := strings.Split(text, "\n")
	if len(lines) <= 3 {
		return text
	}

	return strings.Join(lines[:2], "\n") + "\n" + strings.Join(lines[len(lines)-2:], "\n")
}

// ShortenHash shortens a hash to a more friendly size.
func ShortenHash(hash string) string {
	return hash[0:7]
}
