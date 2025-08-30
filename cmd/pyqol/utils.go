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
func resolveOutputDirectory(targetPath string) string {
	cfg, err := config.LoadConfigWithTarget("", targetPath)
	if err == nil && cfg != nil && cfg.Output.Directory != "" {
		return cfg.Output.Directory
	}
	return "" // Empty means current directory
}

// generateOutputFilePath combines filename generation and directory resolution
// Orchestrates the workflow but delegates specific concerns
func generateOutputFilePath(command, extension, targetPath string) string {
	filename := generateTimestampedFileName(command, extension)
	outputDir := resolveOutputDirectory(targetPath)
	
	if outputDir != "" {
		return filepath.Join(outputDir, filename)
	}
	return filename
}

// getTargetPathFromArgs extracts the first argument as target path, or returns empty string
func getTargetPathFromArgs(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return ""
}

