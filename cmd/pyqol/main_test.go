package main

import (
	"testing"
)

func TestVersion(t *testing.T) {
	if version == "" {
		t.Error("version should not be empty")
	}
	
	expected := "0.0.1-alpha"
	if version != expected {
		t.Errorf("version = %v, want %v", version, expected)
	}
}