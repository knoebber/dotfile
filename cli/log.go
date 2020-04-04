package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/knoebber/dotfile/file"
	"github.com/knoebber/dotfile/local"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	delimChar              = "="
	timestampDisplayFormat = "January 02, 2006 3:04 PM -0700"
)

type logCommand struct {
	getStorage func() (*local.Storage, error)
	fileName   string
}

func (l *logCommand) run(ctx *kingpin.ParseContext) error {
	s, err := l.getStorage()
	if err != nil {
		return err
	}

	tf, err := file.MustGetTracked(s, l.fileName)
	if err != nil {
		return err
	}

	delim := strings.Repeat(delimChar, len(tf.Revision))

	halfHeaderDelim := strings.Repeat(delimChar, (len(tf.Revision)-9)/2)
	currentDelim := halfHeaderDelim + " CURRENT " + halfHeaderDelim + delimChar
	for _, commit := range tf.Commits {
		timeStamp := time.Unix(commit.Timestamp, 0).Format(timestampDisplayFormat)

		fmt.Println("")
		if commit.Hash == tf.Revision {
			fmt.Println(currentDelim)
		} else {
			fmt.Println(delim)
		}

		fmt.Print(timeStamp + "\n")
		if commit.Message != "" {
			fmt.Print(commit.Message + "\n")
		}
		fmt.Print(commit.Hash)
		fmt.Printf("\n%s\n", delim)
	}
	return nil
}

func addLogSubCommandToApplication(app *kingpin.Application, gs func() (*local.Storage, error)) {
	lc := &logCommand{
		getStorage: gs,
	}
	c := app.Command("log", "shows revision history with commit hashes for a tracked file").Action(lc.run)
	c.Arg("file-name", "tracked file to show history for").Required().StringVar(&lc.fileName)
}
