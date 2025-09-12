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

	// CloneDetection holds clone detection configuration
	// DEPRECATED: Use CloneConfig directly instead
	CloneDetection CloneDetectionConfig `mapstructure:"clone_detection" yaml:"clone_detection"`

	// Clones holds the unified clone detection configuration
	Clones *CloneConfig `mapstructure:"clones" yaml:"clones"`

	// Output holds output formatting configuration
	Output OutputConfig `mapstructure:"output" yaml:"output"`

	// Analysis holds general analysis configuration
    Analysis AnalysisConfig `mapstructure:"analysis" yaml:"analysis"`

    // Architecture holds system-level layer configuration (optional)
    Architecture *ArchitectureConfig `mapstructure:"architecture" yaml:"architecture"`
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

    // Directory specifies the output directory for reports (empty = tool default, e.g., ".pyscn/reports" under current working directory)
    Directory string `mapstructure:"directory" yaml:"directory"`
}

// ArchitectureConfig defines layers and dependency rules for system-level validation
type ArchitectureConfig struct {
    Enabled bool `mapstructure:"enabled" yaml:"enabled"`
    Layers  []ArchitectureLayer `mapstructure:"layers" yaml:"layers"`
    Rules   []ArchitectureRule  `mapstructure:"rules" yaml:"rules"`
}

// ArchitectureLayer maps dotted-module patterns to a logical layer name
type ArchitectureLayer struct {
    Name     string   `mapstructure:"name" yaml:"name"`
    Packages []string `mapstructure:"packages" yaml:"packages"`
}

// ArchitectureRule defines allowed outbound dependencies for a given layer
type ArchitectureRule struct {
    From  string   `mapstructure:"from" yaml:"from"`
    Allow []string `mapstructure:"allow" yaml:"allow"`
}

