package main

import (
	"testing"

	"github.com/ludo-technologies/pyscn/internal/version"
)

func TestVersion(t *testing.T) {
	// Version package should provide version info
	if version.Short() == "" {
		t.Error("version should not be empty")
	}

	// In dev mode, version should be "dev"
	if version.Short() != "dev" && version.Short() != "unknown" {
		// Version has been set via ldflags
		t.Logf("Version is set to: %s", version.Short())
	}
}
