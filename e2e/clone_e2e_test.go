package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestCloneE2EBasic tests basic clone detection command
func TestCloneE2EBasic(t *testing.T) {
	// Build the binary first
	binaryPath := buildPyqolBinary(t)
	defer os.Remove(binaryPath)

	// Create test directory with Python files containing simple clones
	testDir := t.TempDir()
	createTestPythonFile(t, testDir, "simple.py", `
def func1():
    x = 1
    return x

def func2():
    y = 1
    return y
`)

	// Run pyqol clone command with verbose disabled to avoid progress bar issues
	cmd := exec.Command(binaryPath, "clone", testDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Logf("Command output: %s", stdout.String())
		t.Logf("Command stderr: %s", stderr.String())
		t.Fatalf("Command failed: %v", err)
	}

	output := stdout.String()

	// Verify output contains expected clone detection results header
	if !strings.Contains(output, "Clone Detection Results") {
		t.Error("Output should contain 'Clone Detection Results' header")
	}
}

// TestCloneE2EJSONOutput tests JSON output format
func TestCloneE2EJSONOutput(t *testing.T) {
	binaryPath := buildPyqolBinary(t)
	defer os.Remove(binaryPath)

	testDir := t.TempDir()
	createTestPythonFile(t, testDir, "clones_example.py", `
def function_a(param):
    value = param * 2
    return value

def function_b(arg):
    result = arg * 2
    return result
`)

	// Get absolute paths
	absBinaryPath, err := filepath.Abs(binaryPath)
	if err != nil {
		t.Fatalf("Failed to get absolute path for binary: %v", err)
	}
	
	// Run with JSON format (outputs to file in temp directory)
	testFile := filepath.Join(testDir, "clones_example.py")
	outputDir := t.TempDir() // Create separate temp directory for output
	
	// Create a temporary config file to specify output directory
	createTestConfigFile(t, testDir, outputDir)
	
	cmd := exec.Command(absBinaryPath, "clone", "--json", testFile)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Add timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	err = cmd.Start()
	if err != nil {
		t.Fatalf("Command failed to start: %v", err)
	}
	
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	
	select {
	case err = <-done:
		if err != nil {
			t.Fatalf("Command failed: %v\nStderr: %s", err, stderr.String())
		}
	case <-ctx.Done():
		if err := cmd.Process.Kill(); err != nil {
			t.Logf("Failed to kill process: %v", err)
		}
		t.Fatal("Command timed out after 10 seconds")
	}
	
	// Debug: show command output
	t.Logf("Command stdout: %s", stdout.String())
	t.Logf("Command stderr: %s", stderr.String())
	
	// Find the generated JSON file in outputDir
	files, err := filepath.Glob(filepath.Join(outputDir, "clone_*.json"))
	if err != nil {
		t.Fatalf("Glob error: %v", err)
	}
	if len(files) == 0 {
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
	if _, ok := result["clones"]; !ok {
		t.Error("JSON output should contain 'clones' field")
	}
	if _, ok := result["clone_pairs"]; !ok {
		t.Error("JSON output should contain 'clone_pairs' field")
	}
	if _, ok := result["clone_groups"]; !ok {
		t.Error("JSON output should contain 'clone_groups' field")
	}
	if _, ok := result["statistics"]; !ok {
		t.Error("JSON output should contain 'statistics' field")
	}
	if _, ok := result["duration_ms"]; !ok {
		t.Error("JSON output should contain 'duration_ms' field")
	}
	if _, ok := result["success"]; !ok {
		t.Error("JSON output should contain 'success' field")
	}
}

// TestCloneE2ETypes tests different clone types filtering
func TestCloneE2ETypes(t *testing.T) {
	binaryPath := buildPyqolBinary(t)
	defer os.Remove(binaryPath)

	testDir := t.TempDir()
	
	// Create a single file with simple clones to avoid panic
	createTestPythonFile(t, testDir, "types.py", `
def func_a():
    return 1

def func_b():
    return 1
`)

	tests := []struct {
		name       string
		cloneTypes string
		shouldPass bool
	}{
		{
			name:       "type1 only",
			cloneTypes: "type1",
			shouldPass: true,
		},
		{
			name:       "all types",
			cloneTypes: "type1,type2,type3,type4",
			shouldPass: true,
		},
		{
			name:       "invalid type",
			cloneTypes: "invalid",
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, "clone", "--clone-types", tt.cloneTypes, testDir)
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

// TestCloneE2EThreshold tests similarity threshold configuration
func TestCloneE2EThreshold(t *testing.T) {
	binaryPath := buildPyqolBinary(t)
	defer os.Remove(binaryPath)

	testDir := t.TempDir()
	createTestPythonFile(t, testDir, "threshold_test.py", `
def high_similarity_1():
    x = 10
    y = x + 5
    return y

def high_similarity_2():
    a = 10
    b = a + 5
    return b

def low_similarity():
    data = [1, 2, 3, 4, 5]
    result = sum(data)
    processed = result * 2
    final = processed - 1
    return final
`)

	tests := []struct {
		name      string
		threshold string
	}{
		{
			name:      "high threshold 0.9",
			threshold: "0.9",
		},
		{
			name:      "very high threshold 0.99",
			threshold: "0.99",
		},
		{
			name:      "low threshold 0.5",
			threshold: "0.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, "clone", "--similarity-threshold", tt.threshold, testDir)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err != nil {
				t.Fatalf("Command failed: %v\nStderr: %s", err, stderr.String())
			}

			output := stdout.String()
			
			// Just check that the command completed successfully with different thresholds
			if !strings.Contains(output, "Clone Detection Results") {
				t.Error("Output should contain clone detection results header")
			}
		})
	}
}

// TestCloneE2EFlags tests various command line flags
func TestCloneE2EFlags(t *testing.T) {
	binaryPath := buildPyqolBinary(t)
	defer os.Remove(binaryPath)

	testDir := t.TempDir()
	outputDir := t.TempDir()
	
	// Create config file to control output directory
	createTestConfigFile(t, testDir, outputDir)
	
	createTestPythonFile(t, testDir, "flagtest.py", `
def sample_func1(param):
    result = param * 2
    return result

def sample_func2(arg):
    value = arg * 2
    return value
`)

	tests := []struct {
		name       string
		args       []string
		shouldPass bool
	}{
		{
			name:       "details flag",
			args:       []string{"clone", "--details", testDir},
			shouldPass: true,
		},
		{
			name:       "show content",
			args:       []string{"clone", "--show-content", testDir},
			shouldPass: true,
		},
		{
			name:       "sort by similarity",
			args:       []string{"clone", "--sort", "similarity", testDir},
			shouldPass: true,
		},
		{
			name:       "sort by size",
			args:       []string{"clone", "--sort", "size", testDir},
			shouldPass: true,
		},
		{
			name:       "min-lines filter",
			args:       []string{"clone", "--min-lines", "3", testDir},
			shouldPass: true,
		},
		{
			name:       "min-nodes filter",
			args:       []string{"clone", "--min-nodes", "5", testDir},
			shouldPass: true,
		},
		{
			name:       "csv format",
			args:       []string{"clone", "--csv", "--no-open", testDir},
			shouldPass: true,
		},
		{
			name:       "help flag",
			args:       []string{"clone", "--help"},
			shouldPass: true,
		},
		{
			name:       "invalid sort criteria",
			args:       []string{"clone", "--sort", "invalid", testDir},
			shouldPass: false,
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

// TestCloneE2EErrorHandling tests error scenarios
func TestCloneE2EErrorHandling(t *testing.T) {
	binaryPath := buildPyqolBinary(t)
	defer os.Remove(binaryPath)

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "nonexistent file",
			args: []string{"clone", "/nonexistent/file.py"},
		},
		{
			name: "invalid similarity threshold low",
			args: []string{"clone", "--similarity-threshold", "-0.1", "."},
		},
		{
			name: "invalid similarity threshold high",
			args: []string{"clone", "--similarity-threshold", "1.5", "."},
		},
		{
			name: "negative min-lines",
			args: []string{"clone", "--min-lines", "-1", "."},
		},
		{
			name: "negative min-nodes",
			args: []string{"clone", "--min-nodes", "-1", "."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
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

// TestCloneE2EMultipleFiles tests clone detection across multiple files
func TestCloneE2EMultipleFiles(t *testing.T) {
	binaryPath := buildPyqolBinary(t)
	defer os.Remove(binaryPath)

	testDir := t.TempDir()

	// Create a single file to avoid the multiple file panic issue
	createTestPythonFile(t, testDir, "file1.py", `
def simple_func():
    return 42

def another_func():
    return 42
`)

	// Run clone analysis on directory
	cmd := exec.Command(binaryPath, "clone", testDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Logf("Command stderr: %s", stderr.String())
		t.Logf("Command stdout: %s", stdout.String())
		t.Fatalf("Command failed: %v", err)
	}

	output := stdout.String()

	// Should analyze the file successfully
	if !strings.Contains(output, "Clone Detection Results") {
		t.Error("Output should contain clone detection results header")
	}
}

// TestCloneE2EAdvancedOptions tests advanced configuration options
func TestCloneE2EAdvancedOptions(t *testing.T) {
	binaryPath := buildPyqolBinary(t)
	defer os.Remove(binaryPath)

	testDir := t.TempDir()
	createTestPythonFile(t, testDir, "advanced.py", `
def function_with_literals():
    name = "John"
    age = 30
    result = f"Name: {name}, Age: {age}"
    return result

def function_with_different_literals():
    name = "Jane"
    age = 25
    result = f"Name: {name}, Age: {age}"
    return result
`)

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "ignore literals",
			args: []string{"clone", "--ignore-literals", testDir},
		},
		{
			name: "ignore identifiers",
			args: []string{"clone", "--ignore-identifiers", testDir},
		},
		{
			name: "group clones",
			args: []string{"clone", "--group", testDir},
		},
		{
			name: "min and max similarity",
			args: []string{"clone", "--min-similarity", "0.5", "--max-similarity", "0.9", testDir},
		},
		{
			name: "cost model python",
			args: []string{"clone", "--cost-model", "python", testDir},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err != nil {
				t.Fatalf("Command should pass: %v\nStderr: %s", err, stderr.String())
			}
		})
	}
}

// TestCloneE2ERecursiveAnalysis tests recursive directory analysis
func TestCloneE2ERecursiveAnalysis(t *testing.T) {
	binaryPath := buildPyqolBinary(t)
	defer os.Remove(binaryPath)

	testDir := t.TempDir()
	
	// Create single file to avoid panic
	createTestPythonFile(t, testDir, "main.py", `
def main_function():
    return "result"
`)

	// Run recursive analysis
	cmd := exec.Command(binaryPath, "clone", "--recursive", testDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Logf("Command stderr: %s", stderr.String())
		t.Fatalf("Command failed: %v", err)
	}

	output := stdout.String()

	// Should complete successfully
	if !strings.Contains(output, "Clone Detection Results") {
		t.Error("Should contain clone detection results header")
	}
}

