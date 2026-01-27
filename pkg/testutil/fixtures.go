package testutil

import (
	"embed"
	"testing"
)

//go:embed fixtures/*
var fixturesFS embed.FS

// LoadFixture loads a fixture file by name
func LoadFixture(t *testing.T, name string) string {
	t.Helper()

	data, err := fixturesFS.ReadFile("fixtures/" + name)
	if err != nil {
		t.Fatalf("Failed to load fixture %s: %v", name, err)
	}
	return string(data)
}

// LoadFixtureBytes loads a fixture file as bytes
func LoadFixtureBytes(t *testing.T, name string) []byte {
	t.Helper()

	data, err := fixturesFS.ReadFile("fixtures/" + name)
	if err != nil {
		t.Fatalf("Failed to load fixture %s: %v", name, err)
	}
	return data
}
