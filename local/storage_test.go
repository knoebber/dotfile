package local

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadFile(t *testing.T) {
	t.Run("error when home is empty", func(t *testing.T) {
		_, err := LoadFile("", "", "")
		assert.Error(t, err)
	})

	t.Run("error when dir is empty", func(t *testing.T) {
		_, err := LoadFile(testHome, "", "")
		assert.Error(t, err)
	})

	t.Run("error when name is empty", func(t *testing.T) {
		_, err := LoadFile(testHome, testDir, "")
		assert.Error(t, err)
	})

	t.Run("error on failure to create storage files", func(t *testing.T) {
		clearTestStorage()
		_, err := LoadFile(testHome, testDir+testDir, testAlias)
		assert.Error(t, err)
	})

	t.Run("error on failure to get", func(t *testing.T) {
		clearTestStorage()
		_ = os.Mkdir(testDir, 0755)
		_ = ioutil.WriteFile(filepath.Join(testDir, testAlias), []byte("invalid json"), 0644)
		_, err := LoadFile(testHome, testDir, testAlias)
		assert.Error(t, err)
	})
}

func TestStorage_get(t *testing.T) {
	t.Run("error when path does not exist", func(t *testing.T) {
		s := &Storage{jsonPath: "/not/exist"}
		assert.Error(t, s.get())
	})

	t.Run("error when no content", func(t *testing.T) {
		s := &Storage{jsonPath: testTrackedFile}
		writeTestFile(t, []byte{})

		assert.Error(t, s.get())
	})

	t.Run("ok", func(t *testing.T) {
		s := setupTestFile(t)
		assert.NoError(t, s.get())
	})
}

func TestStorage_save(t *testing.T) {
	t.Run("error when json file does not exist", func(t *testing.T) {
		s := &Storage{jsonPath: "/not/exist"}
		assert.Error(t, s.save())
	})
}

func TestStorage_HasCommit(t *testing.T) {
	s := &Storage{
		Tracking: &TrackedFile{
			Commits: []Commit{{
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
		s := &Storage{Tracking: &TrackedFile{Path: "/not/exists"}}
		assert.Error(t, s.Revert(new(bytes.Buffer), testHash))
	})

	t.Run("ok", func(t *testing.T) {
		s := setupTestFile(t)
		err := s.Revert(bytes.NewBuffer([]byte(updatedTestContent)), testUpdatedHash)
		assert.NoError(t, err)
		assert.Equal(t, testUpdatedHash, s.Tracking.Revision)
	})
}

func TestStorage_SaveCommit(t *testing.T) {

	t.Run("error when unable to create commit directory", func(t *testing.T) {
		s := &Storage{dir: "/not/exist", Tracking: new(TrackedFile)}
		err := s.SaveCommit(new(bytes.Buffer), testHash, "", time.Now())
		assert.Error(t, err)
	})

	t.Run("error when unable to create commit file", func(t *testing.T) {
		s := setupTestFile(t)
		s.Alias = "/not/exists"
		err := s.SaveCommit(new(bytes.Buffer), testHash, "", time.Now())
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		s := setupTestFile(t)
		timestamp := time.Now()
		err := s.SaveCommit(
			bytes.NewBuffer([]byte(updatedTestContent)),
			testUpdatedHash,
			testMessage,
			timestamp,
		)

		assert.NoError(t, err)
		assert.Equal(t, testUpdatedHash, s.Tracking.Revision)
		assert.Equal(t, testMessage, s.Tracking.Commits[1].Message)
		assert.Equal(t, timestamp.Unix(), s.Tracking.Commits[1].Timestamp)
	})
}
