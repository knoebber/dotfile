module github.com/knoebber/dotfile/cmd/dot

go 1.14

require (
	github.com/knoebber/dotfile v0.0.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
)

replace github.com/knoebber/dotfile v0.0.0 => ../..
