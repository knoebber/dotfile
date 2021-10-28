package local

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/knoebber/dotfile/dotfile"
	"github.com/knoebber/dotfile/dotfileclient"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestStorage_JSON(t *testing.T) {
	t.Run("not tracked error", func(t *testing.T) {
		s := testStorage()
		s.Dir = "/does/not/exist"
		_, err := s.JSON()
		assert.True(t, errors.Is(err, ErrNotTracked))
	})
}

func TestStorage_SetTrackingData(t *testing.T) {
	t.Run("error when storage dir is empty", func(t *testing.T) {
		s := testStorage()
		s.Dir = ""
		assert.Error(t, s.SetTrackingData())
	})

	t.Run("error when storage dir does not exist", func(t *testing.T) {
		s := testStorage()
		s.Dir = "/not/exist"
		assert.Error(t, s.SetTrackingData())
	})

	t.Run("error when alias is empty", func(t *testing.T) {
		s := testStorage()
		s.Alias = ""
		assert.Error(t, s.SetTrackingData())
	})

	t.Run("error when file is not tracked", func(t *testing.T) {
		clearTestStorage()
		s := testStorage()
		assert.Error(t, s.SetTrackingData())
	})

	t.Run("error on invalid json", func(t *testing.T) {
		s := testStorage()
		_ = os.Mkdir(testDir, 0755)
		_ = os.WriteFile(s.jsonPath(), []byte("invalid json"), 0644)
		assert.Error(t, s.SetTrackingData())
	})
}

func TestStorage_save(t *testing.T) {
	t.Run("error when directory doesn't exist", func(t *testing.T) {
		s := &Storage{Dir: "/not/exist"}
		assert.Error(t, s.save())
	})
	t.Run("error when json path doesn't exist", func(t *testing.T) {
		resetTestStorage(t)
		s := testStorage()
		s.Alias = "not/tracked"
		assert.Error(t, s.save())
	})
}

func TestStorage_HasCommit(t *testing.T) {
	s := &Storage{
		FileData: &dotfile.TrackingData{
			Commits: []dotfile.Commit{{
				Hash: "a",
			}},
		}}

	t.Run("returns true", func(t *testing.T) {
		res, _ := s.HasCommit("a")
		assert.True(t, res)
	})
	t.Run("returns false", func(t *testing.T) {
		res, _ := s.HasCommit("b")
		assert.False(t, res)
	})
}

func TestStorage_Revision(t *testing.T) {
	s := setupTestFile(t)

	t.Run("error when revision does not exist", func(t *testing.T) {
		_, err := s.Revision("")
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		s := setupTestFile(t)
		_ = os.Mkdir(filepath.Join(testDir, testAlias), 0755)
		_ = os.WriteFile(filepath.Join(testDir, testAlias, testHash), []byte(testContent), 0644)
		contents, err := s.Revision(testHash)
		assert.NoError(t, err)
		assert.NotEmpty(t, contents)
	})
}

func TestStorage_Revert(t *testing.T) {
	t.Run("error when unable to write", func(t *testing.T) {
		s := &Storage{FileData: &dotfile.TrackingData{Path: "/not/exists"}}
		assert.Error(t, s.Revert(new(bytes.Buffer), testHash))
	})

	t.Run("ok", func(t *testing.T) {
		s := setupTestFile(t)
		err := s.Revert(bytes.NewBuffer([]byte(testUpdatedContent)), testUpdatedHash)
		assert.NoError(t, err)
		assert.Equal(t, testUpdatedHash, s.FileData.Revision)
	})
}

func TestStorage_SaveCommit(t *testing.T) {
	t.Run("error when tracking data is not set", func(t *testing.T) {
		s := testStorage()
		err := s.SaveCommit(new(bytes.Buffer), new(dotfile.Commit))
		assert.Error(t, err)
	})
	t.Run("error when unable to write commit", func(t *testing.T) {
		s := testStorage()
		s.Dir = "/not/exist"
		s.FileData = new(dotfile.TrackingData)
		err := s.SaveCommit(new(bytes.Buffer), new(dotfile.Commit))
		assert.Error(t, err)
	})

	t.Run("error when unable to create commit file", func(t *testing.T) {
		s := setupTestFile(t)
		s.Alias = "/not/exists"
		err := s.SaveCommit(new(bytes.Buffer), new(dotfile.Commit))
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		s := setupTestFile(t)
		timestamp := time.Now()
		c := &dotfile.Commit{
			Hash:      testUpdatedHash,
			Timestamp: time.Now().Unix(),
			Message:   testMessage,
		}

		err := s.SaveCommit(bytes.NewBuffer([]byte(testUpdatedContent)), c)

		assert.NoError(t, err)
		assert.Equal(t, testUpdatedHash, s.FileData.Revision)
		assert.Equal(t, testMessage, s.FileData.Commits[1].Message)
		assert.Equal(t, timestamp.Unix(), s.FileData.Commits[1].Timestamp)
	})
}

