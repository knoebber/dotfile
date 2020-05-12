package cli

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	delimChar              = "="
	timestampDisplayFormat = "January 02, 2006 3:04 PM -0700"
)

type logCommand struct {
	fileName string
}

func (l *logCommand) run(ctx *kingpin.ParseContext) error {
	s, err := loadFile(l.fileName)
	if err != nil {
		return err
	}

	revision := s.Tracking.Revision
	delim := strings.Repeat(delimChar, len(revision))

	halfHeaderDelim := strings.Repeat(delimChar, (len(revision)-9)/2)
	currentDelim := halfHeaderDelim + " CURRENT " + halfHeaderDelim + delimChar
	for _, commit := range s.Tracking.Commits {
		timeStamp := time.Unix(commit.Timestamp, 0).Format(timestampDisplayFormat)

		fmt.Println("")
		if commit.Hash == revision {
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

func addLogSubCommandToApplication(app *kingpin.Application) {
	lc := new(logCommand)

	c := app.Command("log", "shows revision history with commit hashes for a tracked file").Action(lc.run)
	c.Arg("file-name", "tracked file to show history for").Required().StringVar(&lc.fileName)
}
