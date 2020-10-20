package server

import (
	"github.com/gorilla/mux"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupRoutes(t *testing.T) {
	setupTest(t, fileHandler())
	assert.NoError(t, setupRoutes(mux.NewRouter(), Config{}))
}