func TestStorage_Path(t *testing.T) {
	t.Run("error when filedata is nil", func(t *testing.T) {
		s := testStorage()
		_, err := s.Path()
		assert.Error(t, err)
	})
	t.Run("error when path is empty", func(t *testing.T) {
		s := testStorage()
		s.FileData = new(dotfile.TrackingData)
		_, err := s.Path()
		assert.Error(t, err)
	})

	t.Run("error when $HOME not set", func(t *testing.T) {
		defer os.Setenv("HOME", os.Getenv("HOME"))
		_ = os.Unsetenv("HOME")

		s := testStorage()
		s.FileData = &dotfile.TrackingData{Path: "~/relative-path"}

		_, err := s.Path()
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		s := testStorage()
		s.FileData = &dotfile.TrackingData{Path: "~/relative-path"}

		path, err := s.Path()
		assert.NoError(t, err)
		assert.NotEmpty(t, path)
	})
}

func TestStorage_Push(t *testing.T) {
	t.Run("error when data not loaded", func(t *testing.T) {
		s := testStorage()
		assert.Error(t, s.Push(nil))
	})
	t.Run("error when client fails to connect", func(t *testing.T) {
		client := new(dotfileclient.Client)
		s := testStorage()
		assert.Error(t, s.Push(client))
	})
}

func TestStorage_Pull(t *testing.T) {
	s := testStorage()
	client := new(dotfileclient.Client)
	client.Client = http.DefaultClient
	t.Run("error on attempt to load invalid json", func(t *testing.T) {
		setupTestFile(t)

		if err := os.WriteFile(testDir+testAlias+".json", []byte("invalid json"), 0644); err != nil {
			t.Fatalf("writing test json")
		}
		assert.Error(t, s.Pull(client))
	})

	t.Run("uncommitted changes", func(t *testing.T) {
		setupTestFile(t)
		updateTestFile(t)
		err := s.Pull(client)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "uncommitted")
	})

	t.Run("client error", func(t *testing.T) {
		setupTestFile(t)
		err := s.Pull(client)
		assert.Error(t, err)
	})
}

func TestStorage_Move(t *testing.T) {
	s := testStorage()
	setupTestFile(t)

	t.Run("error when no data", func(t *testing.T) {
		assert.Error(t, s.Move(testTrackedFile, false))
	})

	assert.NoError(t, s.SetTrackingData())

	t.Run("parent dirs works", func(t *testing.T) {
		nested := testDir + "new/file.txt"
		assert.Error(t, s.Move(nested, false))
		assert.NoError(t, s.Move(nested, true))
	})

	t.Run("error when path can't be made", func(t *testing.T) {
		assert.Error(t, s.Move("/not/real.txt", true))
	})
}

func TestStorage_Rename(t *testing.T) {
	s := testStorage()
	setupTestFile(t)

	t.Run("error when alias exists", func(t *testing.T) {
		assert.Error(t, s.Rename(testAlias))
	})

	t.Run("error when alias has invalid format", func(t *testing.T) {
		assert.Error(t, s.Rename("invalid/alias"))
	})

	t.Run("error when dir doesn't exist", func(t *testing.T) {
		invalidStorage := &Storage{
			Alias: testAlias,
			Dir:   "/does/not/exist",
		}

		assert.Error(t, invalidStorage.Rename("newalias"))
	})

	t.Run("ok", func(t *testing.T) {
		assert.NoError(t, s.Rename("new"))
	})
}

func TestStorage_Forget(t *testing.T) {
	t.Run("error when dir doesn't exist", func(t *testing.T) {
		invalidStorage := &Storage{Dir: "/does/not/exist"}
		assert.Error(t, invalidStorage.Forget())
	})

	t.Run("ok", func(t *testing.T) {
		setupTestFile(t)
		s := testStorage()
		assert.NoError(t, s.Forget())
	})
}

func TestStorage_RemoveCommits(t *testing.T) {
	setupTestFile(t)
	s := testStorage()

	t.Run("error when file data not set", func(t *testing.T) {
		assert.Error(t, s.RemoveCommits())
	})

	t.Run("ok with no commits", func(t *testing.T) {
		s.FileData = new(dotfile.TrackingData)
		assert.NoError(t, s.RemoveCommits())
	})

	t.Run("removes non current commit", func(t *testing.T) {
		updateTestFile(t)
		assert.NoError(t, s.SetTrackingData())
		assert.NoError(t, dotfile.NewCommit(s, "testing remove commits"))
		assert.NoError(t, s.RemoveCommits())
		assert.Equal(t, 1, len(s.FileData.Commits))
	})
}

func TestStorage_Remove(t *testing.T) {
	setupTestFile(t)
	s := testStorage()

	t.Run("error when tracking data not loaded", func(t *testing.T) {
		assert.Error(t, s.Remove())
	})

	t.Run("ok", func(t *testing.T) {
		assert.NoError(t, s.SetTrackingData())
		assert.NoError(t, s.Remove())
	})
}
