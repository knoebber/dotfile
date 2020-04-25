package local

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/knoebber/dotfile/file"
	"github.com/stretchr/testify/assert"
)

func TestNewStorage(t *testing.T) {
	t.Run("error when home is empty", func(t *testing.T) {
		_, err := NewStorage("", "", "")
		assert.Error(t, err)
	})

	t.Run("error when dir is empty", func(t *testing.T) {
		_, err := NewStorage(testHome, "", "")
		assert.Error(t, err)
	})

	t.Run("error when name is empty", func(t *testing.T) {
		_, err := NewStorage(testHome, testDir, "")
		assert.Error(t, err)
	})

	t.Run("error on failure to create storage files", func(t *testing.T) {
		clearTestStorage()
		_, err := NewStorage(testHome, testDir+testDir, testFiles)
		assert.Error(t, err)
	})

	t.Run("error on failure to get", func(t *testing.T) {
		clearTestStorage()
		_ = os.Mkdir(testDir, 0755)
		_ = ioutil.WriteFile(filepath.Join(testDir, testFiles), []byte("invalid json"), 0644)
		_, err := NewStorage(testHome, testDir, testFiles)
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		clearTestStorage()
		// _ = ioutil.WriteFile(filepath.Join(testDir, testFiles), []byte{}, 0644)
		_, err := NewStorage(testHome, testDir, testFiles)
		assert.NoError(t, err)
	})
}

func TestStorage_get(t *testing.T) {
	t.Run("error when path does not exist", func(t *testing.T) {
		s := &Storage{path: "/not/exist"}
		assert.Error(t, s.get())
	})

	t.Run("ok when file is empty", func(t *testing.T) {
		s := setupTestStorage()
		assert.NoError(t, s.get())
	})
}

func TestStorage_save(t *testing.T) {
	t.Run("error when files.json does not exist", func(t *testing.T) {
		s := &Storage{path: "/not/exist"}
		assert.Error(t, s.save())
	})
}

func TestGetRevision(t *testing.T) {
	s := setupTestStorage()

	t.Run("error when revision does not exist", func(t *testing.T) {
		_, err := s.GetRevision("", "")
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		_ = os.Mkdir(filepath.Join(testDir, testAlias), 0755)
		_ = ioutil.WriteFile(filepath.Join(testDir, testAlias, testHash), []byte(testContent), 0644)
		contents, err := s.GetRevision(testAlias, testHash)
		assert.NoError(t, err)
		assert.NotEmpty(t, contents)
	})
}

func TestGetTracked(t *testing.T) {
	s := setupTestStorage()

	t.Run("returns nil nil when not exist", func(t *testing.T) {
		tf, err := s.GetTracked(testAlias)

		assert.Nil(t, tf)
		assert.NoError(t, err)
	})

	t.Run("returns tracked file with alias set", func(t *testing.T) {
		s.files[testAlias] = new(file.Tracked)

		tf, err := s.GetTracked(testAlias)
		assert.NotNil(t, tf)
		assert.Equal(t, tf.Alias, testAlias)
		assert.NoError(t, err)
	})
}

func TestSaveRevision(t *testing.T) {

	t.Run("error when unable to create commit directory", func(t *testing.T) {
		s := &Storage{path: "/not/exist"}
		err := s.SaveRevision(new(file.Tracked), nil, "")
		assert.Error(t, err)
	})

	s := setupTestStorage()
	t.Run("error when unable to create commit file", func(t *testing.T) {
		err := s.SaveRevision(new(file.Tracked), nil, "")
		assert.Error(t, err)
	})

	t.Run("error when revision already exists", func(t *testing.T) {
		_ = os.Mkdir(filepath.Join(testDir, testAlias), 0755)
		_ = ioutil.WriteFile(filepath.Join(testDir, testAlias, testHash), []byte(testContent), 0644)
		err := s.SaveRevision(&file.Tracked{Alias: testAlias}, nil, testHash)
		assert.Error(t, err)
	})

}
