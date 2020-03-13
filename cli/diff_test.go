package cli

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func fakeDiffCommand(command string, args ...string) *exec.Cmd {
	assert.Equal(sneakyTestingReference, diffCmd, command)
	cs := []string{"-test.run=TestEditHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestDiff(t *testing.T) {
	clearTestStorage()
	initTestFile(t)

	sneakyTestingReference = t

	execCommand = fakeDiffCommand
	defer func() { execCommand = exec.Command }()

	diffCommand := &diffCommand{
		getStorage: getTestStorageClosure(),
	}

	t.Run("returns error when file is not tracked", func(t *testing.T) {
		diffCommand.fileName = notTrackedFile
		assert.Error(t, diffCommand.run(nil))
	})

	t.Run("ok", func(t *testing.T) {
		diffCommand.fileName = trackedFileAlias
		assert.NoError(t, diffCommand.run(nil))
	})
}
