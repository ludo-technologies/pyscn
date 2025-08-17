package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pyqol/pyqol/internal/analyzer"
	"github.com/pyqol/pyqol/internal/config"
)

const testPythonCode = `
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

func TestFileComplexityAnalyzer(t *testing.T) {
	// Create temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.py")
	
	err := os.WriteFile(testFile, []byte(testPythonCode), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("Text output format", func(t *testing.T) {
		var output bytes.Buffer
		
		cfg := config.DefaultConfig()
		cfg.Output.Format = "text"
		
		analyzer, err := analyzer.NewFileComplexityAnalyzer(cfg, &output)
		if err != nil {
			t.Fatalf("Failed to create analyzer: %v", err)
		}
		
		err = analyzer.AnalyzeFile(testFile)
		if err != nil {
			t.Fatalf("Failed to analyze file: %v", err)
		}
		
		result := output.String()
		if !strings.Contains(result, "simple_function") {
			t.Error("Output should contain simple_function")
		}
		if !strings.Contains(result, "complex_function") {
			t.Error("Output should contain complex_function")
		}
		if !strings.Contains(result, "Total Functions: 3") {
			t.Error("Output should show 3 total functions")
		}
	})

	t.Run("JSON output format", func(t *testing.T) {
		var output bytes.Buffer
		
		cfg := config.DefaultConfig()
		cfg.Output.Format = "json"
		
		analyzer, err := analyzer.NewFileComplexityAnalyzer(cfg, &output)
		if err != nil {
			t.Fatalf("Failed to create analyzer: %v", err)
		}
		
		err = analyzer.AnalyzeFile(testFile)
		if err != nil {
			t.Fatalf("Failed to analyze file: %v", err)
		}
		
		result := output.String()
		
		var jsonResult map[string]interface{}
		if err := json.Unmarshal([]byte(result), &jsonResult); err != nil {
			t.Fatalf("Output should be valid JSON: %v\nOutput: %s", err, result)
		}
		
		if summary, ok := jsonResult["summary"]; ok {
			if summaryMap, ok := summary.(map[string]interface{}); ok {
				if totalFunctions, ok := summaryMap["total_functions"]; ok {
					if totalFunctions != float64(3) {
						t.Errorf("Expected 3 total functions, got %v", totalFunctions)
					}
				}
			}
		}
	})

	t.Run("CSV output format", func(t *testing.T) {
		var output bytes.Buffer
		
		cfg := config.DefaultConfig()
		cfg.Output.Format = "csv"
		
		analyzer, err := analyzer.NewFileComplexityAnalyzer(cfg, &output)
		if err != nil {
			t.Fatalf("Failed to create analyzer: %v", err)
		}
		
		err = analyzer.AnalyzeFile(testFile)
		if err != nil {
			t.Fatalf("Failed to analyze file: %v", err)
		}
		
		result := output.String()
		if !strings.Contains(result, "Function,Complexity,Risk") {
			t.Error("CSV output should contain header")
		}
		if !strings.Contains(result, "simple_function,1,low") {
			t.Error("CSV output should contain simple_function data")
		}
		if !strings.Contains(result, "complex_function,3,low") {
			t.Error("CSV output should contain complex_function data")
		}
	})

	t.Run("Minimum complexity filtering", func(t *testing.T) {
		var output bytes.Buffer
		
		cfg := config.DefaultConfig()
		cfg.Output.Format = "text"
		cfg.Output.MinComplexity = 2
		
		analyzer, err := analyzer.NewFileComplexityAnalyzer(cfg, &output)
		if err != nil {
			t.Fatalf("Failed to create analyzer: %v", err)
		}
		
		err = analyzer.AnalyzeFile(testFile)
		if err != nil {
			t.Fatalf("Failed to analyze file: %v", err)
		}
		
		result := output.String()
		if strings.Contains(result, "simple_function") {
			t.Error("simple_function should be filtered out with min=2")
		}
		if !strings.Contains(result, "complex_function") {
			t.Error("complex_function should be included with min=2")
		}
	})
}