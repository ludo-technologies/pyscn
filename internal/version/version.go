package version

import (
	"fmt"
	"runtime"
)

// These variables are set via ldflags during build
var (
	// Version is the semantic version (e.g., v0.1.0)
	Version = "dev"
	
	// Commit is the git commit hash
	Commit = "unknown"
	
	// Date is the build date
	Date = "unknown"
	
	// BuiltBy indicates who built the binary
	BuiltBy = "unknown"
)

// Info returns version information as a formatted string
func Info() string {
	return fmt.Sprintf(
		"pyqol %s\nCommit: %s\nBuilt: %s\nGo: %s\nOS/Arch: %s/%s",
		Version,
		Commit,
		Date,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)
}

// Short returns just the version string
func Short() string {
	return Version
}