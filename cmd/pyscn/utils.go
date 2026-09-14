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

// ensureOutputGitignore creates a .gitignore file with "*" in the given directory.
// If the file already exists, it is left unchanged.
func ensureOutputGitignore(dir string) error {
	gitignorePath := filepath.Join(dir, ".gitignore")
	gitignore, err := os.OpenFile(gitignorePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return nil
		}
		return fmt.Errorf("failed to create .gitignore at %s: %w", gitignorePath, err)
	}

	if _, err := gitignore.WriteString("*\n"); err != nil {
		_ = gitignore.Close()
		return fmt.Errorf("failed to write .gitignore at %s: %w", gitignorePath, err)
	}

	if err := gitignore.Close(); err != nil {
		return fmt.Errorf("failed to close .gitignore at %s: %w", gitignorePath, err)
	}

	return nil
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

	if err := ensureOutputGitignore(outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to create output .gitignore in %s: %v\n", outputDir, err)
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
