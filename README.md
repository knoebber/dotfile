## Dotfile

Dotfile is a simple version control system designed for single files.
It is currently under development and not usable.

## Planned Functionality

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

Check the diff after making changes:

```
dot diff bashrc
```

Reverting a file is easy:

```
dot checkout bashrc # reverts to the last commit

dot log bashrc
<commit hash> | <commit message - timestamp>

dot checkout bashrc <commit hash> 
```


Unlike git, there is no need to stage files before making a commit:

```
dot commit bashrc # Uses the current timestamp as the commit message
dot commit emacs "Add dotfile bindings"
```

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
