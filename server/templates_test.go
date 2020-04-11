package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLink_GetClass(t *testing.T) {
	l := &Link{Active: true}
	assert.NotEmpty(t, l.GetClass())
}
