package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateOutputFilePath_CreatesDefaultDirectory(t *testing.T) {
	// Create a temporary directory to work in
	tempDir := t.TempDir()

	// Change to the temp directory
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(oldCwd) //nolint:errcheck

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Generate output file path
	path, err := generateOutputFilePath("analyze", "html", ".")
	if err != nil {
		t.Fatalf("generateOutputFilePath returned error: %v", err)
	}

	// Verify the path contains the expected directory structure
	expectedDir := filepath.Join(tempDir, ".pyscn", "reports")
	if filepath.Dir(path) != expectedDir {
		t.Errorf("expected directory %q, got %q", expectedDir, filepath.Dir(path))
	}

	// Verify the directory was actually created
	if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
		t.Errorf("expected directory %q to be created, but it does not exist", expectedDir)
	}

	// Verify the filename has the expected format
	filename := filepath.Base(path)
	expectedPrefix := "analyze_"
	if len(filename) < len(expectedPrefix)+10 || filename[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("expected filename to start with %q and have timestamp, got %q", expectedPrefix, filename)
	}
	if filepath.Ext(path) != ".html" {
		t.Errorf("expected extension .html, got %q", filepath.Ext(path))
	}
}

func TestResolveOutputDirectory_DefaultToCWD(t *testing.T) {
	// Create a temporary directory to work in
	tempDir := t.TempDir()

	// Change to the temp directory
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(oldCwd) //nolint:errcheck

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Resolve output directory with no config file
	outputDir, err := resolveOutputDirectory(".")
	if err != nil {
		t.Fatalf("resolveOutputDirectory returned error: %v", err)
	}

	// Verify it defaults to .pyscn/reports under CWD
	expectedDir := filepath.Join(tempDir, ".pyscn", "reports")
	if outputDir != expectedDir {
		t.Errorf("expected directory %q, got %q", expectedDir, outputDir)
	}
}

func TestGenerateOutputFilePath_DifferentExtensions(t *testing.T) {
	tempDir := t.TempDir()

	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(oldCwd) //nolint:errcheck

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	testCases := []struct {
		command   string
		extension string
	}{
		{"complexity", "json"},
		{"deadcode", "yaml"},
		{"clone", "csv"},
		{"analyze", "html"},
	}

	for _, tc := range testCases {
		path, err := generateOutputFilePath(tc.command, tc.extension, ".")
		if err != nil {
			t.Errorf("[%s.%s] generateOutputFilePath returned error: %v", tc.command, tc.extension, err)
			continue
		}

		if filepath.Ext(path) != "."+tc.extension {
			t.Errorf("[%s.%s] expected extension .%s, got %q", tc.command, tc.extension, tc.extension, filepath.Ext(path))
		}
	}
}
