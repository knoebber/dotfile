package file

import (
	"fmt"
	"sort"
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

// MergeTrackingData merges the new data into old.
// Returns the merged data and a slice of hashes that the original data doesn't have.
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
