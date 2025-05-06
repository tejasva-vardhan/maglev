package main

import (
	"testing"
)

func TestBlankKeyIsInvalid(t *testing.T) {
	app := &application{
		config: config{
			apiKeys: []string{"key"},
		},
	}
	result := app.isInvalidAPIKey("")

	if result == false {
		t.Error("isInvalidAPIKey('') = false")
	}
}
