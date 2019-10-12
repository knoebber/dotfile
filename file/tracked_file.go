package file

type Tracked struct {
	RelativePath string   `json:"path"`
	Revision     string   `json:"revision"`
	Commits      []Commit `json:"commits"`

	Alias string `json:"-"`
}
