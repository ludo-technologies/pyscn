package analyzer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pyqol/pyqol/internal/config"
)

func createTestCFGs() []*CFG {
	cfgs := make([]*CFG, 3)
	
	// Simple function (complexity 1)
	cfgs[0] = NewCFG("simple_function")
	block := cfgs[0].CreateBlock("main")
	cfgs[0].ConnectBlocks(cfgs[0].Entry, block, EdgeNormal)
	cfgs[0].ConnectBlocks(block, cfgs[0].Exit, EdgeNormal)
	
	// Medium function (complexity 2) 
	cfgs[1] = NewCFG("medium_function")
	condBlock := cfgs[1].CreateBlock("condition")
	thenBlock := cfgs[1].CreateBlock("then")
	cfgs[1].ConnectBlocks(cfgs[1].Entry, condBlock, EdgeNormal)
	cfgs[1].ConnectBlocks(condBlock, thenBlock, EdgeCondTrue)
	cfgs[1].ConnectBlocks(condBlock, cfgs[1].Exit, EdgeCondFalse)
	cfgs[1].ConnectBlocks(thenBlock, cfgs[1].Exit, EdgeNormal)
	
	// Complex function (complexity 6)
	cfgs[2] = NewCFG("complex_function")
	current := cfgs[2].Entry
	for i := 0; i < 5; i++ {
		condBlock := cfgs[2].CreateBlock("condition")
		thenBlock := cfgs[2].CreateBlock("then")
		elseBlock := cfgs[2].CreateBlock("else")
		
		cfgs[2].ConnectBlocks(current, condBlock, EdgeNormal)
		cfgs[2].ConnectBlocks(condBlock, thenBlock, EdgeCondTrue)
		cfgs[2].ConnectBlocks(condBlock, elseBlock, EdgeCondFalse)
		cfgs[2].ConnectBlocks(thenBlock, elseBlock, EdgeNormal)
		
		current = elseBlock
	}
	cfgs[2].ConnectBlocks(current, cfgs[2].Exit, EdgeNormal)
	
	return cfgs
}

func TestNewComplexityAnalyzer(t *testing.T) {
	cfg := config.DefaultConfig()
	var buffer bytes.Buffer
	
	analyzer := NewComplexityAnalyzer(cfg, &buffer)
	
	if analyzer == nil {
		t.Fatal("Expected analyzer instance, got nil")
	}
	if analyzer.config != cfg {
		t.Error("Analyzer config not set correctly")
	}
	if analyzer.reporter == nil {
		t.Error("Analyzer reporter not set correctly")
	}
}

func TestNewComplexityAnalyzerWithDefaults(t *testing.T) {
	var buffer bytes.Buffer
	
	analyzer := NewComplexityAnalyzerWithDefaults(&buffer)
	
	if analyzer == nil {
		t.Fatal("Expected analyzer instance, got nil")
	}
	if analyzer.config == nil {
		t.Error("Analyzer should have default config")
	}
	
	// Verify it uses default configuration values
	if analyzer.config.Complexity.LowThreshold != 9 {
		t.Errorf("Expected default low threshold 9, got %d", analyzer.config.Complexity.LowThreshold)
	}
}

func TestNewComplexityAnalyzerNilOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	
	analyzer := NewComplexityAnalyzer(cfg, nil)
	
	if analyzer == nil {
		t.Fatal("Expected analyzer instance, got nil")
	}
	// Should handle nil output gracefully (defaults to os.Stdout)
}

func TestAnalyzeFunction(t *testing.T) {
	var buffer bytes.Buffer
	analyzer := NewComplexityAnalyzerWithDefaults(&buffer)
	
	// Test simple function
	simpleCFG := NewCFG("test_function")
	block := simpleCFG.CreateBlock("main")
	simpleCFG.ConnectBlocks(simpleCFG.Entry, block, EdgeNormal)
	simpleCFG.ConnectBlocks(block, simpleCFG.Exit, EdgeNormal)
	
	result := analyzer.AnalyzeFunction(simpleCFG)
	
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.Complexity != 1 {
		t.Errorf("Expected complexity 1, got %d", result.Complexity)
	}
	if result.FunctionName != "test_function" {
		t.Errorf("Expected function name 'test_function', got %s", result.FunctionName)
	}
	if result.RiskLevel != "low" {
		t.Errorf("Expected low risk, got %s", result.RiskLevel)
	}
}

