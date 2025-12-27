package config

import (
	"fmt"
	"io"

	"github.com/ludo-technologies/pyscn/domain"
)

// PyscnConfig represents the universal pyscn configuration from TOML files
// This holds all configuration sections that can be loaded from .pyscn.toml or pyproject.toml
type PyscnConfig struct {
	// Clone Analysis Configuration
	Analysis CloneAnalysisConfig `mapstructure:"analysis" yaml:"analysis" json:"analysis"`

	// Thresholds Configuration
	Thresholds ThresholdConfig `mapstructure:"thresholds" yaml:"thresholds" json:"thresholds"`

	// Filtering Configuration
	Filtering FilteringConfig `mapstructure:"filtering" yaml:"filtering" json:"filtering"`

	// Input Configuration
	Input InputConfig `mapstructure:"input" yaml:"input" json:"input"`

	// Output Configuration (Clone-specific)
	Output CloneOutputConfig `mapstructure:"output" yaml:"output" json:"output"`

	// Performance Configuration
	Performance PerformanceConfig `mapstructure:"performance" yaml:"performance" json:"performance"`

	// Grouping Configuration
	Grouping GroupingConfig `mapstructure:"grouping" yaml:"grouping" json:"grouping"`

	// LSH Configuration
	LSH LSHConfig `mapstructure:"lsh" yaml:"lsh" json:"lsh"`

	// Complexity Configuration (from [complexity] section in TOML)
	ComplexityLowThreshold    int `mapstructure:"complexity_low_threshold" yaml:"complexity_low_threshold" json:"complexity_low_threshold"`
	ComplexityMediumThreshold int `mapstructure:"complexity_medium_threshold" yaml:"complexity_medium_threshold" json:"complexity_medium_threshold"`
	ComplexityMaxComplexity   int `mapstructure:"complexity_max_complexity" yaml:"complexity_max_complexity" json:"complexity_max_complexity"`
	ComplexityMinComplexity   int `mapstructure:"complexity_min_complexity" yaml:"complexity_min_complexity" json:"complexity_min_complexity"`

	// DeadCode Configuration (from [dead_code] section in TOML)
	DeadCodeEnabled                   *bool    `mapstructure:"dead_code_enabled" yaml:"dead_code_enabled" json:"dead_code_enabled"`
	DeadCodeMinSeverity               string   `mapstructure:"dead_code_min_severity" yaml:"dead_code_min_severity" json:"dead_code_min_severity"`
	DeadCodeShowContext               *bool    `mapstructure:"dead_code_show_context" yaml:"dead_code_show_context" json:"dead_code_show_context"`
	DeadCodeContextLines              int      `mapstructure:"dead_code_context_lines" yaml:"dead_code_context_lines" json:"dead_code_context_lines"`
	DeadCodeSortBy                    string   `mapstructure:"dead_code_sort_by" yaml:"dead_code_sort_by" json:"dead_code_sort_by"`
	DeadCodeDetectAfterReturn         *bool    `mapstructure:"dead_code_detect_after_return" yaml:"dead_code_detect_after_return" json:"dead_code_detect_after_return"`
	DeadCodeDetectAfterBreak          *bool    `mapstructure:"dead_code_detect_after_break" yaml:"dead_code_detect_after_break" json:"dead_code_detect_after_break"`
	DeadCodeDetectAfterContinue       *bool    `mapstructure:"dead_code_detect_after_continue" yaml:"dead_code_detect_after_continue" json:"dead_code_detect_after_continue"`
	DeadCodeDetectAfterRaise          *bool    `mapstructure:"dead_code_detect_after_raise" yaml:"dead_code_detect_after_raise" json:"dead_code_detect_after_raise"`
	DeadCodeDetectUnreachableBranches *bool    `mapstructure:"dead_code_detect_unreachable_branches" yaml:"dead_code_detect_unreachable_branches" json:"dead_code_detect_unreachable_branches"`
	DeadCodeIgnorePatterns            []string `mapstructure:"dead_code_ignore_patterns" yaml:"dead_code_ignore_patterns" json:"dead_code_ignore_patterns"`

	// Output Configuration (from [output] section in TOML - general output settings)
	OutputFormat        string `mapstructure:"output_format" yaml:"output_format" json:"output_format"`
	OutputShowDetails   *bool  `mapstructure:"output_show_details" yaml:"output_show_details" json:"output_show_details"`
	OutputSortBy        string `mapstructure:"output_sort_by" yaml:"output_sort_by" json:"output_sort_by"`
	OutputMinComplexity int    `mapstructure:"output_min_complexity" yaml:"output_min_complexity" json:"output_min_complexity"`
	OutputDirectory     string `mapstructure:"output_directory" yaml:"output_directory" json:"output_directory"`

	// Analysis Configuration (from [analysis] section in TOML - general analysis settings)
	AnalysisIncludePatterns []string `mapstructure:"analysis_include_patterns" yaml:"analysis_include_patterns" json:"analysis_include_patterns"`
	AnalysisExcludePatterns []string `mapstructure:"analysis_exclude_patterns" yaml:"analysis_exclude_patterns" json:"analysis_exclude_patterns"`
	AnalysisRecursive       *bool    `mapstructure:"analysis_recursive" yaml:"analysis_recursive" json:"analysis_recursive"`
	AnalysisFollowSymlinks  *bool    `mapstructure:"analysis_follow_symlinks" yaml:"analysis_follow_symlinks" json:"analysis_follow_symlinks"`

	// CBO Configuration (from [cbo] section in TOML)
	CboLowThreshold    int   `mapstructure:"cbo_low_threshold" yaml:"cbo_low_threshold" json:"cbo_low_threshold"`
	CboMediumThreshold int   `mapstructure:"cbo_medium_threshold" yaml:"cbo_medium_threshold" json:"cbo_medium_threshold"`
	CboMinCbo          int   `mapstructure:"cbo_min_cbo" yaml:"cbo_min_cbo" json:"cbo_min_cbo"`
	CboMaxCbo          int   `mapstructure:"cbo_max_cbo" yaml:"cbo_max_cbo" json:"cbo_max_cbo"`
	CboShowZeros       *bool `mapstructure:"cbo_show_zeros" yaml:"cbo_show_zeros" json:"cbo_show_zeros"`
	CboIncludeBuiltins *bool `mapstructure:"cbo_include_builtins" yaml:"cbo_include_builtins" json:"cbo_include_builtins"`
	CboIncludeImports  *bool `mapstructure:"cbo_include_imports" yaml:"cbo_include_imports" json:"cbo_include_imports"`

	// Architecture Configuration (from [architecture] section in TOML)
	ArchitectureEnabled                         *bool    `mapstructure:"architecture_enabled" yaml:"architecture_enabled" json:"architecture_enabled"`
	ArchitectureValidateLayers                  *bool    `mapstructure:"architecture_validate_layers" yaml:"architecture_validate_layers" json:"architecture_validate_layers"`
	ArchitectureValidateCohesion                *bool    `mapstructure:"architecture_validate_cohesion" yaml:"architecture_validate_cohesion" json:"architecture_validate_cohesion"`
	ArchitectureValidateResponsibility          *bool    `mapstructure:"architecture_validate_responsibility" yaml:"architecture_validate_responsibility" json:"architecture_validate_responsibility"`
	ArchitectureMinCohesion                     float64  `mapstructure:"architecture_min_cohesion" yaml:"architecture_min_cohesion" json:"architecture_min_cohesion"`
	ArchitectureMaxCoupling                     int      `mapstructure:"architecture_max_coupling" yaml:"architecture_max_coupling" json:"architecture_max_coupling"`
	ArchitectureMaxResponsibilities             int      `mapstructure:"architecture_max_responsibilities" yaml:"architecture_max_responsibilities" json:"architecture_max_responsibilities"`
	ArchitectureLayerViolationSeverity          string   `mapstructure:"architecture_layer_violation_severity" yaml:"architecture_layer_violation_severity" json:"architecture_layer_violation_severity"`
	ArchitectureCohesionViolationSeverity       string   `mapstructure:"architecture_cohesion_violation_severity" yaml:"architecture_cohesion_violation_severity" json:"architecture_cohesion_violation_severity"`
	ArchitectureResponsibilityViolationSeverity string   `mapstructure:"architecture_responsibility_violation_severity" yaml:"architecture_responsibility_violation_severity" json:"architecture_responsibility_violation_severity"`
	ArchitectureShowAllViolations               *bool    `mapstructure:"architecture_show_all_violations" yaml:"architecture_show_all_violations" json:"architecture_show_all_violations"`
	ArchitectureGroupByType                     *bool    `mapstructure:"architecture_group_by_type" yaml:"architecture_group_by_type" json:"architecture_group_by_type"`
	ArchitectureIncludeSuggestions              *bool    `mapstructure:"architecture_include_suggestions" yaml:"architecture_include_suggestions" json:"architecture_include_suggestions"`
	ArchitectureMaxViolationsToShow             int      `mapstructure:"architecture_max_violations_to_show" yaml:"architecture_max_violations_to_show" json:"architecture_max_violations_to_show"`
	ArchitectureCustomPatterns                  []string `mapstructure:"architecture_custom_patterns" yaml:"architecture_custom_patterns" json:"architecture_custom_patterns"`
	ArchitectureAllowedPatterns                 []string `mapstructure:"architecture_allowed_patterns" yaml:"architecture_allowed_patterns" json:"architecture_allowed_patterns"`
	ArchitectureForbiddenPatterns               []string `mapstructure:"architecture_forbidden_patterns" yaml:"architecture_forbidden_patterns" json:"architecture_forbidden_patterns"`
	ArchitectureStrictMode                      *bool    `mapstructure:"architecture_strict_mode" yaml:"architecture_strict_mode" json:"architecture_strict_mode"`
	ArchitectureFailOnViolations                *bool    `mapstructure:"architecture_fail_on_violations" yaml:"architecture_fail_on_violations" json:"architecture_fail_on_violations"`

	// SystemAnalysis Configuration (from [system_analysis] section in TOML)
	SystemAnalysisEnabled               *bool `mapstructure:"system_analysis_enabled" yaml:"system_analysis_enabled" json:"system_analysis_enabled"`
	SystemAnalysisEnableDependencies    *bool `mapstructure:"system_analysis_enable_dependencies" yaml:"system_analysis_enable_dependencies" json:"system_analysis_enable_dependencies"`
	SystemAnalysisEnableArchitecture    *bool `mapstructure:"system_analysis_enable_architecture" yaml:"system_analysis_enable_architecture" json:"system_analysis_enable_architecture"`
	SystemAnalysisUseComplexityData     *bool `mapstructure:"system_analysis_use_complexity_data" yaml:"system_analysis_use_complexity_data" json:"system_analysis_use_complexity_data"`
	SystemAnalysisUseClonesData         *bool `mapstructure:"system_analysis_use_clones_data" yaml:"system_analysis_use_clones_data" json:"system_analysis_use_clones_data"`
	SystemAnalysisUseDeadCodeData       *bool `mapstructure:"system_analysis_use_dead_code_data" yaml:"system_analysis_use_dead_code_data" json:"system_analysis_use_dead_code_data"`
	SystemAnalysisGenerateUnifiedReport *bool `mapstructure:"system_analysis_generate_unified_report" yaml:"system_analysis_generate_unified_report" json:"system_analysis_generate_unified_report"`

	// Dependencies Configuration (from [dependencies] section in TOML)
	DependenciesEnabled           *bool   `mapstructure:"dependencies_enabled" yaml:"dependencies_enabled" json:"dependencies_enabled"`
	DependenciesIncludeStdLib     *bool   `mapstructure:"dependencies_include_stdlib" yaml:"dependencies_include_stdlib" json:"dependencies_include_stdlib"`
	DependenciesIncludeThirdParty *bool   `mapstructure:"dependencies_include_third_party" yaml:"dependencies_include_third_party" json:"dependencies_include_third_party"`
	DependenciesFollowRelative    *bool   `mapstructure:"dependencies_follow_relative" yaml:"dependencies_follow_relative" json:"dependencies_follow_relative"`
	DependenciesDetectCycles      *bool   `mapstructure:"dependencies_detect_cycles" yaml:"dependencies_detect_cycles" json:"dependencies_detect_cycles"`
	DependenciesCalculateMetrics  *bool   `mapstructure:"dependencies_calculate_metrics" yaml:"dependencies_calculate_metrics" json:"dependencies_calculate_metrics"`
	DependenciesFindLongChains    *bool   `mapstructure:"dependencies_find_long_chains" yaml:"dependencies_find_long_chains" json:"dependencies_find_long_chains"`
	DependenciesMinCoupling       int     `mapstructure:"dependencies_min_coupling" yaml:"dependencies_min_coupling" json:"dependencies_min_coupling"`
	DependenciesMaxCoupling       int     `mapstructure:"dependencies_max_coupling" yaml:"dependencies_max_coupling" json:"dependencies_max_coupling"`
	DependenciesMinInstability    float64 `mapstructure:"dependencies_min_instability" yaml:"dependencies_min_instability" json:"dependencies_min_instability"`
	DependenciesMaxDistance       float64 `mapstructure:"dependencies_max_distance" yaml:"dependencies_max_distance" json:"dependencies_max_distance"`
	DependenciesSortBy            string  `mapstructure:"dependencies_sort_by" yaml:"dependencies_sort_by" json:"dependencies_sort_by"`
	DependenciesShowMatrix        *bool   `mapstructure:"dependencies_show_matrix" yaml:"dependencies_show_matrix" json:"dependencies_show_matrix"`
	DependenciesShowMetrics       *bool   `mapstructure:"dependencies_show_metrics" yaml:"dependencies_show_metrics" json:"dependencies_show_metrics"`
	DependenciesShowChains        *bool   `mapstructure:"dependencies_show_chains" yaml:"dependencies_show_chains" json:"dependencies_show_chains"`
	DependenciesGenerateDotGraph  *bool   `mapstructure:"dependencies_generate_dot_graph" yaml:"dependencies_generate_dot_graph" json:"dependencies_generate_dot_graph"`
	DependenciesCycleReporting    string  `mapstructure:"dependencies_cycle_reporting" yaml:"dependencies_cycle_reporting" json:"dependencies_cycle_reporting"`
	DependenciesMaxCyclesToShow   int     `mapstructure:"dependencies_max_cycles_to_show" yaml:"dependencies_max_cycles_to_show" json:"dependencies_max_cycles_to_show"`
	DependenciesShowCyclePaths    *bool   `mapstructure:"dependencies_show_cycle_paths" yaml:"dependencies_show_cycle_paths" json:"dependencies_show_cycle_paths"`
}

