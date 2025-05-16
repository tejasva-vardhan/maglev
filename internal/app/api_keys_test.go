package app

import (
	"github.com/stretchr/testify/assert"
	"maglev.onebusaway.org/internal/appconf"
	"testing"
)

func TestBlankKeyIsInvalid(t *testing.T) {
	app := &Application{
		Config: appconf.Config{
			ApiKeys: []string{"key"},
		},
	}
	assert.True(t, app.IsInvalidAPIKey(""))
}
