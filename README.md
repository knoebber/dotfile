## Dotfile [![Build Status](https://travis-ci.org/knoebber/dotfile.svg?branch=master)](https://travis-ci.org/knoebber/dotfile)

Dotfile is a simple version control system designed for single files.
It is currently under development.

## Working Features

[Asciinema Demo](https://asciinema.org/a/vEMt14MIf1Imlul8cpaDv9JXh?autoplay=1)

Dotfile commands are based on `git` but simplified. Checking a file in is simple:

```
dot init ~/.bashrc
```

This will create an inital commit with the current time stamp. Dotfile will store the path of the file and
give it a default alias of `bashrc`. After making changes to `.bashrc` use the alias to refer to it
regardless of the current directory. Alternatively provide a name when the file is initialized for clarity:

```
dot init ~/.config/i3/config i3
dot init ~/.emacs.d/init.el emacs
```

Open a tracked file in `$EDITOR` without having to type its path:

```
dot edit bashrc
```

View commit history

```
dot log bashrc
```

Check the diff after making changes:

```
dot diff bashrc # diff against last commit

dot diff bashrc 42d4220dda4d43d639a6b7ac76f2ff4e04b651a6
```

Reverting a file is easy:

```
dot checkout bashrc # reverts to the last commit

dot checkout bashrc 42d4220dda4d43d639a6b7ac76f2ff4e04b651a6
```


Unlike git, there is no need to stage files before making a commit:

```
dot commit bashrc # Uses the current timestamp as the commit message

dot commit emacs "Add dotfile bindings"
```

## Planned Features

Push changes:

```
dot push i3
```

Pull a file:

```
dot pull i3 # Installs as ~/.config/i3/config
```

Or pull everything on a new machine:

```
dot pull --all
```

View and make changes to your tracked files on dotfilehub.com