// CloneAnalysisConfig holds core analysis parameters
type CloneAnalysisConfig struct {
	// Minimum requirements for clone candidates
	MinLines int `mapstructure:"min_lines" yaml:"min_lines" json:"min_lines"`
	MinNodes int `mapstructure:"min_nodes" yaml:"min_nodes" json:"min_nodes"`

	// Edit distance configuration
	MaxEditDistance float64 `mapstructure:"max_edit_distance" yaml:"max_edit_distance" json:"max_edit_distance"`

	// Normalization options
	IgnoreLiterals    *bool `mapstructure:"ignore_literals" yaml:"ignore_literals" json:"ignore_literals"`
	IgnoreIdentifiers *bool `mapstructure:"ignore_identifiers" yaml:"ignore_identifiers" json:"ignore_identifiers"`
	SkipDocstrings    *bool `mapstructure:"skip_docstrings" yaml:"skip_docstrings" json:"skip_docstrings"`

	// Cost model configuration
	CostModelType string `mapstructure:"cost_model_type" yaml:"cost_model_type" json:"cost_model_type"`

	// Advanced analysis
	EnableDFA *bool `mapstructure:"enable_dfa" yaml:"enable_dfa" json:"enable_dfa"` // Data Flow Analysis for Type-4
}

// ThresholdConfig holds similarity thresholds for different clone types
type ThresholdConfig struct {
	// Type-specific thresholds (these determine clone classification)
	Type1Threshold float64 `mapstructure:"type1_threshold" yaml:"type1_threshold" json:"type1_threshold"`
	Type2Threshold float64 `mapstructure:"type2_threshold" yaml:"type2_threshold" json:"type2_threshold"`
	Type3Threshold float64 `mapstructure:"type3_threshold" yaml:"type3_threshold" json:"type3_threshold"`
	Type4Threshold float64 `mapstructure:"type4_threshold" yaml:"type4_threshold" json:"type4_threshold"`

	// General similarity threshold (minimum for any clone to be reported)
	SimilarityThreshold float64 `mapstructure:"similarity_threshold" yaml:"similarity_threshold" json:"similarity_threshold"`
}

