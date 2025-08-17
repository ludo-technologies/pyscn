package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Default complexity thresholds based on McCabe complexity standards
const (
	// DefaultLowComplexityThreshold defines the upper bound for low complexity functions
	// Functions with complexity <= 9 are considered low risk and easy to maintain
	DefaultLowComplexityThreshold = 9

	// DefaultMediumComplexityThreshold defines the upper bound for medium complexity functions
	// Functions with complexity 10-19 are considered medium risk and may need refactoring
	DefaultMediumComplexityThreshold = 19

	// DefaultMinComplexityFilter defines the minimum complexity to report
	// Functions with complexity >= 1 will be included in reports
	DefaultMinComplexityFilter = 1

	// DefaultMaxComplexityLimit defines no upper limit for complexity analysis
	// Setting to 0 means no maximum complexity enforcement
	DefaultMaxComplexityLimit = 0
)

// Default dead code detection settings
const (
	// DefaultDeadCodeMinSeverity defines the minimum severity level to report
	DefaultDeadCodeMinSeverity = "warning"

	// DefaultDeadCodeContextLines defines the number of context lines to show
	DefaultDeadCodeContextLines = 3

	// DefaultDeadCodeSortBy defines the default sorting criteria
	DefaultDeadCodeSortBy = "severity"
)

// Config represents the main configuration structure
type Config struct {
	// Complexity holds complexity analysis configuration
	Complexity ComplexityConfig `mapstructure:"complexity" yaml:"complexity"`

	// DeadCode holds dead code detection configuration
	DeadCode DeadCodeConfig `mapstructure:"dead_code" yaml:"dead_code"`

	// Output holds output formatting configuration
	Output OutputConfig `mapstructure:"output" yaml:"output"`

	// Analysis holds general analysis configuration
	Analysis AnalysisConfig `mapstructure:"analysis" yaml:"analysis"`
}

// ComplexityConfig holds configuration for cyclomatic complexity analysis
type ComplexityConfig struct {
	// LowThreshold is the upper bound for low complexity (inclusive)
	LowThreshold int `mapstructure:"low_threshold" yaml:"low_threshold"`

	// MediumThreshold is the upper bound for medium complexity (inclusive)
	// Values above this are considered high complexity
	MediumThreshold int `mapstructure:"medium_threshold" yaml:"medium_threshold"`

	// Enabled controls whether complexity analysis is performed
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`

	// ReportUnchanged controls whether to report functions with complexity = 1
	ReportUnchanged bool `mapstructure:"report_unchanged" yaml:"report_unchanged"`

	// MaxComplexity is the maximum allowed complexity before failing analysis
	// 0 means no limit
	MaxComplexity int `mapstructure:"max_complexity" yaml:"max_complexity"`
}

// OutputConfig holds configuration for output formatting
type OutputConfig struct {
	// Format specifies the output format: json, yaml, text, csv
	Format string `mapstructure:"format" yaml:"format"`

	// ShowDetails controls whether to show detailed breakdown
	ShowDetails bool `mapstructure:"show_details" yaml:"show_details"`

	// SortBy specifies how to sort results: name, complexity, risk
	SortBy string `mapstructure:"sort_by" yaml:"sort_by"`

	// MinComplexity is the minimum complexity to report (filters low values)
	MinComplexity int `mapstructure:"min_complexity" yaml:"min_complexity"`
}

// DeadCodeConfig holds configuration for dead code detection
type DeadCodeConfig struct {
	// Enabled controls whether dead code detection is performed
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`

	// MinSeverity is the minimum severity level to report
	MinSeverity string `mapstructure:"min_severity" yaml:"min_severity"`

	// ShowContext controls whether to show surrounding code context
	ShowContext bool `mapstructure:"show_context" yaml:"show_context"`

	// ContextLines is the number of context lines to show around dead code
	ContextLines int `mapstructure:"context_lines" yaml:"context_lines"`

	// SortBy specifies how to sort results: severity, line, file, function
	SortBy string `mapstructure:"sort_by" yaml:"sort_by"`

	// Detection options
	DetectAfterReturn bool `mapstructure:"detect_after_return" yaml:"detect_after_return"`
	DetectAfterBreak bool `mapstructure:"detect_after_break" yaml:"detect_after_break"`
	DetectAfterContinue bool `mapstructure:"detect_after_continue" yaml:"detect_after_continue"`
	DetectAfterRaise bool `mapstructure:"detect_after_raise" yaml:"detect_after_raise"`
	DetectUnreachableBranches bool `mapstructure:"detect_unreachable_branches" yaml:"detect_unreachable_branches"`

	// IgnorePatterns specifies patterns for code to ignore (e.g., comments, debug code)
	IgnorePatterns []string `mapstructure:"ignore_patterns" yaml:"ignore_patterns"`
}

