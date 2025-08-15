package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Test complexity defaults
	if config.Complexity.LowThreshold != 9 {
		t.Errorf("Expected low threshold 9, got %d", config.Complexity.LowThreshold)
	}
	if config.Complexity.MediumThreshold != 19 {
		t.Errorf("Expected medium threshold 19, got %d", config.Complexity.MediumThreshold)
	}
	if !config.Complexity.Enabled {
		t.Error("Expected complexity analysis to be enabled by default")
	}
	if !config.Complexity.ReportUnchanged {
		t.Error("Expected report_unchanged to be true by default for backward compatibility")
	}
	if config.Complexity.MaxComplexity != 0 {
		t.Errorf("Expected max complexity 0 (no limit), got %d", config.Complexity.MaxComplexity)
	}

	// Test output defaults
	if config.Output.Format != "text" {
		t.Errorf("Expected format 'text', got %s", config.Output.Format)
	}
	if config.Output.ShowDetails {
		t.Error("Expected show_details to be false by default")
	}
	if config.Output.SortBy != "name" {
		t.Errorf("Expected sort_by 'name', got %s", config.Output.SortBy)
	}
	if config.Output.MinComplexity != 1 {
		t.Errorf("Expected min complexity 1, got %d", config.Output.MinComplexity)
	}

	// Test analysis defaults
	if len(config.Analysis.IncludePatterns) != 1 || config.Analysis.IncludePatterns[0] != "*.py" {
		t.Errorf("Expected include patterns ['*.py'], got %v", config.Analysis.IncludePatterns)
	}
	if len(config.Analysis.ExcludePatterns) != 3 {
		t.Errorf("Expected 3 exclude patterns, got %d", len(config.Analysis.ExcludePatterns))
	}
	if !config.Analysis.Recursive {
		t.Error("Expected recursive to be true by default")
	}
	if config.Analysis.FollowSymlinks {
		t.Error("Expected follow_symlinks to be false by default")
	}
}

func TestConfigValidation(t *testing.T) {
	testCases := []struct {
		name        string
		modifyConfig func(*Config)
		expectError bool
		errorContains string
	}{
		{
			name: "ValidConfig",
			modifyConfig: func(c *Config) {
				// Default config should be valid
			},
			expectError: false,
		},
		{
			name: "InvalidLowThreshold",
			modifyConfig: func(c *Config) {
				c.Complexity.LowThreshold = 0
			},
			expectError: true,
			errorContains: "low_threshold must be >= 1",
		},
		{
			name: "InvalidMediumThreshold",
			modifyConfig: func(c *Config) {
				c.Complexity.LowThreshold = 10
				c.Complexity.MediumThreshold = 10
			},
			expectError: true,
			errorContains: "medium_threshold (10) must be > low_threshold (10)",
		},
		{
			name: "InvalidMaxComplexity",
			modifyConfig: func(c *Config) {
				c.Complexity.MaxComplexity = -1
			},
			expectError: true,
			errorContains: "max_complexity must be >= 0",
		},
		{
			name: "MaxComplexityTooLow",
			modifyConfig: func(c *Config) {
				c.Complexity.MaxComplexity = 15
				c.Complexity.MediumThreshold = 19
			},
			expectError: true,
			errorContains: "max_complexity (15) must be > medium_threshold (19)",
		},
		{
			name: "InvalidOutputFormat",
			modifyConfig: func(c *Config) {
				c.Output.Format = "invalid"
			},
			expectError: true,
			errorContains: "invalid output.format 'invalid'",
		},
		{
			name: "InvalidSortBy",
			modifyConfig: func(c *Config) {
				c.Output.SortBy = "invalid"
			},
			expectError: true,
			errorContains: "invalid output.sort_by 'invalid'",
		},
		{
			name: "InvalidMinComplexity",
			modifyConfig: func(c *Config) {
				c.Output.MinComplexity = 0
			},
			expectError: true,
			errorContains: "min_complexity must be >= 1",
		},
		{
			name: "EmptyIncludePatterns",
			modifyConfig: func(c *Config) {
				c.Analysis.IncludePatterns = []string{}
			},
			expectError: true,
			errorContains: "include_patterns cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := DefaultConfig()
			tc.modifyConfig(config)

			err := config.Validate()

			if tc.expectError {
				if err == nil {
					t.Error("Expected validation error, but got none")
				} else if tc.errorContains != "" && !containsString(err.Error(), tc.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tc.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, got: %v", err)
				}
			}
		})
	}
}