// FilteringConfig holds filtering and selection criteria
type FilteringConfig struct {
	// Similarity range filtering
	MinSimilarity float64 `mapstructure:"min_similarity" yaml:"min_similarity" json:"min_similarity"`
	MaxSimilarity float64 `mapstructure:"max_similarity" yaml:"max_similarity" json:"max_similarity"`

	// Clone type filtering
	EnabledCloneTypes []string `mapstructure:"enabled_clone_types" yaml:"enabled_clone_types" json:"enabled_clone_types"`

	// Result limiting
	MaxResults int `mapstructure:"max_results" yaml:"max_results" json:"max_results"`
}

// InputConfig holds input processing configuration
type InputConfig struct {
	// File selection
	Paths           []string `mapstructure:"paths" yaml:"paths" json:"paths"`
	Recursive       *bool    `mapstructure:"recursive" yaml:"recursive" json:"recursive"`
	IncludePatterns []string `mapstructure:"include_patterns" yaml:"include_patterns" json:"include_patterns"`
	ExcludePatterns []string `mapstructure:"exclude_patterns" yaml:"exclude_patterns" json:"exclude_patterns"`
}

// CloneOutputConfig holds output formatting configuration
// (This extends the existing OutputConfig with clone-specific fields)
type CloneOutputConfig struct {
	// Format and display
	Format      string `mapstructure:"format" yaml:"format" json:"format"`
	ShowDetails *bool  `mapstructure:"show_details" yaml:"show_details" json:"show_details"`
	ShowContent *bool  `mapstructure:"show_content" yaml:"show_content" json:"show_content"`

	// Sorting and grouping
	SortBy      string `mapstructure:"sort_by" yaml:"sort_by" json:"sort_by"`
	GroupClones *bool  `mapstructure:"group_clones" yaml:"group_clones" json:"group_clones"`

	// Output destination (not serialized)
	Writer io.Writer `json:"-" yaml:"-" mapstructure:"-"`
}