func TestAnalyzeFunctions(t *testing.T) {
	var buffer bytes.Buffer
	analyzer := NewComplexityAnalyzerWithDefaults(&buffer)
	
	cfgs := createTestCFGs()
	results := analyzer.AnalyzeFunctions(cfgs)
	
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
	
	// Verify results are correct
	expectedComplexities := []int{1, 2, 6}
	for i, result := range results {
		if result.Complexity != expectedComplexities[i] {
			t.Errorf("Expected complexity %d for function %d, got %d", 
				expectedComplexities[i], i, result.Complexity)
		}
	}
}

func TestAnalyzeAndReport(t *testing.T) {
	t.Run("TextOutput", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Output.Format = "text"
		
		var buffer bytes.Buffer
		analyzer := NewComplexityAnalyzer(cfg, &buffer)
		
		cfgs := createTestCFGs()
		err := analyzer.AnalyzeAndReport(cfgs)
		
		if err != nil {
			t.Fatalf("Failed to analyze and report: %v", err)
		}
		
		output := buffer.String()
		
		// Verify report contents
		if !strings.Contains(output, "Complexity Analysis Report") {
			t.Error("Missing report title")
		}
		if !strings.Contains(output, "Total Functions: 3") {
			t.Error("Missing total functions count")
		}
		
		// Verify all functions are mentioned
		expectedFunctions := []string{"simple_function", "medium_function", "complex_function"}
		for _, function := range expectedFunctions {
			if !strings.Contains(output, function) {
				t.Errorf("Missing function %s in output", function)
			}
		}
	})
	
	t.Run("JSONOutput", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Output.Format = "json"
		
		var buffer bytes.Buffer
		analyzer := NewComplexityAnalyzer(cfg, &buffer)
		
		cfgs := createTestCFGs()
		err := analyzer.AnalyzeAndReport(cfgs)
		
		if err != nil {
			t.Fatalf("Failed to analyze and report JSON: %v", err)
		}
		
		output := buffer.String()
		
		// Should be valid JSON containing expected data
		if !strings.Contains(output, `"total_functions": 3`) {
			t.Error("JSON output missing total functions")
		}
		if !strings.Contains(output, `"simple_function"`) {
			t.Error("JSON output missing function name")
		}
	})
}

func TestCheckComplexityLimits(t *testing.T) {
	t.Run("WithinLimits", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Complexity.MaxComplexity = 10 // All test functions should be within this limit
		
		var buffer bytes.Buffer
		analyzer := NewComplexityAnalyzer(cfg, &buffer)
		
		cfgs := createTestCFGs()
		withinLimits, violations := analyzer.CheckComplexityLimits(cfgs)
		
		if !withinLimits {
			t.Error("Expected all functions to be within limits")
		}
		if len(violations) != 0 {
			t.Errorf("Expected no violations, got %d", len(violations))
		}
	})
	
	t.Run("ExceedsLimits", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Complexity.MaxComplexity = 3 // complex_function (complexity 6) should exceed this
		
		var buffer bytes.Buffer
		analyzer := NewComplexityAnalyzer(cfg, &buffer)
		
		cfgs := createTestCFGs()
		withinLimits, violations := analyzer.CheckComplexityLimits(cfgs)
		
		if withinLimits {
			t.Error("Expected some functions to exceed limits")
		}
		if len(violations) != 1 {
			t.Errorf("Expected 1 violation, got %d", len(violations))
		}
		
		if len(violations) > 0 {
			if violations[0].FunctionName != "complex_function" {
				t.Errorf("Expected violation for complex_function, got %s", violations[0].FunctionName)
			}
			if violations[0].Complexity != 6 {
				t.Errorf("Expected violation complexity 6, got %d", violations[0].Complexity)
			}
		}
	})
	
	t.Run("NoLimits", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Complexity.MaxComplexity = 0 // No limits
		
		var buffer bytes.Buffer
		analyzer := NewComplexityAnalyzer(cfg, &buffer)
		
		cfgs := createTestCFGs()
		withinLimits, violations := analyzer.CheckComplexityLimits(cfgs)
		
		if !withinLimits {
			t.Error("Expected all functions to be within limits when no limit is set")
		}
		if len(violations) != 0 {
			t.Errorf("Expected no violations when no limit is set, got %d", len(violations))
		}
	})
}

func TestGetConfiguration(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Complexity.LowThreshold = 5 // Custom value
	
	var buffer bytes.Buffer
	analyzer := NewComplexityAnalyzer(cfg, &buffer)
	
	retrievedCfg := analyzer.GetConfiguration()
	
	if retrievedCfg != cfg {
		t.Error("Expected same config instance")
	}
	if retrievedCfg.Complexity.LowThreshold != 5 {
		t.Errorf("Expected custom low threshold 5, got %d", retrievedCfg.Complexity.LowThreshold)
	}
}

