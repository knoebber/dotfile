package local

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/knoebber/dotfile/file"
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

func TestStorage_Close(t *testing.T) {
	t.Run("error when directory does not exist", func(t *testing.T) {
		s := &Storage{Dir: "/not/exist"}
		assert.Error(t, s.Close())
	})
}

func TestStorage_HasCommit(t *testing.T) {
	s := &Storage{
		FileData: &file.TrackingData{
			Commits: []file.Commit{{
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

func TestStorage_GetRevision(t *testing.T) {
	s := setupTestFile(t)

	t.Run("error when revision does not exist", func(t *testing.T) {
		_, err := s.GetRevision("")
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		s := setupTestFile(t)
		_ = os.Mkdir(filepath.Join(testDir, testAlias), 0755)
		_ = ioutil.WriteFile(filepath.Join(testDir, testAlias, testHash), []byte(testContent), 0644)
		contents, err := s.GetRevision(testHash)
		assert.NoError(t, err)
		assert.NotEmpty(t, contents)
	})
}

func TestStorage_Revert(t *testing.T) {
	t.Run("error when unable to write", func(t *testing.T) {
		s := &Storage{FileData: &file.TrackingData{Path: "/not/exists"}}
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
		err := s.SaveCommit(new(bytes.Buffer), new(file.Commit))
		assert.Error(t, err)
	})
	t.Run("error when unable to write commit", func(t *testing.T) {
		s := testStorage()
		s.Dir = "/not/exist"
		s.FileData = new(file.TrackingData)
		err := s.SaveCommit(new(bytes.Buffer), new(file.Commit))
		assert.Error(t, err)
	})

	t.Run("error when unable to create commit file", func(t *testing.T) {
		s := setupTestFile(t)
		s.Alias = "/not/exists"
		err := s.SaveCommit(new(bytes.Buffer), new(file.Commit))
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		s := setupTestFile(t)
		timestamp := time.Now()
		c := &file.Commit{
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

func TestStorage_GetPath(t *testing.T) {
	t.Run("error when filedata is nil", func(t *testing.T) {
		s := testStorage()
		_, err := s.GetPath()
		assert.Error(t, err)
	})
	t.Run("error when path is empty", func(t *testing.T) {
		s := testStorage()
		s.FileData = new(file.TrackingData)
		_, err := s.GetPath()
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		s := testStorage()
		s.FileData = &file.TrackingData{Path: "~/relative-path"}

		path, err := s.GetPath()
		assert.NoError(t, err)
		assert.NotEmpty(t, path)
	})
}