// PerformanceConfig holds performance-related settings
type PerformanceConfig struct {
	// Memory management
	MaxMemoryMB    int   `mapstructure:"max_memory_mb" yaml:"max_memory_mb" json:"max_memory_mb"`
	BatchSize      int   `mapstructure:"batch_size" yaml:"batch_size" json:"batch_size"`
	EnableBatching *bool `mapstructure:"enable_batching" yaml:"enable_batching" json:"enable_batching"`

	// Parallelization
	MaxGoroutines int `mapstructure:"max_goroutines" yaml:"max_goroutines" json:"max_goroutines"`

	// Early termination
	TimeoutSeconds int `mapstructure:"timeout_seconds" yaml:"timeout_seconds" json:"timeout_seconds"`
}

// GroupingConfig holds clone grouping configuration
type GroupingConfig struct {
	// Grouping strategy: connected, star, complete_linkage, k_core
	Mode string `mapstructure:"mode" yaml:"mode" json:"mode"`

	// Minimum similarity threshold for group membership
	Threshold float64 `mapstructure:"threshold" yaml:"threshold" json:"threshold"`

	// K value for k-core mode (minimum neighbors)
	KCoreK int `mapstructure:"k_core_k" yaml:"k_core_k" json:"k_core_k"`
}

// LSHConfig holds LSH acceleration configuration
type LSHConfig struct {
	// Whether to enable LSH acceleration: true, false, "auto"
	Enabled string `mapstructure:"enabled" yaml:"enabled" json:"enabled"`

	// Fragment count threshold for auto-enabling LSH
	AutoThreshold int `mapstructure:"auto_threshold" yaml:"auto_threshold" json:"auto_threshold"`

	// LSH similarity threshold for candidate generation
	SimilarityThreshold float64 `mapstructure:"similarity_threshold" yaml:"similarity_threshold" json:"similarity_threshold"`

	// LSH parameters (advanced)
	Bands  int `mapstructure:"bands" yaml:"bands" json:"bands"`
	Rows   int `mapstructure:"rows" yaml:"rows" json:"rows"`
	Hashes int `mapstructure:"hashes" yaml:"hashes" json:"hashes"`
}

