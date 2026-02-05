package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/pelletier/go-toml/v2"
)

// Default complexity thresholds - re-exported from domain for backward compatibility
// All default values are defined in domain/defaults.go
const (
	// DefaultLowComplexityThreshold defines the upper bound for low complexity functions
	DefaultLowComplexityThreshold = domain.DefaultComplexityLowThreshold

	// DefaultMediumComplexityThreshold defines the upper bound for medium complexity functions
	DefaultMediumComplexityThreshold = domain.DefaultComplexityMediumThreshold

	// DefaultMinComplexityFilter defines the minimum complexity to report
	DefaultMinComplexityFilter = domain.DefaultComplexityMinFilter

	// DefaultMaxComplexityLimit defines no upper limit for complexity analysis
	DefaultMaxComplexityLimit = domain.DefaultComplexityMaxLimit
)

// Default dead code detection settings - re-exported from domain for backward compatibility
const (
	// DefaultDeadCodeMinSeverity defines the minimum severity level to report
	DefaultDeadCodeMinSeverity = domain.DefaultDeadCodeMinSeverity

	// DefaultDeadCodeContextLines defines the number of context lines to show
	DefaultDeadCodeContextLines = domain.DefaultDeadCodeContextLines

	// DefaultDeadCodeSortBy defines the default sorting criteria
	DefaultDeadCodeSortBy = domain.DefaultDeadCodeSortBy
)

// Config represents the main configuration structure
type Config struct {
	// Complexity holds complexity analysis configuration
	Complexity ComplexityConfig `mapstructure:"complexity" yaml:"complexity"`

	// DeadCode holds dead code detection configuration
	DeadCode DeadCodeConfig `mapstructure:"dead_code" yaml:"dead_code"`

	// Clones holds the unified clone detection configuration
	Clones *PyscnConfig `mapstructure:"clones" yaml:"clones"`

	// SystemAnalysis holds system-level analysis configuration
	SystemAnalysis SystemAnalysisConfig `mapstructure:"system_analysis" yaml:"system_analysis"`

	// Dependencies holds dependency analysis configuration
	Dependencies DependencyAnalysisConfig `mapstructure:"dependencies" yaml:"dependencies"`

	// Architecture holds architecture validation configuration
	Architecture ArchitectureConfig `mapstructure:"architecture" yaml:"architecture"`

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

	// Directory specifies the output directory for reports (empty = tool default, e.g., ".pyscn/reports" under current working directory)
	Directory string `mapstructure:"directory" yaml:"directory"`
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
		// Use unified pyscn configuration
		Clones: DefaultPyscnConfig(),

		// System analysis configuration
		SystemAnalysis: SystemAnalysisConfig{
			Enabled:               false, // Disabled by default - opt-in feature
			EnableDependencies:    true,
			EnableArchitecture:    true,
			UseComplexityData:     true,
			UseClonesData:         true,
			UseDeadCodeData:       true,
			GenerateUnifiedReport: true,
		},

		// Dependency analysis configuration
		Dependencies: DependencyAnalysisConfig{
			Enabled:           false, // Disabled by default - opt-in feature
			IncludeStdLib:     false,
			IncludeThirdParty: true,
			FollowRelative:    true,
			DetectCycles:      true,
			CalculateMetrics:  true,
			FindLongChains:    true,
			MinCoupling:       0,
			MaxCoupling:       0, // No limit
			MinInstability:    0.0,
			MaxDistance:       1.0,
			SortBy:            "name",
			ShowMatrix:        false,
			ShowMetrics:       false,
			ShowChains:        false,
			GenerateDotGraph:  false,
			CycleReporting:    "summary", // all, critical, summary
			MaxCyclesToShow:   10,
			ShowCyclePaths:    false,
		},

		// Architecture validation configuration
		Architecture: ArchitectureConfig{
			Enabled:                         false, // Disabled by default - opt-in feature
			ValidateLayers:                  true,
			ValidateCohesion:                true,
			ValidateResponsibility:          true,
			Layers:                          []LayerDefinition{}, // Empty by default
			Rules:                           []LayerRule{},       // Empty by default
			MinCohesion:                     0.5,
			MaxCoupling:                     10,
			MaxResponsibilities:             3,
			LayerViolationSeverity:          "error",
			CohesionViolationSeverity:       "warning",
			ResponsibilityViolationSeverity: "warning",
			ShowAllViolations:               false,
			GroupByType:                     true,
			IncludeSuggestions:              true,
			MaxViolationsToShow:             20,
			CustomPatterns:                  []string{},
			AllowedPatterns:                 []string{},
			ForbiddenPatterns:               []string{},
			StrictMode:                      true,
			FailOnViolations:                false,
		},

		Output: OutputConfig{
			Format:        "text",
			ShowDetails:   false,
			SortBy:        "complexity",
			MinComplexity: DefaultMinComplexityFilter,
		},
		Analysis: AnalysisConfig{
			IncludePatterns: []string{"**/*.py"},
			ExcludePatterns: []string{"test_*.py", "*_test.py"},
			Recursive:       true,
			FollowSymlinks:  false,
		},
	}

	return config
}

