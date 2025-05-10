package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBlankKeyIsInvalid(t *testing.T) {
	app := &Application{
		config: config{
			apiKeys: []string{"key"},
		},
	}
	assert.True(t, app.isInvalidAPIKey(""))
}
