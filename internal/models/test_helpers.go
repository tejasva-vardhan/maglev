package models

import (
	"path/filepath"
	"testing"
)

// GetFixturePath returns the absolute path to a fixture file in the "testdata" directory relative to the project's root.
func GetFixturePath(t *testing.T, fixturePath string) string {
	t.Helper()

	absPath, err := filepath.Abs(filepath.Join("..", "..", "testdata", fixturePath))
	if err != nil {
		t.Fatalf("Failed to get absolute path to testdata/%s: %v", fixturePath, err)
	}

	return absPath
}
