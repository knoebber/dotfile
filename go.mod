module github.com/knoebber/dotfile

go 1.12

require (
	github.com/knoebber/dotfile/cli/commands v0.0.0
	github.com/pkg/errors v0.8.1 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
)

replace github.com/knoebber/dotfile/cli/commands v0.0.0 => ./cli/commands
