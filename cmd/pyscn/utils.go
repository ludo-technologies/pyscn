package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ludo-technologies/pyscn/internal/config"
)

// generateTimestampedFileName generates a filename with timestamp suffix
// Single responsibility: filename generation only
func generateTimestampedFileName(command, extension string) string {
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s_%s.%s", command, timestamp, extension)
}

// resolveOutputDirectory determines the output directory from configuration
// Single responsibility: directory resolution only
// Returns directory path and any error encountered during config loading
func resolveOutputDirectory(targetPath string) (string, error) {
	cfg, err := config.LoadConfigWithTarget("", targetPath)
	if err != nil {
		// Don't hide configuration errors - they should be visible to users
		return "", fmt.Errorf("failed to load configuration: %w", err)
	}

	if cfg != nil && cfg.Output.Directory != "" {
		return cfg.Output.Directory, nil
	}

	// Default output directory when not specified in config
	// Use a tool-specific hidden directory under the current working directory
	// (avoids writing into analyzed source directories by default)
	cwd, err := os.Getwd()
	if err != nil {
		// Fallback to relative path if CWD not available
		return filepath.Join(".pyscn", "reports"), nil
	}
	return filepath.Join(cwd, ".pyscn", "reports"), nil
}

// generateOutputFilePath combines filename generation and directory resolution
// Orchestrates the workflow but delegates specific concerns
// Returns the full file path and any error encountered
func generateOutputFilePath(command, extension, targetPath string) (string, error) {
	filename := generateTimestampedFileName(command, extension)
	outputDir, err := resolveOutputDirectory(targetPath)
	if err != nil {
		return "", err
	}

	// Ensure the directory exists before returning the path. At this point,
	// outputDir is always non-empty because resolveOutputDirectory provides
	// a default (e.g., .pyscn/reports under CWD) when config is unset.
	if mkErr := os.MkdirAll(outputDir, 0o755); mkErr != nil {
		return "", fmt.Errorf("failed to create output directory %s: %w", outputDir, mkErr)
	}
	return filepath.Join(outputDir, filename), nil
}

// getTargetPathFromArgs extracts the first argument as target path, or returns empty string
func getTargetPathFromArgs(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return ""
}

// isInteractiveEnvironment returns true if the environment appears to be
// an interactive TTY session (and not CI), used to decide auto-open behavior.
func isInteractiveEnvironment() bool {
	if os.Getenv("CI") != "" {
		return false
	}
	// Best-effort TTY detection without external deps
	if fi, err := os.Stderr.Stat(); err == nil {
		return (fi.Mode() & os.ModeCharDevice) != 0
	}
	return false
}
