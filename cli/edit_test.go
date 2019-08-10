package cli

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

const arbitraryEditor = "nano"

// Based on https://npf.io/2015/06/testing-exec-command/

var sneakyTestingReference *testing.T

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	assert.Equal(sneakyTestingReference, arbitraryEditor, command)
	cs := []string{"-test.run=TestEditHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestEditCommandLaunchesEditor(t *testing.T) {
	sneakyTestingReference = t

	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	defer os.Setenv("EDITOR", os.Getenv("EDITOR"))
	os.Setenv("EDITOR", arbitraryEditor)

	editCommand := &editCommand{
		fileName: testAlias,
		data:     getTestData(),
	}

	clearTestData()

	t.Run("error before init", func(t *testing.T) {
		err := editCommand.run(nil)
		assert.Error(t, err)
	})

	t.Run("no error after init", func(t *testing.T) {
		initTestFile(t)
		err := editCommand.run(nil)
		assert.NoError(t, err)
	})
}

func TestErrorIfEditorNotSet(t *testing.T) {
	defer os.Setenv("EDITOR", os.Getenv("EDITOR"))
	os.Unsetenv("EDITOR")

	command := &editCommand{
		fileName: testAlias,
	}
	err := command.run(nil)
	assert.Equal(t, ErrEditorEnvVarNotSet, err)
}

func TestEditHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	assert.Equal(t, testAlias, os.Args[1])
	os.Exit(0)
}