// AnalysisConfig holds general analysis configuration
type AnalysisConfig struct {
	// IncludePatterns specifies file patterns to include
	IncludePatterns []string `mapstructure:"include_patterns" yaml:"include_patterns"`

	// ExcludePatterns specifies file patterns to exclude
	ExcludePatterns []string `mapstructure:"exclude_patterns" yaml:"exclude_patterns"`

	// Recursive controls whether to analyze directories recursively
	Recursive bool `mapstructure:"recursive" yaml:"recursive"`

	// FollowSymlinks controls whether to follow symbolic links
	FollowSymlinks bool `mapstructure:"follow_symlinks" yaml:"follow_symlinks"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Complexity: ComplexityConfig{
			LowThreshold:    DefaultLowComplexityThreshold,
			MediumThreshold: DefaultMediumComplexityThreshold,
			Enabled:         true,
			ReportUnchanged: true,
			MaxComplexity:   DefaultMaxComplexityLimit,
		},
		DeadCode: DeadCodeConfig{
			Enabled:                   true,
			MinSeverity:               DefaultDeadCodeMinSeverity,
			ShowContext:               false,
			ContextLines:              DefaultDeadCodeContextLines,
			SortBy:                    DefaultDeadCodeSortBy,
			DetectAfterReturn:         true,
			DetectAfterBreak:          true,
			DetectAfterContinue:       true,
			DetectAfterRaise:          true,
			DetectUnreachableBranches: true,
			IgnorePatterns:            []string{},
		},
		Output: OutputConfig{
			Format:        "text",
			ShowDetails:   false,
			SortBy:        "name",
			MinComplexity: DefaultMinComplexityFilter,
		},
		Analysis: AnalysisConfig{
			IncludePatterns: []string{"*.py"},
			ExcludePatterns: []string{"*test*.py", "*_test.py", "test_*.py"},
			Recursive:       true,
			FollowSymlinks:  false,
		},
	}
}

// LoadConfig loads configuration from file or returns default config
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	// If no config path specified, try to find default config files
	if configPath == "" {
		configPath = findDefaultConfig()
	}

	// If still no config found, return default
	if configPath == "" {
		return config, nil
	}

	viper.SetConfigFile(configPath)

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Unmarshal into config struct
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// findDefaultConfig looks for default configuration files in common locations
func findDefaultConfig() string {
	candidates := []string{
		"pyqol.yaml",
		"pyqol.yml",
		".pyqol.yaml",
		".pyqol.yml",
		"pyqol.json",
		".pyqol.json",
	}

	// Check current directory first
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	// Check home directory
	if home, err := os.UserHomeDir(); err == nil {
		for _, candidate := range candidates {
			path := filepath.Join(home, candidate)
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	return ""
}

// Validate validates the configuration values
func (c *Config) Validate() error {
	// Validate complexity thresholds
	if c.Complexity.LowThreshold < 1 {
		return fmt.Errorf("complexity.low_threshold must be >= 1, got %d", c.Complexity.LowThreshold)
	}

	if c.Complexity.MediumThreshold <= c.Complexity.LowThreshold {
		return fmt.Errorf("complexity.medium_threshold (%d) must be > low_threshold (%d)",
			c.Complexity.MediumThreshold, c.Complexity.LowThreshold)
	}

	if c.Complexity.MaxComplexity < 0 {
		return fmt.Errorf("complexity.max_complexity must be >= 0, got %d", c.Complexity.MaxComplexity)
	}

	if c.Complexity.MaxComplexity > 0 && c.Complexity.MaxComplexity <= c.Complexity.MediumThreshold {
		return fmt.Errorf("complexity.max_complexity (%d) must be > medium_threshold (%d) or 0 for no limit",
			c.Complexity.MaxComplexity, c.Complexity.MediumThreshold)
	}

	// Validate output format
	validFormats := map[string]bool{
		"text": true,
		"json": true,
		"yaml": true,
		"csv":  true,
	}

	if !validFormats[c.Output.Format] {
		return fmt.Errorf("invalid output.format '%s', must be one of: text, json, yaml, csv", c.Output.Format)
	}

	// Validate sort options
	validSortBy := map[string]bool{
		"name":       true,
		"complexity": true,
		"risk":       true,
	}

	if !validSortBy[c.Output.SortBy] {
		return fmt.Errorf("invalid output.sort_by '%s', must be one of: name, complexity, risk", c.Output.SortBy)
	}

	if c.Output.MinComplexity < 1 {
		return fmt.Errorf("output.min_complexity must be >= 1, got %d", c.Output.MinComplexity)
	}

	// Validate include patterns (at least one must be specified)
	if len(c.Analysis.IncludePatterns) == 0 {
		return fmt.Errorf("analysis.include_patterns cannot be empty")
	}

	// Validate dead code configuration
	if err := c.validateDeadCodeConfig(); err != nil {
		return err
	}

	return nil
}

// AssessRiskLevel determines risk level based on complexity and thresholds
func (c *ComplexityConfig) AssessRiskLevel(complexity int) string {
	if complexity <= c.LowThreshold {
		return "low"
	} else if complexity <= c.MediumThreshold {
		return "medium"
	}
	return "high"
}

// ShouldReport determines if a complexity result should be reported
func (c *ComplexityConfig) ShouldReport(complexity int) bool {
	if !c.Enabled {
		return false
	}

	if complexity == 1 && !c.ReportUnchanged {
		return false
	}

	return true
}

// ExceedsMaxComplexity checks if complexity exceeds the maximum allowed
func (c *ComplexityConfig) ExceedsMaxComplexity(complexity int) bool {
	return c.MaxComplexity > 0 && complexity > c.MaxComplexity
}

// SaveConfig saves configuration to a YAML file
func SaveConfig(config *Config, path string) error {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	// Set all config values in viper
	viper.Set("complexity", config.Complexity)
	viper.Set("dead_code", config.DeadCode)
	viper.Set("output", config.Output)
	viper.Set("analysis", config.Analysis)

	return viper.WriteConfig()
}

// validateDeadCodeConfig validates the dead code configuration
func (c *Config) validateDeadCodeConfig() error {
	// Validate severity level
	validSeverities := map[string]bool{
		"critical": true,
		"warning":  true,
		"info":     true,
	}

	if !validSeverities[c.DeadCode.MinSeverity] {
		return fmt.Errorf("invalid dead_code.min_severity '%s', must be one of: critical, warning, info", c.DeadCode.MinSeverity)
	}

	// Validate context lines
	if c.DeadCode.ContextLines < 0 {
		return fmt.Errorf("dead_code.context_lines must be >= 0, got %d", c.DeadCode.ContextLines)
	}

	if c.DeadCode.ContextLines > 20 {
		return fmt.Errorf("dead_code.context_lines cannot exceed 20, got %d", c.DeadCode.ContextLines)
	}

	// Validate sort criteria
	validSortBy := map[string]bool{
		"severity": true,
		"line":     true,
		"file":     true,
		"function": true,
	}

	if !validSortBy[c.DeadCode.SortBy] {
		return fmt.Errorf("invalid dead_code.sort_by '%s', must be one of: severity, line, file, function", c.DeadCode.SortBy)
	}

	return nil
}

// ShouldDetectDeadCode determines if dead code detection should be performed
func (c *DeadCodeConfig) ShouldDetectDeadCode() bool {
	return c.Enabled
}

// GetMinSeverityLevel returns the minimum severity level as an integer for comparison
func (c *DeadCodeConfig) GetMinSeverityLevel() int {
	switch c.MinSeverity {
	case "info":
		return 1
	case "warning":
		return 2
	case "critical":
		return 3
	default:
		return 2 // Default to warning
	}
}

// HasAnyDetectionEnabled checks if any detection type is enabled
func (c *DeadCodeConfig) HasAnyDetectionEnabled() bool {
	return c.DetectAfterReturn ||
		c.DetectAfterBreak ||
		c.DetectAfterContinue ||
		c.DetectAfterRaise ||
		c.DetectUnreachableBranches
}
