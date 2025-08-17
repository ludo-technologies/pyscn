package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Test data: simple Python files
const simplePythonCode = `
def simple_function(x):
    return x * 2

def complex_function(x):
    if x > 0:
        if x > 10:
            return "high"
        else:
            return "low" 
    else:
        return "negative"
`

const veryComplexPythonCode = `
def very_complex_function(x, y, z):
    if x > 0:
        if y > 0:
            if z > 0:
                for i in range(x):
                    if i % 2 == 0:
                        try:
                            result = x / i
                        except ZeroDivisionError:
                            continue
                        if result > 10:
                            return "high"
                        elif result > 5:
                            return "medium"
                        else:
                            return "low"
                    else:
                        while i < y:
                            i += 1
                            if i > z:
                                break
            else:
                return "z_negative"
        else:
            return "y_negative"
    else:
        return "x_negative"
`

func TestComplexityCommand(t *testing.T) {
	// Create temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.py")
	
	err := os.WriteFile(testFile, []byte(simplePythonCode), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name           string
		args           []string
		expectedError  bool
		checkOutput    func(t *testing.T, output string)
	}{
		{
			name: "Basic complexity analysis",
			args: []string{"complexity", "--format", "text", testFile},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "simple_function") {
					t.Error("Output should contain simple_function")
				}
				if !strings.Contains(output, "complex_function") {
					t.Error("Output should contain complex_function")
				}
				if !strings.Contains(output, "Total Functions: 3") {
					t.Error("Output should show 3 total functions")
				}
			},
		},
		{
			name: "JSON output format",
			args: []string{"complexity", "--format", "json", testFile},
			checkOutput: func(t *testing.T, output string) {
				// Trim any whitespace that might cause JSON parsing issues
				trimmedOutput := strings.TrimSpace(output)
				
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(trimmedOutput), &result); err != nil {
					t.Errorf("Output should be valid JSON: %v\nOutput was: %q", err, output)
					return
				}
				
				if summary, ok := result["summary"]; ok {
					if summaryMap, ok := summary.(map[string]interface{}); ok {
						if totalFunctions, ok := summaryMap["total_functions"]; ok {
							if totalFunctions != float64(3) {
								t.Errorf("Expected 3 total functions, got %v", totalFunctions)
							}
						}
					}
				}
			},
		},
		{
			name: "CSV output format",
			args: []string{"complexity", "--format", "csv", testFile},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Function,Complexity,Risk") {
					t.Error("CSV output should contain header")
				}
				if !strings.Contains(output, "simple_function,1,low") {
					t.Error("CSV output should contain simple_function data")
				}
				if !strings.Contains(output, "complex_function,3,low") {
					t.Error("CSV output should contain complex_function data")
				}
			},
		},
		{
			name: "Minimum complexity filtering",
			args: []string{"complexity", "--format", "text", "--min", "2", testFile},
			checkOutput: func(t *testing.T, output string) {
				if strings.Contains(output, "simple_function") {
					t.Error("simple_function should be filtered out with min=2")
				}
				if !strings.Contains(output, "complex_function") {
					t.Error("complex_function should be included with min=2")
				}
			},
		},
		{
			name: "Sort by name",
			args: []string{"complexity", "--format", "text", "--sort", "name", testFile},
			checkOutput: func(t *testing.T, output string) {
				// Check that functions appear in name order
				complexPos := strings.Index(output, "complex_function")
				mainPos := strings.Index(output, "main")
				simplePos := strings.Index(output, "simple_function")
				
				if complexPos == -1 || mainPos == -1 || simplePos == -1 {
					t.Error("All functions should be present in output")
				}
				
				// complex_function should come before main and simple_function alphabetically
				if complexPos > mainPos {
					t.Error("complex_function should come before main when sorted by name")
				}
			},
		},
		{
			name:          "No files provided",
			args:          []string{"complexity"},
			expectedError: true,
		},
		{
			name:          "Non-existent file",
			args:          []string{"complexity", "/nonexistent/file.py"},
			expectedError: true,
		},
		{
			name:          "Non-Python file",
			args:          []string{"complexity", "/tmp/test.txt"},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags to default values for each test
			outputFormat = "text"
			minComplexity = 1
			maxComplexity = 0
			sortBy = "complexity"
			showDetails = false
			configFile = ""
			lowThreshold = 9
			mediumThreshold = 19
			
			// Capture output
			var output bytes.Buffer
			
			// Create a new root command for testing
			testCmd := &cobra.Command{
				Use: "pyqol",
			}
			
			// Create a new complexity command for this test
			newComplexityCmd := &cobra.Command{
				Use:   "complexity [files...]",
				Short: "Analyze cyclomatic complexity of Python files",
				Args:  cobra.MinimumNArgs(1),
				RunE:  runComplexityAnalysis,
			}
			
			// Re-add all flags to the new command
			newComplexityCmd.Flags().StringVarP(&outputFormat, "format", "f", "text", "Output format (text, json, yaml, csv)")
			newComplexityCmd.Flags().IntVar(&minComplexity, "min", 1, "Minimum complexity to report")
			newComplexityCmd.Flags().IntVar(&maxComplexity, "max", 0, "Maximum complexity limit (0 = no limit)")
			newComplexityCmd.Flags().StringVar(&sortBy, "sort", "complexity", "Sort results by (name, complexity, risk)")
			newComplexityCmd.Flags().BoolVar(&showDetails, "details", false, "Show detailed complexity breakdown")
			newComplexityCmd.Flags().StringVarP(&configFile, "config", "c", "", "Configuration file path")
			newComplexityCmd.Flags().IntVar(&lowThreshold, "low-threshold", 9, "Low complexity threshold")
			newComplexityCmd.Flags().IntVar(&mediumThreshold, "medium-threshold", 19, "Medium complexity threshold")
			
			testCmd.AddCommand(newComplexityCmd)
			testCmd.SetOut(&output)
			testCmd.SetErr(&output)
			testCmd.SetArgs(tt.args)
			
			// Execute command
			err := testCmd.Execute()
			
			// Check error expectation
			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			// Check output if no error expected
			if !tt.expectedError && tt.checkOutput != nil {
				tt.checkOutput(t, output.String())
			}
		})
	}
}