func TestComplexityConfigMethods(t *testing.T) {
	config := DefaultConfig()

	t.Run("AssessRiskLevel", func(t *testing.T) {
		testCases := []struct {
			complexity int
			expected   string
		}{
			{1, "low"},
			{5, "low"},
			{9, "low"},
			{10, "medium"},
			{15, "medium"},
			{19, "medium"},
			{20, "high"},
			{50, "high"},
		}

		for _, tc := range testCases {
			result := config.Complexity.AssessRiskLevel(tc.complexity)
			if result != tc.expected {
				t.Errorf("For complexity %d, expected risk '%s', got '%s'", 
					tc.complexity, tc.expected, result)
			}
		}
	})

	t.Run("ShouldReport", func(t *testing.T) {
		testCases := []struct {
			name           string
			enabled        bool
			reportUnchanged bool
			complexity     int
			expected       bool
		}{
			{"DisabledAnalysis", false, false, 5, false},
			{"EnabledAnalysis", true, false, 5, true},
			{"UnchangedNotReported", true, false, 1, false},
			{"UnchangedReported", true, true, 1, true},
			{"HighComplexity", true, false, 20, true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				config.Complexity.Enabled = tc.enabled
				config.Complexity.ReportUnchanged = tc.reportUnchanged

				result := config.Complexity.ShouldReport(tc.complexity)
				if result != tc.expected {
					t.Errorf("Expected should report %t, got %t", tc.expected, result)
				}
			})
		}
	})

	t.Run("ExceedsMaxComplexity", func(t *testing.T) {
		testCases := []struct {
			name         string
			maxComplexity int
			complexity   int
			expected     bool
		}{
			{"NoLimit", 0, 100, false},
			{"WithinLimit", 20, 15, false},
			{"AtLimit", 20, 20, false},
			{"ExceedsLimit", 20, 25, true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				config.Complexity.MaxComplexity = tc.maxComplexity

				result := config.Complexity.ExceedsMaxComplexity(tc.complexity)
				if result != tc.expected {
					t.Errorf("Expected exceeds max %t, got %t", tc.expected, result)
				}
			})
		}
	})
}

func TestLoadConfig(t *testing.T) {
	t.Run("LoadNonExistentConfig", func(t *testing.T) {
		config, err := LoadConfig("nonexistent.yaml")
		if err == nil {
			t.Error("Expected error for non-existent config file")
		}
		if config != nil {
			t.Error("Expected nil config for non-existent file")
		}
	})

	t.Run("LoadEmptyPath", func(t *testing.T) {
		config, err := LoadConfig("")
		if err != nil {
			t.Errorf("Expected no error for empty path, got: %v", err)
		}
		if config == nil {
			t.Error("Expected default config for empty path")
		}

		// Should return default config
		defaultConfig := DefaultConfig()
		if config.Complexity.LowThreshold != defaultConfig.Complexity.LowThreshold {
			t.Error("Expected default config values")
		}
	})

	t.Run("LoadValidYAMLConfig", func(t *testing.T) {
		// Create temporary config file
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test_config.yaml")

		yamlContent := `
complexity:
  low_threshold: 5
  medium_threshold: 15
  enabled: true
  report_unchanged: true
  max_complexity: 50

output:
  format: json
  show_details: true
  sort_by: complexity
  min_complexity: 2

analysis:
  include_patterns:
    - "*.py"
    - "*.pyx"
  exclude_patterns:
    - "*test*.py"
  recursive: true
  follow_symlinks: false
`

		err := os.WriteFile(configPath, []byte(yamlContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test config file: %v", err)
		}

		config, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// Verify loaded values
		if config.Complexity.LowThreshold != 5 {
			t.Errorf("Expected low threshold 5, got %d", config.Complexity.LowThreshold)
		}
		if config.Complexity.MediumThreshold != 15 {
			t.Errorf("Expected medium threshold 15, got %d", config.Complexity.MediumThreshold)
		}
		if !config.Complexity.ReportUnchanged {
			t.Error("Expected report_unchanged to be true")
		}
		if config.Complexity.MaxComplexity != 50 {
			t.Errorf("Expected max complexity 50, got %d", config.Complexity.MaxComplexity)
		}
		if config.Output.Format != "json" {
			t.Errorf("Expected format json, got %s", config.Output.Format)
		}
		if !config.Output.ShowDetails {
			t.Error("Expected show_details to be true")
		}
		if config.Output.SortBy != "complexity" {
			t.Errorf("Expected sort_by complexity, got %s", config.Output.SortBy)
		}
		if len(config.Analysis.IncludePatterns) != 2 {
			t.Errorf("Expected 2 include patterns, got %d", len(config.Analysis.IncludePatterns))
		}
	})

	t.Run("LoadInvalidYAMLConfig", func(t *testing.T) {
		// Create temporary invalid config file
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "invalid_config.yaml")

		yamlContent := `
complexity:
  low_threshold: 0  # Invalid: must be >= 1
  medium_threshold: 15
`

		err := os.WriteFile(configPath, []byte(yamlContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test config file: %v", err)
		}

		config, err := LoadConfig(configPath)
		if err == nil {
			t.Error("Expected validation error for invalid config")
		}
		if config != nil {
			t.Error("Expected nil config for invalid file")
		}
	})
}

func TestSaveConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "saved_config.yaml")

	config := DefaultConfig()
	config.Complexity.LowThreshold = 7
	config.Output.Format = "json"

	err := SaveConfig(config, configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Load the saved config and verify
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedConfig.Complexity.LowThreshold != 7 {
		t.Errorf("Expected saved low threshold 7, got %d", loadedConfig.Complexity.LowThreshold)
	}
	if loadedConfig.Output.Format != "json" {
		t.Errorf("Expected saved format json, got %s", loadedConfig.Output.Format)
	}
}

