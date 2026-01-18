package config

import "testing"

type testConfig struct {
	Name string
	Port int
}

func TestIsZero(t *testing.T) {
	var empty testConfig
	nonEmpty := testConfig{Name: "chaos", Port: 8080}

	if !IsZero(empty) {
		t.Fatalf("expected empty config to be zero")
	}

	if IsZero(nonEmpty) {
		t.Fatalf("expected non-empty config to not be zero")
	}
}