func TestComplexityCommandWithComplexFile(t *testing.T) {
	// Create temporary test file with very complex function
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "complex.py")
	
	err := os.WriteFile(testFile, []byte(veryComplexPythonCode), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("Complex function analysis", func(t *testing.T) {
		// Reset flags
		outputFormat = "text"
		minComplexity = 1
		
		var output bytes.Buffer
		
		testCmd := &cobra.Command{Use: "pyqol"}
		newComplexityCmd := &cobra.Command{
			Use:   "complexity [files...]",
			Short: "Analyze cyclomatic complexity of Python files",
			Args:  cobra.MinimumNArgs(1),
			RunE:  runComplexityAnalysis,
		}
		newComplexityCmd.Flags().StringVarP(&outputFormat, "format", "f", "text", "Output format (text, json, yaml, csv)")
		newComplexityCmd.Flags().IntVar(&minComplexity, "min", 1, "Minimum complexity to report")
		
		testCmd.AddCommand(newComplexityCmd)
		testCmd.SetOut(&output)
		testCmd.SetErr(&output)
		testCmd.SetArgs([]string{"complexity", "--format", "json", testFile})
		
		err := testCmd.Execute()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		trimmedOutput := strings.TrimSpace(output.String())
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(trimmedOutput), &result); err != nil {
			t.Fatalf("Output should be valid JSON: %v\nOutput was: %q", err, output.String())
		}
		
		// Check that the very complex function has high complexity
		if results, ok := result["results"].([]interface{}); ok {
			found := false
			for _, r := range results {
				if resultMap, ok := r.(map[string]interface{}); ok {
					if name, ok := resultMap["function_name"].(string); ok && name == "very_complex_function" {
						if complexity, ok := resultMap["complexity"].(float64); ok {
							if complexity < 10 {
								t.Errorf("very_complex_function should have complexity >= 10, got %v", complexity)
							}
							found = true
						}
					}
				}
			}
			if !found {
				t.Error("very_complex_function not found in results")
			}
		}
	})
}

func TestComplexityCommandFlags(t *testing.T) {
	// Test flag parsing and validation
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.py")
	
	err := os.WriteFile(testFile, []byte(simplePythonCode), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	flagTests := []struct {
		name string
		args []string
		checkOutput func(t *testing.T, output string)
	}{
		{
			name: "Custom thresholds",
			args: []string{"complexity", "--format", "text", "--low-threshold", "2", "--medium-threshold", "5", testFile},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Total Functions") {
					t.Error("Output should contain analysis results")
				}
			},
		},
		{
			name: "Details flag",
			args: []string{"complexity", "--format", "text", "--details", testFile},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Total Functions") {
					t.Error("Output should contain analysis results")
				}
			},
		},
		{
			name: "Max complexity limit",
			args: []string{"complexity", "--format", "text", "--max", "20", testFile},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Total Functions") {
					t.Error("Output should contain analysis results")
				}
			},
		},
	}

	for _, tt := range flagTests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			
			testCmd := &cobra.Command{Use: "pyqol"}
			testCmd.AddCommand(complexityCmd)
			testCmd.SetOut(&output)
			testCmd.SetErr(&output)
			testCmd.SetArgs(tt.args)
			
			err := testCmd.Execute()
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if tt.checkOutput != nil {
				tt.checkOutput(t, output.String())
			}
		})
	}
}