func TestFindDefaultConfig(t *testing.T) {
	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	t.Run("NoDefaultConfigFound", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Chdir(tempDir)

		result := findDefaultConfig()
		if result != "" {
			t.Errorf("Expected empty result for no config files, got %s", result)
		}
	})

	t.Run("FindDefaultConfigInCurrentDir", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Chdir(tempDir)

		// Create a default config file
		configPath := filepath.Join(tempDir, "pyqol.yaml")
		err := os.WriteFile(configPath, []byte("complexity:\n  enabled: true"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test config: %v", err)
		}

		result := findDefaultConfig()
		if result != "pyqol.yaml" {
			t.Errorf("Expected to find pyqol.yaml, got %s", result)
		}
	})
}

func TestConfigRaceConditions(t *testing.T) {
	// Test concurrent access to config
	config := DefaultConfig()

	done := make(chan bool, 10)

	// Start multiple goroutines accessing config
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Read operations
			_ = config.Complexity.AssessRiskLevel(10)
			_ = config.Complexity.ShouldReport(5)
			_ = config.Complexity.ExceedsMaxComplexity(25)
			_ = config.Validate()
		}()
	}

	// Wait for all goroutines with timeout
	timeout := time.After(5 * time.Second)
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// Goroutine completed
		case <-timeout:
			t.Fatal("Test timed out - potential deadlock")
		}
	}
}

func TestConfigEdgeCases(t *testing.T) {
	t.Run("ExtremThresholds", func(t *testing.T) {
		config := DefaultConfig()
		config.Complexity.LowThreshold = 1
		config.Complexity.MediumThreshold = 2
		config.Complexity.MaxComplexity = 1000

		err := config.Validate()
		if err != nil {
			t.Errorf("Valid extreme thresholds should not cause error: %v", err)
		}

		// Test edge cases
		if config.Complexity.AssessRiskLevel(1) != "low" {
			t.Error("Boundary case failed for low risk")
		}
		if config.Complexity.AssessRiskLevel(2) != "medium" {
			t.Error("Boundary case failed for medium risk")
		}
		if config.Complexity.AssessRiskLevel(3) != "high" {
			t.Error("Boundary case failed for high risk")
		}
	})

	t.Run("AllFormats", func(t *testing.T) {
		config := DefaultConfig()
		validFormats := []string{"text", "json", "yaml", "csv"}

		for _, format := range validFormats {
			config.Output.Format = format
			err := config.Validate()
			if err != nil {
				t.Errorf("Format %s should be valid: %v", format, err)
			}
		}
	})

	t.Run("AllSortOptions", func(t *testing.T) {
		config := DefaultConfig()
		validSortOptions := []string{"name", "complexity", "risk"}

		for _, sortBy := range validSortOptions {
			config.Output.SortBy = sortBy
			err := config.Validate()
			if err != nil {
				t.Errorf("Sort option %s should be valid: %v", sortBy, err)
			}
		}
	})
}

// Helper function to check if a string contains a substring
func containsString(str, substr string) bool {
	return len(str) >= len(substr) && 
		   (len(substr) == 0 || 
		    (len(str) > 0 && 
		     (str[:len(substr)] == substr || containsString(str[1:], substr))))
}