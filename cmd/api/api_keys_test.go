package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBlankKeyIsInvalid(t *testing.T) {
	app := &Application{
		config: Config{
			apiKeys: []string{"key"},
		},
	}
	assert.True(t, app.isInvalidAPIKey(""))
}
