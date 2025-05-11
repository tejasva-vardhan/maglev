package app

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBlankKeyIsInvalid(t *testing.T) {
	app := &Application{
		Config: Config{
			ApiKeys: []string{"key"},
		},
	}
	assert.True(t, app.IsInvalidAPIKey(""))
}
