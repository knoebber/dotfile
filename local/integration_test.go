package local

import (
	"fmt"
	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/dotfile/dotfileclient"
	"github.com/knoebber/dotfile/server"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const (
	dotfilehubAddr     = ":26900"
	dotfilehubUsername = "user"
	dotfilehubPassword = "password12345"
)

func failIf(t *testing.T, err error, context ...string) {
	if err != nil {
		t.Fatalf("%s %v", err, context)
	}
}

func TestDotfilehubIntegration(t *testing.T) {
	failIf(t, os.Chdir("../server"), "changing directory so that assets work in integration test")

	setupTestFile(t)
	dotfilehub, err := server.New(server.Config{
		Addr: dotfilehubAddr,
	})
	failIf(t, err)

	defer dotfilehub.Close()

	go func() {
		if err := dotfilehub.ListenAndServe(); err != nil {
			// Expected after close.
			fmt.Printf("dotfilehub listen and serve: %s\n", err)
		}
	}()

	user, err := db.CreateUser(db.Connection, dotfilehubUsername, "", dotfilehubPassword)
	failIf(t, err)

	client := dotfileclient.New("http://"+dotfilehubAddr, dotfilehubUsername, user.CLIToken)

	s := testStorage()
	failIf(t, s.SetTrackingData())
	failIf(t, s.Push(client))

	convertedPath, err := convertPath(testTrackedFile)
	failIf(t, err, "converting path")

	temp := &db.TempFileRecord{
		UserID:  user.ID,
		Content: []byte(testUpdatedContent),
		Alias:   testAlias,
		Path:    convertedPath,
	}
	failIf(t, temp.Create(db.Connection), "creating temp file")
	failIf(t, db.InitOrCommit(user.ID, testAlias, "new content on server"), "committing to file on server")

	failIf(t, s.Pull(client))
	content, err := s.DirtyContent()
	failIf(t, err)

	assert.Equal(t, testUpdatedContent, string(content))
}
