// Package dotfile provides functions and interfaces for dotfile operations.
package dotfile

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"github.com/knoebber/dotfile/usererror"
	"github.com/pkg/errors"
	"io"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	// Captures the last part of the path without its file ending.
	pathToAliasRegex = regexp.MustCompile(`([^/.]+)(\..*)?$`)

	// Alias must only contain letters, dashes, numbers, underscores
	validAliasRegex = regexp.MustCompile(`^[a-z0-9-_]+$`)

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

func hashContent(contents []byte) string {
	return fmt.Sprintf("%x", sha1.Sum(contents))
}

// MergeTrackingData merges the new data into old.
// Returns the merged data and a slice of the hashes that are new.
func MergeTrackingData(old, new *TrackingData) (merged *TrackingData, newHashes []string, err error) {
	if new == nil {
		return nil, nil, errors.New("new tracking data must be set")
	}

	if old == nil {
		old = &TrackingData{}
	} else if old.Path != new.Path {
		return nil, nil, fmt.Errorf("merging tracking data: old path %q does not match new %q", old.Path, new.Path)
	}

	merged = &TrackingData{
		Path:     new.Path,
		Revision: new.Revision,
		Commits:  old.Commits,
	}

	newHashes = []string{}

	oldMap := old.MapCommits()
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

// Alias creates an alias when the passed in alias is empty.
// It works by removing leading dots and file extensions from the path.
// Examples: ~/.vimrc: vimrc
//           ~/.config/i3/config: config
//           ~/.config/alacritty/alacritty.yml: alacritty
func Alias(alias, path string) (string, error) {
	alias = strings.ToLower(alias)
	path = strings.ToLower(path)

	if alias != "" {
		return alias, nil
	}

	matches := pathToAliasRegex.FindStringSubmatch(path)
	if len(matches) < 2 {
		return "", fmt.Errorf("creating alias for %q", path)
	}
	return matches[1], nil
}

// CheckAlias checks whether the alias is a valid format.
func CheckAlias(alias string) error {
	if !validAliasRegex.Match([]byte(alias)) {
		return usererror.Invalid(fmt.Sprintf("%q has non allowed characters", alias))
	}

	return nil
}

// CheckPath checks whether the path is a valid format.
func CheckPath(path string) error {
	l := len(path)
	if l == 0 {
		return usererror.Invalid("File path cannot be empty")
	}

	if path[l-1] == filepath.Separator {
		return usererror.Invalid("File path cannot be directory")
	}

	if len(path) > 2 && path[:2] == "~/" {
		return nil
	}

	if !filepath.IsAbs(path) {
		return usererror.Invalid("File path must start with ~/ or be absolute")
	}

	return nil
}

func hashAndCompress(contents []byte) (*bytes.Buffer, string, error) {
	compressed, err := Compress(contents)

	if err != nil {
		return nil, "", err
	}

	return compressed, hashContent(contents), nil
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

// ShortenHash shortens a hash to a more friendly size.
func ShortenHash(hash string) string {
	return hash[0:7]
}
