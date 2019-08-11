package file

// Dotfile tracks files by writing to a json file in the users home direcory.
// This file provides functions for reading and writing data to the file system.
//
// Example generated json:
// {
//   "bashrc": {
//     "path": "~/.bashrc"
//     "current": "451de414632e08c0ca3adf7a473b15f37c1b2e60"
//     "commits": [{
//       "hash":"451de414632e08c0ca3adf7a473b15f37c1b2e60",
//       "timestamp":"1558896245",
// 	 "message": "add aliases for dotfile"
//    }],
//  },
//   "emacs": {
//     "path": "~/.emacs.d/init.el",
//     "commits": [{
//       "hash":"8f94c7720a648af9cf9dab33e7f297d28b8bf7cd",
//       "timestamp":"1558896290",
//       "message": "add bindings for dotfile"
//  }]
// }

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

// Storage provides methods for reading and writing json data and compressed commit bytes.
// All exported fields should be set.
type Storage struct {
	Home string // The path to the users home directory.
	Dir  string // The path to the folder where data will be stored.
	Name string // The name of the json file.

	path  string
	files map[string]*trackedFile
}

func (d *Storage) saveCommit(c *commit, alias string, t *trackedFile, bytes []byte) error {
	// Create the directory for the files commits if it doesn't exist
	commitDir := fmt.Sprintf("%s%s", d.Dir, alias)
	_, err := os.Stat(commitDir)

	//
	// TODO these blocks are similar to setup(); pull logic out into a function.
	//

	if os.IsNotExist(err) {
		os.Mkdir(commitDir, 0755)
		fmt.Printf("Created %s\n", commitDir)
	} else if err != nil {
		return err
	}

	// The name of the file will be the hash
	commitPath := fmt.Sprintf("%s/%s", commitDir, c.Hash)
	_, err = os.Stat(commitPath)
	if os.IsNotExist(err) {
		f, createErr := os.Create(commitPath)
		f.Close()

		if createErr != nil {
			fmt.Printf("create err, %s\n", createErr)
			return createErr
		}
		fmt.Printf("Created %s\n", commitPath)

		if err := ioutil.WriteFile(commitPath, bytes, 0644); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return d.save(alias, t)
}

// Reads the json and makes the tracked file map.
func (d *Storage) get() error {
	if err := d.setPath(); err != nil {
		return err
	}

	d.files = make(map[string]*trackedFile)

	f, err := os.Open(d.path)
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

// Gets a tracked file by its alias.
func (d *Storage) getTrackedFile(alias string) (*trackedFile, error) {
	if err := d.get(); err != nil {
		return nil, err
	}

	t, ok := d.files[alias]
	if !ok {
		errors.New("file not tracked, use 'dot init <file>' first")
	}
	return t, nil
}

// Saves the trackedFile map to json.
func (d *Storage) save(alias string, t *trackedFile) error {
	d.files[alias] = t

	json, err := json.MarshalIndent(d.files, "", " ")
	if err != nil {
		return err
	}

	if err := d.setPath(); err != nil {
		return err
	}

	return ioutil.WriteFile(d.path, json, 0644)
}

// Sets up the dotfile directory if it hasn't been done yet.
func (d *Storage) setup() error {
	// Create the directory if it doesn't exist.
	_, err := os.Stat(d.Dir)
	if os.IsNotExist(err) {
		os.Mkdir(d.Dir, 0755)
		fmt.Printf("Created %#v\n", d.Dir)
	} else if err != nil {
		return err
	}

	if err := d.setPath(); err != nil {
		return err
	}

	// Create the data file if it doesn't exist.
	_, err = os.Stat(d.path)
	if os.IsNotExist(err) {
		f, createErr := os.Create(d.path)
		f.Close()

		if createErr != nil {
			fmt.Printf("create err, %s\n", createErr)
			return createErr
		}
		fmt.Printf("Created %#v\n", d.path)
	} else if err != nil {
		return err
	}
	return d.get()
}

func (d *Storage) setPath() error {
	if d.Home == "" {
		return errors.New("home not set")
	}
	if d.Dir == "" {
		return errors.New("dir not set")
	}
	if d.Name == "" {
		return errors.New("name not set")
	}

	if d.Dir[len(d.Dir)-1] != '/' {
		d.Dir += "/"
	}

	d.path = fmt.Sprintf("%s%s", d.Dir, d.Name)
	return nil
}
