package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/pyqol/pyqol/internal/config"
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
	
	return "", nil // Empty means current directory
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
	
	if outputDir != "" {
		return filepath.Join(outputDir, filename), nil
	}
	return filename, nil
}

// getTargetPathFromArgs extracts the first argument as target path, or returns empty string
func getTargetPathFromArgs(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return ""
}

