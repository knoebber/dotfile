package local

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/knoebber/dotfile/dotfile"
	"github.com/stretchr/testify/assert"
)

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
		_ = ioutil.WriteFile(s.jsonPath(), []byte("invalid json"), 0644)
		assert.Error(t, s.SetTrackingData())
	})
}

func TestStorage_save(t *testing.T) {
	t.Run("error when directory does not exist", func(t *testing.T) {
		s := &Storage{Dir: "/not/exist"}
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
		_ = ioutil.WriteFile(filepath.Join(testDir, testAlias, testHash), []byte(testContent), 0644)
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
		err := s.Revert(bytes.NewBuffer([]byte(updatedTestContent)), testUpdatedHash)
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

		err := s.SaveCommit(bytes.NewBuffer([]byte(updatedTestContent)), c)

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

	t.Run("ok", func(t *testing.T) {
		s := testStorage()
		s.FileData = &dotfile.TrackingData{Path: "~/relative-path"}

		path, err := s.Path()
		assert.NoError(t, err)
		assert.NotEmpty(t, path)
	})
}