// BoolPtr returns a pointer to the given bool value
// This helper function is used to create *bool values in struct literals
func BoolPtr(b bool) *bool {
	return &b
}

// BoolValue safely dereferences a boolean pointer, returning defaultVal if nil
// This allows safe access to pointer booleans with explicit defaults
func BoolValue(b *bool, defaultVal bool) bool {
	if b == nil {
		return defaultVal
	}
	return *b
}

// DefaultPyscnConfig returns a configuration with sensible defaults
// All default values are sourced from domain/defaults.go
func DefaultPyscnConfig() *PyscnConfig {
	return &PyscnConfig{
		// Clone detection configuration
		Analysis: CloneAnalysisConfig{
			MinLines:          domain.DefaultCloneMinLines,
			MinNodes:          domain.DefaultCloneMinNodes,
			MaxEditDistance:   domain.DefaultCloneMaxEditDistance,
			IgnoreLiterals:    BoolPtr(false),
			IgnoreIdentifiers: BoolPtr(false),
			SkipDocstrings:    BoolPtr(true),
			CostModelType:     "python",
			EnableDFA:         BoolPtr(true), // Enable Data Flow Analysis by default for multi-dimensional classification
		},
		Thresholds: ThresholdConfig{
			Type1Threshold:      domain.DefaultType1CloneThreshold,
			Type2Threshold:      domain.DefaultType2CloneThreshold,
			Type3Threshold:      domain.DefaultType3CloneThreshold,
			Type4Threshold:      domain.DefaultType4CloneThreshold,
			SimilarityThreshold: domain.DefaultCloneSimilarityThreshold,
		},
		Filtering: FilteringConfig{
			MinSimilarity:     0.0,
			MaxSimilarity:     1.0,
			EnabledCloneTypes: domain.DefaultEnabledCloneTypeStrings,
			MaxResults:        10000,
		},
		Input: InputConfig{
			Paths:           []string{"."},
			Recursive:       BoolPtr(true),
			IncludePatterns: []string{"**/*.py"},
			ExcludePatterns: []string{"test_*.py", "*_test.py"},
		},
		Output: CloneOutputConfig{
			Format:      "text",
			ShowDetails: BoolPtr(false),
			ShowContent: BoolPtr(false),
			SortBy:      "similarity",
			GroupClones: BoolPtr(true),
		},
		Performance: PerformanceConfig{
			MaxMemoryMB:    domain.DefaultMaxMemoryMB,
			BatchSize:      domain.DefaultBatchSize,
			EnableBatching: BoolPtr(true),
			MaxGoroutines:  domain.DefaultMaxGoroutines,
			TimeoutSeconds: domain.DefaultTimeoutSeconds,
		},
		Grouping: GroupingConfig{
			Mode:      "connected", // Conservative default
			Threshold: domain.DefaultCloneGroupingThreshold,
			KCoreK:    2,
		},
		LSH: LSHConfig{
			Enabled:             "auto", // Auto-enable based on project size
			AutoThreshold:       domain.DefaultLSHAutoThreshold,
			SimilarityThreshold: domain.DefaultLSHSimilarityThreshold,
			Bands:               domain.DefaultLSHBands,
			Rows:                domain.DefaultLSHRows,
			Hashes:              domain.DefaultLSHHashes,
		},

		// Complexity defaults (from [complexity] section)
		ComplexityLowThreshold:    DefaultLowComplexityThreshold,
		ComplexityMediumThreshold: DefaultMediumComplexityThreshold,
		ComplexityMaxComplexity:   DefaultMaxComplexityLimit,
		ComplexityMinComplexity:   DefaultMinComplexityFilter,

		// DeadCode defaults (from [dead_code] section)
		DeadCodeEnabled:                   BoolPtr(true),
		DeadCodeMinSeverity:               DefaultDeadCodeMinSeverity,
		DeadCodeShowContext:               BoolPtr(false),
		DeadCodeContextLines:              DefaultDeadCodeContextLines,
		DeadCodeSortBy:                    DefaultDeadCodeSortBy,
		DeadCodeDetectAfterReturn:         BoolPtr(true),
		DeadCodeDetectAfterBreak:          BoolPtr(true),
		DeadCodeDetectAfterContinue:       BoolPtr(true),
		DeadCodeDetectAfterRaise:          BoolPtr(true),
		DeadCodeDetectUnreachableBranches: BoolPtr(true),
		DeadCodeIgnorePatterns:            []string{},

		// Output defaults (from [output] section - general output settings)
		OutputFormat:        "text",
		OutputShowDetails:   BoolPtr(false),
		OutputSortBy:        "complexity",
		OutputMinComplexity: DefaultMinComplexityFilter,
		OutputDirectory:     "", // empty = tool default (.pyscn/reports)

		// Analysis defaults (from [analysis] section - general analysis settings)
		AnalysisIncludePatterns: []string{"**/*.py"},
		AnalysisExcludePatterns: []string{"test_*.py", "*_test.py"},
		AnalysisRecursive:       BoolPtr(true),
		AnalysisFollowSymlinks:  BoolPtr(false),

		// CBO defaults (from [cbo] section)
		CboLowThreshold:    domain.DefaultCBOLowThreshold,
		CboMediumThreshold: domain.DefaultCBOMediumThreshold,
		CboMinCbo:          0,
		CboMaxCbo:          0, // No limit
		CboShowZeros:       BoolPtr(false),
		CboIncludeBuiltins: BoolPtr(false),
		CboIncludeImports:  BoolPtr(true),

		// Architecture defaults (from [architecture] section)
		ArchitectureEnabled:                         BoolPtr(false), // Disabled by default - opt-in
		ArchitectureValidateLayers:                  BoolPtr(true),
		ArchitectureValidateCohesion:                BoolPtr(true),
		ArchitectureValidateResponsibility:          BoolPtr(true),
		ArchitectureMinCohesion:                     0.5,
		ArchitectureMaxCoupling:                     10,
		ArchitectureMaxResponsibilities:             3,
		ArchitectureLayerViolationSeverity:          "error",
		ArchitectureCohesionViolationSeverity:       "warning",
		ArchitectureResponsibilityViolationSeverity: "warning",
		ArchitectureShowAllViolations:               BoolPtr(false),
		ArchitectureGroupByType:                     BoolPtr(true),
		ArchitectureIncludeSuggestions:              BoolPtr(true),
		ArchitectureMaxViolationsToShow:             20,
		ArchitectureCustomPatterns:                  []string{},
		ArchitectureAllowedPatterns:                 []string{},
		ArchitectureForbiddenPatterns:               []string{},
		ArchitectureStrictMode:                      BoolPtr(true),
		ArchitectureFailOnViolations:                BoolPtr(false),

		// SystemAnalysis defaults (from [system_analysis] section)
		SystemAnalysisEnabled:               BoolPtr(false), // Disabled by default - opt-in
		SystemAnalysisEnableDependencies:    BoolPtr(true),
		SystemAnalysisEnableArchitecture:    BoolPtr(true),
		SystemAnalysisUseComplexityData:     BoolPtr(true),
		SystemAnalysisUseClonesData:         BoolPtr(true),
		SystemAnalysisUseDeadCodeData:       BoolPtr(true),
		SystemAnalysisGenerateUnifiedReport: BoolPtr(true),

		// Dependencies defaults (from [dependencies] section)
		DependenciesEnabled:           BoolPtr(false), // Disabled by default - opt-in
		DependenciesIncludeStdLib:     BoolPtr(false),
		DependenciesIncludeThirdParty: BoolPtr(true),
		DependenciesFollowRelative:    BoolPtr(true),
		DependenciesDetectCycles:      BoolPtr(true),
		DependenciesCalculateMetrics:  BoolPtr(true),
		DependenciesFindLongChains:    BoolPtr(true),
		DependenciesMinCoupling:       0,
		DependenciesMaxCoupling:       0, // No limit
		DependenciesMinInstability:    0.0,
		DependenciesMaxDistance:       1.0,
		DependenciesSortBy:            "name",
		DependenciesShowMatrix:        BoolPtr(false),
		DependenciesShowMetrics:       BoolPtr(false),
		DependenciesShowChains:        BoolPtr(false),
		DependenciesGenerateDotGraph:  BoolPtr(false),
		DependenciesCycleReporting:    "summary", // all, critical, summary
		DependenciesMaxCyclesToShow:   10,
		DependenciesShowCyclePaths:    BoolPtr(false),
	}
}

