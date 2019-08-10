package file

import (
	"regexp"
)

type trackedFile struct {
	Path    string    `json:"path"`
	Current string    `json:"current"`
	Commits []*commit `json:"commits"`
}

type commit struct {
	Hash      string `json:"hash"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// Paths are stored as relative paths so that dotfile can work with different home directories.
// getFullPath returns the full path to a tracked file.
func (tf *trackedFile) getFullPath(home string) string {
	re := regexp.MustCompile("~")
	return re.ReplaceAllString(tf.Path, home)
}