// LoadConfig loads configuration from file or returns default config
// This is a TOML-only implementation that replaces the previous Viper-based loader
func LoadConfig(configPath string) (*Config, error) {
	return LoadConfigWithTarget(configPath, "")
}

// LoadConfigWithTarget loads configuration with target path context
// Uses TOML-only configuration loader
func LoadConfigWithTarget(configPath string, targetPath string) (*Config, error) {
	loader := NewTomlConfigLoader()

	// Resolve once so every analysis phase reads the same config source.
	resolvedConfigPath, err := loader.ResolveConfigPath(configPath, targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve configuration: %w", err)
	}

	var pyscnCfg *PyscnConfig
	if resolvedConfigPath == "" {
		pyscnCfg = DefaultPyscnConfig()
	} else {
		// Load PyscnConfig using TOML loader
		pyscnCfg, err = loader.LoadConfig(resolvedConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load configuration: %w", err)
		}
	}

	// Convert to legacy Config struct
	cfg := PyscnConfigToConfig(pyscnCfg)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// PyscnConfigToConfig converts PyscnConfig (from TOML loader) to legacy Config struct
func PyscnConfigToConfig(pyscn *PyscnConfig) *Config {
	cfg := DefaultConfig()

	// Complexity settings
	if pyscn.ComplexityLowThreshold > 0 {
		cfg.Complexity.LowThreshold = pyscn.ComplexityLowThreshold
	}
	if pyscn.ComplexityMediumThreshold > 0 {
		cfg.Complexity.MediumThreshold = pyscn.ComplexityMediumThreshold
	}
	if pyscn.ComplexityMaxComplexity > 0 {
		cfg.Complexity.MaxComplexity = pyscn.ComplexityMaxComplexity
	}

	// DeadCode settings
	if pyscn.DeadCodeEnabled != nil {
		cfg.DeadCode.Enabled = *pyscn.DeadCodeEnabled
	}
	if pyscn.DeadCodeMinSeverity != "" {
		cfg.DeadCode.MinSeverity = pyscn.DeadCodeMinSeverity
	}
	if pyscn.DeadCodeShowContext != nil {
		cfg.DeadCode.ShowContext = *pyscn.DeadCodeShowContext
	}
	if pyscn.DeadCodeContextLines > 0 {
		cfg.DeadCode.ContextLines = pyscn.DeadCodeContextLines
	}
	if pyscn.DeadCodeSortBy != "" {
		cfg.DeadCode.SortBy = pyscn.DeadCodeSortBy
	}
	if pyscn.DeadCodeDetectAfterReturn != nil {
		cfg.DeadCode.DetectAfterReturn = *pyscn.DeadCodeDetectAfterReturn
	}
	if pyscn.DeadCodeDetectAfterBreak != nil {
		cfg.DeadCode.DetectAfterBreak = *pyscn.DeadCodeDetectAfterBreak
	}
	if pyscn.DeadCodeDetectAfterContinue != nil {
		cfg.DeadCode.DetectAfterContinue = *pyscn.DeadCodeDetectAfterContinue
	}
	if pyscn.DeadCodeDetectAfterRaise != nil {
		cfg.DeadCode.DetectAfterRaise = *pyscn.DeadCodeDetectAfterRaise
	}
	if pyscn.DeadCodeDetectUnreachableBranches != nil {
		cfg.DeadCode.DetectUnreachableBranches = *pyscn.DeadCodeDetectUnreachableBranches
	}
	if len(pyscn.DeadCodeIgnorePatterns) > 0 {
		cfg.DeadCode.IgnorePatterns = pyscn.DeadCodeIgnorePatterns
	}

	// Output settings
	if pyscn.OutputFormat != "" {
		cfg.Output.Format = pyscn.OutputFormat
	}
	if pyscn.OutputShowDetails != nil {
		cfg.Output.ShowDetails = *pyscn.OutputShowDetails
	}
	if pyscn.OutputSortBy != "" {
		cfg.Output.SortBy = pyscn.OutputSortBy
	}
	if pyscn.OutputMinComplexity > 0 {
		cfg.Output.MinComplexity = pyscn.OutputMinComplexity
	}
	if pyscn.OutputDirectory != "" {
		cfg.Output.Directory = pyscn.OutputDirectory
	}

	// Analysis settings
	if len(pyscn.AnalysisIncludePatterns) > 0 {
		cfg.Analysis.IncludePatterns = pyscn.AnalysisIncludePatterns
	}
	if len(pyscn.AnalysisExcludePatterns) > 0 {
		cfg.Analysis.ExcludePatterns = pyscn.AnalysisExcludePatterns
	}
	if pyscn.AnalysisRecursive != nil {
		cfg.Analysis.Recursive = *pyscn.AnalysisRecursive
	}
	if pyscn.AnalysisFollowSymlinks != nil {
		cfg.Analysis.FollowSymlinks = *pyscn.AnalysisFollowSymlinks
	}

	// Clone settings - assign PyscnConfig directly as Clones
	cfg.Clones = pyscn

	// SystemAnalysis settings
	if pyscn.SystemAnalysisEnabled != nil {
		cfg.SystemAnalysis.Enabled = *pyscn.SystemAnalysisEnabled
	}
	if pyscn.SystemAnalysisEnableDependencies != nil {
		cfg.SystemAnalysis.EnableDependencies = *pyscn.SystemAnalysisEnableDependencies
	}
	if pyscn.SystemAnalysisEnableArchitecture != nil {
		cfg.SystemAnalysis.EnableArchitecture = *pyscn.SystemAnalysisEnableArchitecture
	}
	if pyscn.SystemAnalysisUseComplexityData != nil {
		cfg.SystemAnalysis.UseComplexityData = *pyscn.SystemAnalysisUseComplexityData
	}
	if pyscn.SystemAnalysisUseClonesData != nil {
		cfg.SystemAnalysis.UseClonesData = *pyscn.SystemAnalysisUseClonesData
	}
	if pyscn.SystemAnalysisUseDeadCodeData != nil {
		cfg.SystemAnalysis.UseDeadCodeData = *pyscn.SystemAnalysisUseDeadCodeData
	}
	if pyscn.SystemAnalysisGenerateUnifiedReport != nil {
		cfg.SystemAnalysis.GenerateUnifiedReport = *pyscn.SystemAnalysisGenerateUnifiedReport
	}

	// Dependencies settings
	if pyscn.DependenciesEnabled != nil {
		cfg.Dependencies.Enabled = *pyscn.DependenciesEnabled
	}
	if pyscn.DependenciesIncludeStdLib != nil {
		cfg.Dependencies.IncludeStdLib = *pyscn.DependenciesIncludeStdLib
	}
	if pyscn.DependenciesIncludeThirdParty != nil {
		cfg.Dependencies.IncludeThirdParty = *pyscn.DependenciesIncludeThirdParty
	}
	if pyscn.DependenciesFollowRelative != nil {
		cfg.Dependencies.FollowRelative = *pyscn.DependenciesFollowRelative
	}
	if pyscn.DependenciesDetectCycles != nil {
		cfg.Dependencies.DetectCycles = *pyscn.DependenciesDetectCycles
	}
	if pyscn.DependenciesCalculateMetrics != nil {
		cfg.Dependencies.CalculateMetrics = *pyscn.DependenciesCalculateMetrics
	}
	if pyscn.DependenciesFindLongChains != nil {
		cfg.Dependencies.FindLongChains = *pyscn.DependenciesFindLongChains
	}
	if pyscn.DependenciesMinCoupling > 0 {
		cfg.Dependencies.MinCoupling = pyscn.DependenciesMinCoupling
	}
	if pyscn.DependenciesMaxCoupling > 0 {
		cfg.Dependencies.MaxCoupling = pyscn.DependenciesMaxCoupling
	}
	if pyscn.DependenciesMinInstability > 0 {
		cfg.Dependencies.MinInstability = pyscn.DependenciesMinInstability
	}
	if pyscn.DependenciesMaxDistance > 0 {
		cfg.Dependencies.MaxDistance = pyscn.DependenciesMaxDistance
	}
	if pyscn.DependenciesSortBy != "" {
		cfg.Dependencies.SortBy = pyscn.DependenciesSortBy
	}
	if pyscn.DependenciesShowMatrix != nil {
		cfg.Dependencies.ShowMatrix = *pyscn.DependenciesShowMatrix
	}
	if pyscn.DependenciesShowMetrics != nil {
		cfg.Dependencies.ShowMetrics = *pyscn.DependenciesShowMetrics
	}
	if pyscn.DependenciesShowChains != nil {
		cfg.Dependencies.ShowChains = *pyscn.DependenciesShowChains
	}
	if pyscn.DependenciesGenerateDotGraph != nil {
		cfg.Dependencies.GenerateDotGraph = *pyscn.DependenciesGenerateDotGraph
	}
	if pyscn.DependenciesCycleReporting != "" {
		cfg.Dependencies.CycleReporting = pyscn.DependenciesCycleReporting
	}
	if pyscn.DependenciesMaxCyclesToShow > 0 {
		cfg.Dependencies.MaxCyclesToShow = pyscn.DependenciesMaxCyclesToShow
	}
	if pyscn.DependenciesShowCyclePaths != nil {
		cfg.Dependencies.ShowCyclePaths = *pyscn.DependenciesShowCyclePaths
	}

	// Architecture settings
	if pyscn.ArchitectureEnabled != nil {
		cfg.Architecture.Enabled = *pyscn.ArchitectureEnabled
	}
	if pyscn.ArchitectureValidateLayers != nil {
		cfg.Architecture.ValidateLayers = *pyscn.ArchitectureValidateLayers
	}
	if pyscn.ArchitectureValidateCohesion != nil {
		cfg.Architecture.ValidateCohesion = *pyscn.ArchitectureValidateCohesion
	}
	if pyscn.ArchitectureValidateResponsibility != nil {
		cfg.Architecture.ValidateResponsibility = *pyscn.ArchitectureValidateResponsibility
	}
	if pyscn.ArchitectureMinCohesion > 0 {
		cfg.Architecture.MinCohesion = pyscn.ArchitectureMinCohesion
	}
	if pyscn.ArchitectureMaxCoupling > 0 {
		cfg.Architecture.MaxCoupling = pyscn.ArchitectureMaxCoupling
	}
	if pyscn.ArchitectureMaxResponsibilities > 0 {
		cfg.Architecture.MaxResponsibilities = pyscn.ArchitectureMaxResponsibilities
	}
	if pyscn.ArchitectureLayerViolationSeverity != "" {
		cfg.Architecture.LayerViolationSeverity = pyscn.ArchitectureLayerViolationSeverity
	}
	if pyscn.ArchitectureCohesionViolationSeverity != "" {
		cfg.Architecture.CohesionViolationSeverity = pyscn.ArchitectureCohesionViolationSeverity
	}
	if pyscn.ArchitectureResponsibilityViolationSeverity != "" {
		cfg.Architecture.ResponsibilityViolationSeverity = pyscn.ArchitectureResponsibilityViolationSeverity
	}
	if pyscn.ArchitectureShowAllViolations != nil {
		cfg.Architecture.ShowAllViolations = *pyscn.ArchitectureShowAllViolations
	}
	if pyscn.ArchitectureGroupByType != nil {
		cfg.Architecture.GroupByType = *pyscn.ArchitectureGroupByType
	}
	if pyscn.ArchitectureIncludeSuggestions != nil {
		cfg.Architecture.IncludeSuggestions = *pyscn.ArchitectureIncludeSuggestions
	}
	if pyscn.ArchitectureMaxViolationsToShow > 0 {
		cfg.Architecture.MaxViolationsToShow = pyscn.ArchitectureMaxViolationsToShow
	}
	if len(pyscn.ArchitectureCustomPatterns) > 0 {
		cfg.Architecture.CustomPatterns = pyscn.ArchitectureCustomPatterns
	}
	if len(pyscn.ArchitectureAllowedPatterns) > 0 {
		cfg.Architecture.AllowedPatterns = pyscn.ArchitectureAllowedPatterns
	}
	if len(pyscn.ArchitectureForbiddenPatterns) > 0 {
		cfg.Architecture.ForbiddenPatterns = pyscn.ArchitectureForbiddenPatterns
	}
	if pyscn.ArchitectureStrictMode != nil {
		cfg.Architecture.StrictMode = *pyscn.ArchitectureStrictMode
	}
	if pyscn.ArchitectureFailOnViolations != nil {
		cfg.Architecture.FailOnViolations = *pyscn.ArchitectureFailOnViolations
	}

	return cfg
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
	if c.Clones != nil {
		if err := c.Clones.Validate(); err != nil {
			return err
		}
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

// SaveConfig saves configuration to a TOML file
func SaveConfig(config *Config, path string) error {
	// Convert Config to PyscnTomlConfig for TOML serialization
	tomlCfg := ConfigToPyscnTomlConfig(config)

	data, err := toml.Marshal(tomlCfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	return os.WriteFile(path, data, 0644)
}

// ConfigToPyscnTomlConfig converts a Config to PyscnTomlConfig for TOML serialization
func ConfigToPyscnTomlConfig(cfg *Config) *PyscnTomlConfig {
	return &PyscnTomlConfig{
		Complexity: ComplexityTomlConfig{
			LowThreshold:    &cfg.Complexity.LowThreshold,
			MediumThreshold: &cfg.Complexity.MediumThreshold,
			MaxComplexity:   &cfg.Complexity.MaxComplexity,
		},
		DeadCode: DeadCodeTomlConfig{
			Enabled:                   &cfg.DeadCode.Enabled,
			MinSeverity:               cfg.DeadCode.MinSeverity,
			ShowContext:               &cfg.DeadCode.ShowContext,
			ContextLines:              &cfg.DeadCode.ContextLines,
			SortBy:                    cfg.DeadCode.SortBy,
			DetectAfterReturn:         &cfg.DeadCode.DetectAfterReturn,
			DetectAfterBreak:          &cfg.DeadCode.DetectAfterBreak,
			DetectAfterContinue:       &cfg.DeadCode.DetectAfterContinue,
			DetectAfterRaise:          &cfg.DeadCode.DetectAfterRaise,
			DetectUnreachableBranches: &cfg.DeadCode.DetectUnreachableBranches,
			IgnorePatterns:            cfg.DeadCode.IgnorePatterns,
		},
		Output: OutputTomlConfig{
			Format:        cfg.Output.Format,
			ShowDetails:   &cfg.Output.ShowDetails,
			SortBy:        cfg.Output.SortBy,
			MinComplexity: &cfg.Output.MinComplexity,
			Directory:     cfg.Output.Directory,
		},
		Analysis: AnalysisTomlConfig{
			IncludePatterns: cfg.Analysis.IncludePatterns,
			ExcludePatterns: cfg.Analysis.ExcludePatterns,
			Recursive:       &cfg.Analysis.Recursive,
			FollowSymlinks:  &cfg.Analysis.FollowSymlinks,
		},
	}
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

// SystemAnalysisConfig holds configuration for system-level analysis
type SystemAnalysisConfig struct {
	// Enabled controls whether system analysis is performed
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`

	// Analysis components to enable
	EnableDependencies bool `mapstructure:"enable_dependencies" yaml:"enable_dependencies"`
	EnableArchitecture bool `mapstructure:"enable_architecture" yaml:"enable_architecture"`

	// Integration with other analyses
	UseComplexityData bool `mapstructure:"use_complexity_data" yaml:"use_complexity_data"`
	UseClonesData     bool `mapstructure:"use_clones_data" yaml:"use_clones_data"`
	UseDeadCodeData   bool `mapstructure:"use_dead_code_data" yaml:"use_dead_code_data"`

	// Output options
	GenerateUnifiedReport bool `mapstructure:"generate_unified_report" yaml:"generate_unified_report"`
}

// DependencyAnalysisConfig holds configuration for dependency analysis
type DependencyAnalysisConfig struct {
	// Enabled controls whether dependency analysis is performed
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`

	// Scope options
	IncludeStdLib     bool `mapstructure:"include_stdlib" yaml:"include_stdlib"`
	IncludeThirdParty bool `mapstructure:"include_third_party" yaml:"include_third_party"`
	FollowRelative    bool `mapstructure:"follow_relative" yaml:"follow_relative"`

	// Analysis options
	DetectCycles     bool `mapstructure:"detect_cycles" yaml:"detect_cycles"`
	CalculateMetrics bool `mapstructure:"calculate_metrics" yaml:"calculate_metrics"`
	FindLongChains   bool `mapstructure:"find_long_chains" yaml:"find_long_chains"`

	// Filtering thresholds
	MinCoupling    int     `mapstructure:"min_coupling" yaml:"min_coupling"`
	MaxCoupling    int     `mapstructure:"max_coupling" yaml:"max_coupling"`
	MinInstability float64 `mapstructure:"min_instability" yaml:"min_instability"`
	MaxDistance    float64 `mapstructure:"max_distance" yaml:"max_distance"`

	// Reporting options
	SortBy           string `mapstructure:"sort_by" yaml:"sort_by"` // name, coupling, instability, distance, risk
	ShowMatrix       bool   `mapstructure:"show_matrix" yaml:"show_matrix"`
	ShowMetrics      bool   `mapstructure:"show_metrics" yaml:"show_metrics"`
	ShowChains       bool   `mapstructure:"show_chains" yaml:"show_chains"`
	GenerateDotGraph bool   `mapstructure:"generate_dot_graph" yaml:"generate_dot_graph"`

	// Cycle analysis
	CycleReporting  string `mapstructure:"cycle_reporting" yaml:"cycle_reporting"` // all, critical, summary
	MaxCyclesToShow int    `mapstructure:"max_cycles_to_show" yaml:"max_cycles_to_show"`
	ShowCyclePaths  bool   `mapstructure:"show_cycle_paths" yaml:"show_cycle_paths"`
}

// ArchitectureConfig holds configuration for architecture validation
type ArchitectureConfig struct {
	// Enabled controls whether architecture validation is performed
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`

	// Validation modes
	ValidateLayers         bool `mapstructure:"validate_layers" yaml:"validate_layers"`
	ValidateCohesion       bool `mapstructure:"validate_cohesion" yaml:"validate_cohesion"`
	ValidateResponsibility bool `mapstructure:"validate_responsibility" yaml:"validate_responsibility"`

	// Layer definitions
	Layers []LayerDefinition `mapstructure:"layers" yaml:"layers"`
	Rules  []LayerRule       `mapstructure:"rules" yaml:"rules"`

	// Thresholds
	MinCohesion         float64 `mapstructure:"min_cohesion" yaml:"min_cohesion"`
	MaxCoupling         int     `mapstructure:"max_coupling" yaml:"max_coupling"`
	MaxResponsibilities int     `mapstructure:"max_responsibilities" yaml:"max_responsibilities"`

	// Violation severity levels
	LayerViolationSeverity          string `mapstructure:"layer_violation_severity" yaml:"layer_violation_severity"`
	CohesionViolationSeverity       string `mapstructure:"cohesion_violation_severity" yaml:"cohesion_violation_severity"`
	ResponsibilityViolationSeverity string `mapstructure:"responsibility_violation_severity" yaml:"responsibility_violation_severity"`

	// Reporting options
	ShowAllViolations   bool `mapstructure:"show_all_violations" yaml:"show_all_violations"`
	GroupByType         bool `mapstructure:"group_by_type" yaml:"group_by_type"`
	IncludeSuggestions  bool `mapstructure:"include_suggestions" yaml:"include_suggestions"`
	MaxViolationsToShow int  `mapstructure:"max_violations_to_show" yaml:"max_violations_to_show"`

	// Custom rules
	CustomPatterns    []string `mapstructure:"custom_patterns" yaml:"custom_patterns"`
	AllowedPatterns   []string `mapstructure:"allowed_patterns" yaml:"allowed_patterns"`
	ForbiddenPatterns []string `mapstructure:"forbidden_patterns" yaml:"forbidden_patterns"`

	// Strict mode enforcement
	StrictMode       bool `mapstructure:"strict_mode" yaml:"strict_mode"`
	FailOnViolations bool `mapstructure:"fail_on_violations" yaml:"fail_on_violations"`
}

// LayerDefinition defines an architectural layer
type LayerDefinition struct {
	Name        string   `mapstructure:"name" yaml:"name"`
	Packages    []string `mapstructure:"packages" yaml:"packages"`
	Description string   `mapstructure:"description" yaml:"description"`
	IsAbstract  bool     `mapstructure:"is_abstract" yaml:"is_abstract"`
}

// LayerRule defines dependency rules between layers
type LayerRule struct {
	From        string   `mapstructure:"from" yaml:"from"`
	Allow       []string `mapstructure:"allow" yaml:"allow"`
	Deny        []string `mapstructure:"deny" yaml:"deny"`
	Description string   `mapstructure:"description" yaml:"description"`
}
