package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const invalidAddr = "example.com"

func TestStartServer(t *testing.T) {
	assert.Panics(t, func() { Start(invalidAddr) })
}
