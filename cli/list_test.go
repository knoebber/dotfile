package cli

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestList(t *testing.T) {
	listCommand := new(listCommand)
	assert.NoError(t, listCommand.run(nil))
}