// Validate checks if the configuration is valid
func (c *PyscnConfig) Validate() error {
	// Validate analysis config
	if err := c.Analysis.Validate(); err != nil {
		return fmt.Errorf("analysis config invalid: %w", err)
	}

	// Validate thresholds
	if err := c.Thresholds.Validate(); err != nil {
		return fmt.Errorf("thresholds config invalid: %w", err)
	}

	// Validate filtering config
	if err := c.Filtering.Validate(); err != nil {
		return fmt.Errorf("filtering config invalid: %w", err)
	}

	// Validate input config
	if err := c.Input.Validate(); err != nil {
		return fmt.Errorf("input config invalid: %w", err)
	}

	// Validate output config
	if err := c.Output.Validate(); err != nil {
		return fmt.Errorf("output config invalid: %w", err)
	}

	// Validate performance config
	if err := c.Performance.Validate(); err != nil {
		return fmt.Errorf("performance config invalid: %w", err)
	}

	return nil
}

// Validate validates the analysis configuration
func (a *CloneAnalysisConfig) Validate() error {
	if a.MinLines < 1 {
		return fmt.Errorf("min_lines must be >= 1, got %d", a.MinLines)
	}
	if a.MinNodes < 1 {
		return fmt.Errorf("min_nodes must be >= 1, got %d", a.MinNodes)
	}
	if a.MaxEditDistance < 0 {
		return fmt.Errorf("max_edit_distance must be >= 0, got %f", a.MaxEditDistance)
	}

	validCostModels := []string{"default", "python", "weighted"}
	valid := false
	for _, model := range validCostModels {
		if a.CostModelType == model {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("cost_model_type must be one of %v, got %s", validCostModels, a.CostModelType)
	}

	return nil
}

// Validate validates the threshold configuration
func (t *ThresholdConfig) Validate() error {
	// Check range
	thresholds := []float64{t.Type1Threshold, t.Type2Threshold, t.Type3Threshold, t.Type4Threshold}
	for i, threshold := range thresholds {
		if threshold < 0.0 || threshold > 1.0 {
			return fmt.Errorf("threshold %d is out of range [0.0, 1.0]: %f", i+1, threshold)
		}
	}

	// Check ordering: Type1 > Type2 > Type3 > Type4
	if t.Type1Threshold <= t.Type2Threshold {
		return fmt.Errorf("Type1 threshold (%.3f) should be > Type2 threshold (%.3f)", t.Type1Threshold, t.Type2Threshold)
	}
	if t.Type2Threshold <= t.Type3Threshold {
		return fmt.Errorf("Type2 threshold (%.3f) should be > Type3 threshold (%.3f)", t.Type2Threshold, t.Type3Threshold)
	}
	if t.Type3Threshold <= t.Type4Threshold {
		return fmt.Errorf("Type3 threshold (%.3f) should be > Type4 threshold (%.3f)", t.Type3Threshold, t.Type4Threshold)
	}

	// Validate general similarity threshold
	if t.SimilarityThreshold < 0.0 || t.SimilarityThreshold > 1.0 {
		return fmt.Errorf("similarity_threshold must be between 0.0 and 1.0, got %f", t.SimilarityThreshold)
	}

	return nil
}

// Validate validates the filtering configuration
func (f *FilteringConfig) Validate() error {
	if f.MinSimilarity < 0.0 || f.MinSimilarity > 1.0 {
		return fmt.Errorf("min_similarity must be between 0.0 and 1.0, got %f", f.MinSimilarity)
	}
	if f.MaxSimilarity < 0.0 || f.MaxSimilarity > 1.0 {
		return fmt.Errorf("max_similarity must be between 0.0 and 1.0, got %f", f.MaxSimilarity)
	}
	if f.MinSimilarity > f.MaxSimilarity {
		return fmt.Errorf("min_similarity (%f) must be <= max_similarity (%f)", f.MinSimilarity, f.MaxSimilarity)
	}
	if f.MaxResults < 0 {
		return fmt.Errorf("max_results must be >= 0, got %d", f.MaxResults)
	}

	// Validate clone types
	validTypes := []string{"type1", "type2", "type3", "type4"}
	for _, cloneType := range f.EnabledCloneTypes {
		valid := false
		for _, validType := range validTypes {
			if cloneType == validType {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid clone type %s, must be one of %v", cloneType, validTypes)
		}
	}

	return nil
}

// Validate validates the input configuration
func (i *InputConfig) Validate() error {
	if len(i.Paths) == 0 {
		return fmt.Errorf("paths cannot be empty")
	}
	return nil
}

// Validate validates the output configuration
func (o *CloneOutputConfig) Validate() error {
	validFormats := []string{"text", "json", "yaml", "csv"}
	valid := false
	for _, format := range validFormats {
		if o.Format == format {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("format must be one of %v, got %s", validFormats, o.Format)
	}

	validSortBy := []string{"similarity", "size", "location", "type"}
	valid = false
	for _, sort := range validSortBy {
		if o.SortBy == sort {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("sort_by must be one of %v, got %s", validSortBy, o.SortBy)
	}

	return nil
}

// Validate validates the performance configuration
func (p *PerformanceConfig) Validate() error {
	if p.MaxMemoryMB <= 0 {
		return fmt.Errorf("max_memory_mb must be > 0, got %d", p.MaxMemoryMB)
	}
	if p.BatchSize <= 0 {
		return fmt.Errorf("batch_size must be > 0, got %d", p.BatchSize)
	}
	if p.MaxGoroutines <= 0 {
		return fmt.Errorf("max_goroutines must be > 0, got %d", p.MaxGoroutines)
	}
	if p.TimeoutSeconds <= 0 {
		return fmt.Errorf("timeout_seconds must be > 0, got %d", p.TimeoutSeconds)
	}

	return nil
}
