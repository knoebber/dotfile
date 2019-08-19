package file

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	s := &Storage{}

	t.Run("error when home is empty", func(t *testing.T) {
		assert.Error(t, s.Setup("", testDir, testName))
	})

	t.Run("error when dir is empty", func(t *testing.T) {
		assert.Error(t, s.Setup("home", "", testName))
	})

	t.Run("error when name is empty", func(t *testing.T) {
		assert.Error(t, s.Setup("home", testDir, ""))
	})

}

func TestGetTrackedFile(t *testing.T) {
	clearTestStorage()
	s := getTestStorage()

	t.Run("returns error when file is not tracked", func(t *testing.T) {
		f, err := s.getTrackedFile(testAlias)
		fmt.Println(err)
		assert.Error(t, err)
		assert.Nil(t, f)
	})

	t.Run("ok when file is tracked", func(t *testing.T) {
		initTestFile(t, s)
		f, err := s.getTrackedFile(testAlias)
		assert.NoError(t, err)
		assert.NotNil(t, f)
	})
}
