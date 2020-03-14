package file

// Tracked is a file that dotfile is tracking.
type Tracked struct {
	RelativePath string   `json:"path"`     // The relative path to the file - something like '~/.vimrc'
	Revision     string   `json:"revision"` // The hash of the files current commit.
	Commits      []Commit `json:"commits"`  // The commits the file has.

	Alias string `json:"-"` // The alias of the file. For example, 'vim' might map to ~/.vimrc.
}
