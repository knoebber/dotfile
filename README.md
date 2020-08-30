## Dotfile [![Build Status](https://travis-ci.org/knoebber/dotfile.svg?branch=master)](https://travis-ci.org/knoebber/dotfile)

Dotfile is a simple version control system designed for single files.

Dotfile is comprised of two main programs:

`dot` - manage files locally
[docs](docs/cli.org)

`dotfilehub` - manage files remotely
[docs](docs/web.org)

[dotfilehub.com](https://dotfilehub.com)

## Example CLI Usage
Dotfile commands are based on `git` but simplified.

Check a file in:

```
dot init ~/.bashrc
```

This creates an initial commit. Dotfile will store the path of the file and
give it a default alias of `bashrc`. Use the alias to refer to it regardless of the current directory.

Open in `$EDITOR`

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

```

Commit a file's changes:

```
dot commit bashrc "Add dotfile alias"
```

Revert a file to its last commit:

```
dot checkout bashrc

```

Push changes to a dotfilehub server:

```
dot push bashrc
```

Install all of your dotfiles

```
dot pull --all
```