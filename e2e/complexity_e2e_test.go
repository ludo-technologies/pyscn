package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestComplexityE2EBasic tests basic complexity analysis command
func TestComplexityE2EBasic(t *testing.T) {
	// Build the binary first
	binaryPath := buildPyqolBinary(t)
	defer os.Remove(binaryPath)

	// Create test directory with Python files
	testDir := t.TempDir()
	createTestPythonFile(t, testDir, "simple.py", `
def simple_function():
    return 42

def complex_function(x):
    if x > 0:
        if x > 10:
            return x * 2
        else:
            return x + 1
    else:
        return 0
`)

	// Run pyqol complexity command
	cmd := exec.Command(binaryPath, "complexity", testDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Command failed: %v\nStderr: %s", err, stderr.String())
	}

	output := stdout.String()

	// Verify output contains expected function names and complexity info
	if !strings.Contains(output, "simple_function") {
		t.Error("Output should contain 'simple_function'")
	}
	if !strings.Contains(output, "complex_function") {
		t.Error("Output should contain 'complex_function'")
	}
	if !strings.Contains(output, "Complexity") {
		t.Error("Output should contain complexity information")
	}
}

// TestComplexityE2EJSONOutput tests JSON output format
func TestComplexityE2EJSONOutput(t *testing.T) {
	binaryPath := buildPyqolBinary(t)
	defer os.Remove(binaryPath)

	testDir := t.TempDir()
	createTestPythonFile(t, testDir, "sample.py", `
def sample_function(x):
    if x > 0:
        return x * 2
    return 0
`)

	// Run with JSON format (outputs to file in temp directory)
	outputDir := t.TempDir() // Create separate temp directory for output
	
	// Create a temporary config file to specify output directory
	configFile := filepath.Join(testDir, ".pyqol.yaml")
	configContent := fmt.Sprintf("output:\n  directory: \"%s\"\n", outputDir)
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	
	cmd := exec.Command(binaryPath, "complexity", "--json", testDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Command failed: %v\nStderr: %s", err, stderr.String())
	}

	// Find the generated JSON file in outputDir
	files, err := filepath.Glob(filepath.Join(outputDir, "complexity_*.json"))
	if err != nil || len(files) == 0 {
		// List all files in outputDir for debugging
		allFiles, _ := os.ReadDir(outputDir)
		var fileNames []string
		for _, f := range allFiles {
			fileNames = append(fileNames, f.Name())
		}
		t.Fatalf("No JSON file generated in %s, files present: %v", outputDir, fileNames)
	}
	
	// Read and verify JSON file content
	jsonContent, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("Failed to read JSON file: %v", err)
	}
	
	// No need to clean up - t.TempDir() handles it automatically

	// Verify JSON output is valid
	var result map[string]interface{}
	if err := json.Unmarshal(jsonContent, &result); err != nil {
		t.Fatalf("Invalid JSON output: %v\nContent: %s", err, string(jsonContent))
	}

	// Check that JSON contains expected structure
	if _, ok := result["results"]; !ok {
		t.Error("JSON output should contain 'results' field")
	}
	if _, ok := result["summary"]; !ok {
		t.Error("JSON output should contain 'summary' field")
	}
	if _, ok := result["metadata"]; !ok {
		t.Error("JSON output should contain 'metadata' field")
	}
}

// TestComplexityE2EFlags tests various command line flags
func TestComplexityE2EFlags(t *testing.T) {
	binaryPath := buildPyqolBinary(t)
	defer os.Remove(binaryPath)

	testDir := t.TempDir()
	createTestPythonFile(t, testDir, "complex.py", `
def low_complexity():
    return 1

def medium_complexity(x):
    if x > 0:
        if x > 5:
            if x > 10:
                return x * 3
            return x * 2
        return x + 1
    return 0
`)

	tests := []struct {
		name       string
		args       []string
		shouldPass bool
	}{
		{
			name:       "min complexity filter",
			args:       []string{"complexity", "--min", "3", testDir},
			shouldPass: true,
		},
		{
			name:       "help flag",
			args:       []string{"complexity", "--help"},
			shouldPass: true,
		},
		{
			name:       "details flag",
			args:       []string{"complexity", "--details", testDir},
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
				cmd := exec.Command(binaryPath, tt.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			if tt.shouldPass && err != nil {
				t.Errorf("Command should pass but failed: %v\nStderr: %s", err, stderr.String())
			} else if !tt.shouldPass && err == nil {
				t.Error("Command should fail but passed")
			}
		})
	}
}

// TestComplexityE2EErrorHandling tests error scenarios
func TestComplexityE2EErrorHandling(t *testing.T) {
	binaryPath := buildPyqolBinary(t)
	defer os.Remove(binaryPath)

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "no arguments",
			args: []string{"complexity"},
		},
		{
			name: "nonexistent file",
			args: []string{"complexity", "/nonexistent/file.py"},
		},
		{
			name: "directory with no Python files",
			args: []string{"complexity", "EMPTY_DIR_PLACEHOLDER"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replace placeholder with actual empty directory
			args := make([]string, len(tt.args))
			copy(args, tt.args)
			for i, arg := range args {
				if arg == "EMPTY_DIR_PLACEHOLDER" {
					args[i] = t.TempDir() // Create empty directory for this test
				}
			}
			
			cmd := exec.Command(binaryPath, args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err == nil {
				t.Error("Command should fail but passed")
			}

			// Should have meaningful error message
			output := stderr.String() + stdout.String()
			if len(output) == 0 {
				t.Error("Should provide error message")
			}
		})
	}
}

// TestComplexityE2EMultipleFiles tests analysis of multiple files
func TestComplexityE2EMultipleFiles(t *testing.T) {
	binaryPath := buildPyqolBinary(t)
	defer os.Remove(binaryPath)

	testDir := t.TempDir()

	// Create multiple Python files
	createTestPythonFile(t, testDir, "file1.py", `
def func1():
    return 1
`)

	createTestPythonFile(t, testDir, "file2.py", `
def func2(x):
    if x > 0:
        return x
    return 0
`)

	// Run complexity analysis on directory
	cmd := exec.Command(binaryPath, "complexity", testDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Command failed: %v\nStderr: %s", err, stderr.String())
	}

	output := stdout.String()

	// Should contain functions from both files
	if !strings.Contains(output, "func1") {
		t.Error("Output should contain 'func1' from file1.py")
	}
	if !strings.Contains(output, "func2") {
		t.Error("Output should contain 'func2' from file2.py")
	}
}

// Helper functions

func buildPyqolBinary(t *testing.T) string {
	t.Helper()

	// Create temporary binary
	binaryPath := filepath.Join(t.TempDir(), "pyqol")

	// Build the binary from the project root (one level up from e2e directory)
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/pyqol")

	// Set working directory to project root
	projectRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}
	cmd.Dir = projectRoot

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build pyqol binary: %v", err)
	}

	return binaryPath
}

func createTestPythonFile(t *testing.T, dir, filename, content string) {
	t.Helper()

	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file %s: %v", filename, err)
	}
}
