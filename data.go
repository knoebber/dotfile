package dotfile

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

const (
	dotfileDir string = ".dotfile"
	dotfile    string = "files.json"
)

// Dotfile tracks files by writing to a json file in the users home direcory.
// This file provides functions for saving and retreiving json data.
//
// Example generated json:
// {
//   "bashrc": {
//     "path": "/home/nicolas/.bashrc"
//     "current": "451de414632e08c0ca3adf7a473b15f37c1b2e60"
//     "commits": [{
//       "hash":"451de414632e08c0ca3adf7a473b15f37c1b2e60",
//       "timestamp":"1558896245",
// 	 "message": "add aliases for dotfile"
//    }],
//  },
//   "emacs": {
//     "path": "/home/nicolas/.emacs.d/init.el",
//     "commits": [{
//       "hash":"8f94c7720a648af9cf9dab33e7f297d28b8bf7cd",
//       "timestamp":"1558896290",
//       "message": "add bindings for dotfile"
//  }]
// }

type data struct {
	files map[string]trackedFile
	home  string
}

// Gets the current data.
func (d *data) get() error {
	if d.home == "" {
		if err := d.getHome(); err != nil {
			return err
		}
	}

	d.files = make(map[string]trackedFile)

	f, err := os.Open(d.filePath())
	if err != nil {
		return err
	}
	defer f.Close()

	// Read the entire file into bytes so it can be unmarshalled.
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	if len(bytes) == 0 {
		return nil
	}

	if err := json.Unmarshal(bytes, &d.files); err != nil {
		return err
	}
	return nil
}

func (d *data) save() error {
	path := d.filePath()
	json, err := json.MarshalIndent(d.files, "", " ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, json, 0644)
}

// Sets up the dotfile directory if it hasn't been done yet.
func (d *data) setup() error {
	if err := d.getHome(); err != nil {
		return err
	}

	// Create the directory if it doesn't exist.
	dir := d.dirPath()
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
		fmt.Printf("Created %#v\n", dir)
	} else if err != nil {
		return err
	}

	// Create the data file if it doesn't exist.
	path := d.filePath()
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		f, createErr := os.Create(path)
		f.Close()
		if createErr != nil {
			fmt.Printf("create err, %s\n", createErr)
			return createErr
		}
		fmt.Printf("Created %#v\n", path)
	} else if err != nil {
		return err
	}
	return nil
}

func (d *data) dirPath() string {
	return fmt.Sprintf("%s/%s/", d.home, dotfileDir)
}

func (d *data) filePath() string {
	return fmt.Sprintf("%s%s", d.dirPath(), dotfile)
}

// Gets the users home directory
func (d *data) getHome() (err error) {
	d.home, err = os.UserHomeDir()
	return
}