func TestUpdateConfiguration(t *testing.T) {
	oldCfg := config.DefaultConfig()
	var buffer bytes.Buffer
	analyzer := NewComplexityAnalyzer(oldCfg, &buffer)
	
	newCfg := config.DefaultConfig()
	newCfg.Complexity.LowThreshold = 15
	newCfg.Output.Format = "json"
	
	analyzer.UpdateConfiguration(newCfg)
	
	if analyzer.config != newCfg {
		t.Error("Configuration not updated correctly")
	}
	if analyzer.config.Complexity.LowThreshold != 15 {
		t.Errorf("Expected updated low threshold 15, got %d", analyzer.config.Complexity.LowThreshold)
	}
	if analyzer.config.Output.Format != "json" {
		t.Errorf("Expected updated format json, got %s", analyzer.config.Output.Format)
	}
}

func TestSetOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	var buffer1 bytes.Buffer
	analyzer := NewComplexityAnalyzer(cfg, &buffer1)
	
	var buffer2 bytes.Buffer
	analyzer.SetOutput(&buffer2)
	
	// Test that output goes to new buffer
	cfgs := createTestCFGs()[:1] // Just one function for simplicity
	err := analyzer.AnalyzeAndReport(cfgs)
	
	if err != nil {
		t.Fatalf("Failed to analyze and report: %v", err)
	}
	
	// Original buffer should be empty, new buffer should have content
	if buffer1.Len() > 0 {
		t.Error("Output should not go to original buffer")
	}
	if buffer2.Len() == 0 {
		t.Error("Output should go to new buffer")
	}
}

func TestGenerateReport(t *testing.T) {
	var buffer bytes.Buffer
	analyzer := NewComplexityAnalyzerWithDefaults(&buffer)
	
	cfgs := createTestCFGs()
	report := analyzer.GenerateReport(cfgs)
	
	if report == nil {
		t.Fatal("Expected report, got nil")
	}
	
	// Verify report structure
	if report.Summary.TotalFunctions != 3 {
		t.Errorf("Expected 3 total functions, got %d", report.Summary.TotalFunctions)
	}
	if len(report.Results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(report.Results))
	}
	
	// Verify summary calculations
	expectedAvg := (1.0 + 2.0 + 6.0) / 3.0 // 3.0
	if report.Summary.AverageComplexity != expectedAvg {
		t.Errorf("Expected average complexity %.1f, got %.1f", expectedAvg, report.Summary.AverageComplexity)
	}
	if report.Summary.MaxComplexity != 6 {
		t.Errorf("Expected max complexity 6, got %d", report.Summary.MaxComplexity)
	}
	if report.Summary.MinComplexity != 1 {
		t.Errorf("Expected min complexity 1, got %d", report.Summary.MinComplexity)
	}
}

func TestComplexityAnalyzerWithFiltering(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Complexity.ReportUnchanged = false // Filter out complexity 1 functions
	cfg.Output.MinComplexity = 2            // Additional filtering
	
	var buffer bytes.Buffer
	analyzer := NewComplexityAnalyzer(cfg, &buffer)
	
	cfgs := createTestCFGs()
	results := analyzer.AnalyzeFunctions(cfgs)
	
	// Should filter out simple_function (complexity 1)
	// Should include medium_function (complexity 2) and complex_function (complexity 6)
	if len(results) != 2 {
		t.Errorf("Expected 2 results after filtering, got %d", len(results))
	}
	
	for _, result := range results {
		if result.Complexity < 2 {
			t.Errorf("Function with complexity %d should have been filtered", result.Complexity)
		}
	}
}

func TestComplexityAnalyzerErrorHandling(t *testing.T) {
	t.Run("EmptyCFGList", func(t *testing.T) {
		var buffer bytes.Buffer
		analyzer := NewComplexityAnalyzerWithDefaults(&buffer)
		
		err := analyzer.AnalyzeAndReport([]*CFG{})
		
		if err != nil {
			t.Errorf("Should handle empty CFG list gracefully, got error: %v", err)
		}
		
		// Should generate valid report with zero functions
		output := buffer.String()
		if !strings.Contains(output, "Total Functions: 0") {
			t.Error("Should report zero functions for empty input")
		}
	})
	
	t.Run("NilCFGs", func(t *testing.T) {
		var buffer bytes.Buffer
		analyzer := NewComplexityAnalyzerWithDefaults(&buffer)
		
		cfgs := []*CFG{nil, createTestCFGs()[0], nil}
		results := analyzer.AnalyzeFunctions(cfgs)
		
		// Should ignore nil CFGs and process valid one
		if len(results) != 1 {
			t.Errorf("Expected 1 result (nil CFGs ignored), got %d", len(results))
		}
	})
}