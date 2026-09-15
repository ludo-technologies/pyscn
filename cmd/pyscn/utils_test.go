package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateOutputFilePath_CreatesDefaultDirectory(t *testing.T) {
	// macOS: t.TempDir() returns /var/folders/... but os.Getwd() after Chdir
	// returns /private/var/folders/... — normalise with EvalSymlinks
	tempDir := t.TempDir()
	tempDir, err := filepath.EvalSymlinks(tempDir)
	if err != nil {
		t.Fatalf("failed to eval symlinks on tempDir: %v", err)
	}

	// Change to the temp directory
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldCwd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	})

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

	gitignorePath := filepath.Join(expectedDir, ".gitignore")
	gitignoreContent, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("expected .gitignore to be created at %q: %v", gitignorePath, err)
	}
	if string(gitignoreContent) != "*\n" {
		t.Errorf("expected .gitignore content %q, got %q", "*\\n", string(gitignoreContent))
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
	// macOS: t.TempDir() returns /var/folders/... but os.Getwd() after Chdir
	// returns /private/var/folders/... — normalise with EvalSymlinks
	tempDir := t.TempDir()
	tempDir, err := filepath.EvalSymlinks(tempDir)
	if err != nil {
		t.Fatalf("failed to eval symlinks on tempDir: %v", err)
	}

	// Change to the temp directory
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldCwd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	})

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

func TestEnsureOutputGitignore_CreatesFileWhenMissing(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "reports")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("failed to create outputDir: %v", err)
	}

	if err := ensureOutputGitignore(outputDir); err != nil {
		t.Fatalf("ensureOutputGitignore returned error: %v", err)
	}

	gitignorePath := filepath.Join(outputDir, ".gitignore")
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("expected .gitignore at %q: %v", gitignorePath, err)
	}
	if string(content) != "*\n" {
		t.Errorf("expected .gitignore content %q, got %q", "*\\n", string(content))
	}
}

func TestEnsureOutputGitignore_DoesNotOverwriteExistingFile(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "reports")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("failed to create outputDir: %v", err)
	}

	gitignorePath := filepath.Join(outputDir, ".gitignore")
	originalContent := []byte("keep-this\n")
	if err := os.WriteFile(gitignorePath, originalContent, 0o644); err != nil {
		t.Fatalf("failed to seed .gitignore: %v", err)
	}

	if err := ensureOutputGitignore(outputDir); err != nil {
		t.Fatalf("ensureOutputGitignore returned error: %v", err)
	}

	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}
	if !bytes.Equal(content, originalContent) {
		t.Errorf("expected existing .gitignore to be unchanged, got %q", string(content))
	}
}

func TestGenerateOutputFilePath_ContinuesWhenGitignoreCreationFails(t *testing.T) {
	// macOS: t.TempDir() returns /var/folders/... but os.Getwd() after Chdir
	// returns /private/var/folders/... — normalise with EvalSymlinks
	tempDir := t.TempDir()
	tempDir, err := filepath.EvalSymlinks(tempDir)
	if err != nil {
		t.Fatalf("failed to eval symlinks on tempDir: %v", err)
	}

	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldCwd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	})

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	outputDir := filepath.Join(tempDir, ".pyscn", "reports")
	if err := os.MkdirAll(filepath.Join(outputDir, ".gitignore"), 0o755); err != nil {
		t.Fatalf("failed to create directory at .gitignore path: %v", err)
	}

	path, err := generateOutputFilePath("analyze", "html", ".")
	if err != nil {
		t.Fatalf("generateOutputFilePath returned error: %v", err)
	}

	if filepath.Dir(path) != outputDir {
		t.Errorf("expected directory %q, got %q", outputDir, filepath.Dir(path))
	}
}

func TestGenerateOutputFilePath_DifferentExtensions(t *testing.T) {
	// macOS: t.TempDir() returns /var/folders/... but os.Getwd() after Chdir
	// returns /private/var/folders/... — normalise with EvalSymlinks
	tempDir := t.TempDir()
	tempDir, err := filepath.EvalSymlinks(tempDir)
	if err != nil {
		t.Fatalf("failed to eval symlinks on tempDir: %v", err)
	}

	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldCwd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	})

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
