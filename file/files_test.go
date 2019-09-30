package file

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	clearTestStorage()
	s := getTestStorage()

	t.Run("init makes initial commit", func(t *testing.T) {
		initTestFile(t, s)
		s.get()
		file, _ := s.files[testAlias]
		assert.NotEmpty(t, file.Commits)
	})

	t.Run("ok when file is already tracked", func(t *testing.T) {
		initTestFile(t, s)
		initTestFile(t, s)
	})
}

func TestCommit(t *testing.T) {
	clearTestStorage()
	s := getTestStorage()

	t.Run("ok when files is unchanged", func(t *testing.T) {
		initTestFile(t, s)
		assert.NoError(t, Commit(s, testAlias, ""))
	})
}