// Validate checks architecture config consistency
func (a *ArchitectureConfig) Validate() error {
    if len(a.Layers) == 0 {
        return fmt.Errorf("no layers defined")
    }
    // unique layer names
    names := make(map[string]struct{})
    for _, l := range a.Layers {
        if l.Name == "" {
            return fmt.Errorf("layer name cannot be empty")
        }
        if _, ok := names[l.Name]; ok {
            return fmt.Errorf("duplicate layer name: %s", l.Name)
        }
        names[l.Name] = struct{}{}
        if len(l.Packages) == 0 {
            return fmt.Errorf("layer %s has no packages", l.Name)
        }
    }
    // rules reference known layers
    for _, r := range a.Rules {
        if r.From == "" {
            return fmt.Errorf("rule has empty 'from'")
        }
        if _, ok := names[r.From]; !ok {
            return fmt.Errorf("rule 'from' references unknown layer: %s", r.From)
        }
        for _, al := range r.Allow {
            if _, ok := names[al]; !ok {
                return fmt.Errorf("rule from %s allows unknown layer: %s", r.From, al)
            }
        }
    }
    return nil
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
	DetectAfterReturn         bool `mapstructure:"detect_after_return" yaml:"detect_after_return"`
	DetectAfterBreak          bool `mapstructure:"detect_after_break" yaml:"detect_after_break"`
	DetectAfterContinue       bool `mapstructure:"detect_after_continue" yaml:"detect_after_continue"`
	DetectAfterRaise          bool `mapstructure:"detect_after_raise" yaml:"detect_after_raise"`
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

// CloneDetectionConfig holds configuration for code clone detection
// DEPRECATED: Use CloneConfig directly instead
type CloneDetectionConfig struct {
	// Enabled controls whether clone detection is performed
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`

	// Minimum requirements for clone candidates
	MinLines int `mapstructure:"min_lines" yaml:"min_lines"`
	MinNodes int `mapstructure:"min_nodes" yaml:"min_nodes"`

	// Similarity thresholds for different clone types
	Type1Threshold float64 `mapstructure:"type1_threshold" yaml:"type1_threshold"` // Identical clones
	Type2Threshold float64 `mapstructure:"type2_threshold" yaml:"type2_threshold"` // Renamed clones
	Type3Threshold float64 `mapstructure:"type3_threshold" yaml:"type3_threshold"` // Near-miss clones
	Type4Threshold float64 `mapstructure:"type4_threshold" yaml:"type4_threshold"` // Semantic clones

	// General similarity threshold
	SimilarityThreshold float64 `mapstructure:"similarity_threshold" yaml:"similarity_threshold"`
	MaxEditDistance     float64 `mapstructure:"max_edit_distance" yaml:"max_edit_distance"`

	// Cost model configuration
	CostModelType     string `mapstructure:"cost_model_type" yaml:"cost_model_type"` // "default", "python", "weighted"
	IgnoreLiterals    bool   `mapstructure:"ignore_literals" yaml:"ignore_literals"`
	IgnoreIdentifiers bool   `mapstructure:"ignore_identifiers" yaml:"ignore_identifiers"`

	// Output configuration
	ShowContent bool   `mapstructure:"show_content" yaml:"show_content"`
	GroupClones bool   `mapstructure:"group_clones" yaml:"group_clones"`
	SortBy      string `mapstructure:"sort_by" yaml:"sort_by"` // "similarity", "size", "location", "type"

	// Filtering
	MinSimilarity float64  `mapstructure:"min_similarity" yaml:"min_similarity"`
	MaxSimilarity float64  `mapstructure:"max_similarity" yaml:"max_similarity"`
	CloneTypes    []string `mapstructure:"clone_types" yaml:"clone_types"` // ["type1", "type2", "type3", "type4"]
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
    config := &Config{
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
		// Use unified clone configuration
		Clones: DefaultCloneConfig(),
		Output: OutputConfig{
			Format:        "text",
			ShowDetails:   false,
			SortBy:        "name",
			MinComplexity: DefaultMinComplexityFilter,
		},
        Analysis: AnalysisConfig{
            IncludePatterns: []string{"*.py"},
            ExcludePatterns: []string{"test_*.py", "*_test.py"},
            Recursive:       true,
            FollowSymlinks:  false,
        },
        Architecture: nil,
    }

	// For backward compatibility, populate legacy CloneDetection field
	config.CloneDetection = config.Clones.ToCloneDetectionConfig()

	return config
}

// LoadConfig loads configuration from file or returns default config
func LoadConfig(configPath string) (*Config, error) {
	return LoadConfigWithTarget(configPath, "")
}

// discoverConfigFile finds the appropriate config file path
// Single responsibility: configuration file discovery only
func discoverConfigFile(targetPath string) string {
	return findDefaultConfig(targetPath)
}

// loadConfigFromFile reads and parses a configuration file
// Single responsibility: file loading and parsing only
func loadConfigFromFile(configPath string) (*Config, error) {
	if configPath == "" {
		return DefaultConfig(), nil
	}

	// Create a new viper instance to avoid race conditions
	v := viper.New()
	config := DefaultConfig()
	v.SetConfigFile(configPath)

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Unmarshal into config struct
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// LoadConfigWithTarget loads configuration with target path context
// Orchestrates discovery and loading but delegates specific concerns
func LoadConfigWithTarget(configPath string, targetPath string) (*Config, error) {
	// If no config path specified, discover one
	if configPath == "" {
		configPath = discoverConfigFile(targetPath)
	}

	// Load the configuration from the determined path
	return loadConfigFromFile(configPath)
}

// searchConfigInDirectory searches for configuration files in a specific directory
func searchConfigInDirectory(dir string, candidates []string) string {
	for _, candidate := range candidates {
		path := filepath.Join(dir, candidate)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

// findDefaultConfig looks for default configuration files in common locations
// targetPath is the path being analyzed (e.g., the Python file or directory)
func findDefaultConfig(targetPath string) string {
	candidates := []string{
		"pyscn.yaml",
		"pyscn.yml",
		".pyscn.toml",
		".pyscn.yml",
		"pyscn.json",
		".pyscn.json",
	}

	// If targetPath is provided, search from there upward
	if targetPath != "" {
		// Convert to absolute path
		absPath, err := filepath.Abs(targetPath)
		if err == nil {
			// If it's a file, start from its directory
			info, err := os.Stat(absPath)
			if err == nil && !info.IsDir() {
				absPath = filepath.Dir(absPath)
			}
			
			// Search from target directory up to root with robust termination
			// Handle Windows edge cases: volume roots (C:\), UNC paths (\\server\share), long paths
			volume := filepath.VolumeName(absPath)
			for dir := absPath; ; dir = filepath.Dir(dir) {
				if config := searchConfigInDirectory(dir, candidates); config != "" {
					return config
				}
				
				// Robust termination conditions for cross-platform compatibility
				parent := filepath.Dir(dir)
				if parent == dir || // Unix-style root reached (/), Windows UNC root (\\server)  
					dir == volume || // Windows volume root reached (C:\)
					(volume != "" && dir == volume+string(filepath.Separator)) { // Alternative volume root format
					break
				}
			}
		}
	}
	
	// Fallback to current directory
	if config := searchConfigInDirectory(".", candidates); config != "" {
		return config
	}

	// Check XDG config directory (Linux/Mac standard)
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		if config := searchConfigInDirectory(filepath.Join(xdgConfig, "pyscn"), candidates); config != "" {
			return config
		}
	}
	
	// Check ~/.config/pyscn/ (XDG default)
	if home, err := os.UserHomeDir(); err == nil {
		configDir := filepath.Join(home, ".config", "pyscn")
		if config := searchConfigInDirectory(configDir, candidates); config != "" {
			return config
		}
		
		// Check home directory (backward compatibility)
		if config := searchConfigInDirectory(home, candidates); config != "" {
			return config
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
		"html": true,
	}

    if !validFormats[c.Output.Format] {
        return fmt.Errorf("invalid output.format '%s', must be one of: text, json, yaml, csv, html", c.Output.Format)
    }

    // Validate architecture configuration if present
    if c.Architecture != nil && c.Architecture.Enabled {
        if err := c.Architecture.Validate(); err != nil {
            return fmt.Errorf("architecture: %w", err)
        }
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

	// Validate clone detection configuration
	if err := c.validateCloneDetectionConfig(); err != nil {
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
	// Create a new viper instance to avoid race conditions
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// Set all config values in viper
	v.Set("complexity", config.Complexity)
	v.Set("dead_code", config.DeadCode)
	v.Set("output", config.Output)
	v.Set("analysis", config.Analysis)

	return v.WriteConfig()
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

// validateCloneDetectionConfig validates the clone detection configuration
func (c *Config) validateCloneDetectionConfig() error {
	// Validate minimum requirements
	if c.CloneDetection.MinLines < 1 {
		return fmt.Errorf("clone_detection.min_lines must be >= 1, got %d", c.CloneDetection.MinLines)
	}

	if c.CloneDetection.MinNodes < 1 {
		return fmt.Errorf("clone_detection.min_nodes must be >= 1, got %d", c.CloneDetection.MinNodes)
	}

	// Validate thresholds
	if c.CloneDetection.Type1Threshold < 0.0 || c.CloneDetection.Type1Threshold > 1.0 {
		return fmt.Errorf("clone_detection.type1_threshold must be between 0.0 and 1.0, got %f", c.CloneDetection.Type1Threshold)
	}

	if c.CloneDetection.Type2Threshold < 0.0 || c.CloneDetection.Type2Threshold > 1.0 {
		return fmt.Errorf("clone_detection.type2_threshold must be between 0.0 and 1.0, got %f", c.CloneDetection.Type2Threshold)
	}

	if c.CloneDetection.Type3Threshold < 0.0 || c.CloneDetection.Type3Threshold > 1.0 {
		return fmt.Errorf("clone_detection.type3_threshold must be between 0.0 and 1.0, got %f", c.CloneDetection.Type3Threshold)
	}

	if c.CloneDetection.Type4Threshold < 0.0 || c.CloneDetection.Type4Threshold > 1.0 {
		return fmt.Errorf("clone_detection.type4_threshold must be between 0.0 and 1.0, got %f", c.CloneDetection.Type4Threshold)
	}

	// Validate threshold ordering (Type1 > Type2 > Type3 > Type4)
	if c.CloneDetection.Type1Threshold <= c.CloneDetection.Type2Threshold {
		return fmt.Errorf("clone_detection.type1_threshold (%f) should be > type2_threshold (%f)",
			c.CloneDetection.Type1Threshold, c.CloneDetection.Type2Threshold)
	}

	if c.CloneDetection.Type2Threshold <= c.CloneDetection.Type3Threshold {
		return fmt.Errorf("clone_detection.type2_threshold (%f) should be > type3_threshold (%f)",
			c.CloneDetection.Type2Threshold, c.CloneDetection.Type3Threshold)
	}

	if c.CloneDetection.Type3Threshold <= c.CloneDetection.Type4Threshold {
		return fmt.Errorf("clone_detection.type3_threshold (%f) should be > type4_threshold (%f)",
			c.CloneDetection.Type3Threshold, c.CloneDetection.Type4Threshold)
	}

	// Validate similarity threshold
	if c.CloneDetection.SimilarityThreshold < 0.0 || c.CloneDetection.SimilarityThreshold > 1.0 {
		return fmt.Errorf("clone_detection.similarity_threshold must be between 0.0 and 1.0, got %f", c.CloneDetection.SimilarityThreshold)
	}

	// Validate max edit distance
	if c.CloneDetection.MaxEditDistance < 0.0 {
		return fmt.Errorf("clone_detection.max_edit_distance must be >= 0.0, got %f", c.CloneDetection.MaxEditDistance)
	}

	// Validate cost model type
	validCostModels := map[string]bool{
		"default":  true,
		"python":   true,
		"weighted": true,
	}

	if !validCostModels[c.CloneDetection.CostModelType] {
		return fmt.Errorf("invalid clone_detection.cost_model_type '%s', must be one of: default, python, weighted", c.CloneDetection.CostModelType)
	}

	// Validate sort criteria
	validSortBy := map[string]bool{
		"similarity": true,
		"size":       true,
		"location":   true,
		"type":       true,
	}

	if !validSortBy[c.CloneDetection.SortBy] {
		return fmt.Errorf("invalid clone_detection.sort_by '%s', must be one of: similarity, size, location, type", c.CloneDetection.SortBy)
	}

	// Validate similarity range
	if c.CloneDetection.MinSimilarity < 0.0 || c.CloneDetection.MinSimilarity > 1.0 {
		return fmt.Errorf("clone_detection.min_similarity must be between 0.0 and 1.0, got %f", c.CloneDetection.MinSimilarity)
	}

	if c.CloneDetection.MaxSimilarity < 0.0 || c.CloneDetection.MaxSimilarity > 1.0 {
		return fmt.Errorf("clone_detection.max_similarity must be between 0.0 and 1.0, got %f", c.CloneDetection.MaxSimilarity)
	}

	if c.CloneDetection.MinSimilarity > c.CloneDetection.MaxSimilarity {
		return fmt.Errorf("clone_detection.min_similarity (%f) cannot be greater than max_similarity (%f)",
			c.CloneDetection.MinSimilarity, c.CloneDetection.MaxSimilarity)
	}

	// Validate clone types
	validCloneTypes := map[string]bool{
		"type1": true,
		"type2": true,
		"type3": true,
		"type4": true,
	}

	for _, cloneType := range c.CloneDetection.CloneTypes {
		if !validCloneTypes[cloneType] {
			return fmt.Errorf("invalid clone type '%s' in clone_detection.clone_types, must be one of: type1, type2, type3, type4", cloneType)
		}
	}

	return nil
}

// ShouldDetectClones determines if clone detection should be performed
func (c *CloneDetectionConfig) ShouldDetectClones() bool {
	return c.Enabled
}

// GetEnabledCloneTypes returns the enabled clone types as a slice
func (c *CloneDetectionConfig) GetEnabledCloneTypes() []string {
	return c.CloneTypes
}

// IsCloneTypeEnabled checks if a specific clone type is enabled
func (c *CloneDetectionConfig) IsCloneTypeEnabled(cloneType string) bool {
	for _, enabledType := range c.CloneTypes {
		if enabledType == cloneType {
			return true
		}
	}
	return false
}
