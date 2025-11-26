package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// PyscnTomlConfig represents the structure of .pyscn.toml
type PyscnTomlConfig struct {
	Complexity     ComplexityTomlConfig     `toml:"complexity"`      // [complexity] section
	DeadCode       DeadCodeTomlConfig       `toml:"dead_code"`       // [dead_code] section
	Output         OutputTomlConfig         `toml:"output"`          // [output] section
	Analysis       AnalysisTomlConfig       `toml:"analysis"`        // [analysis] section
	Cbo            CboTomlConfig            `toml:"cbo"`             // [cbo] section
	Architecture   ArchitectureTomlConfig   `toml:"architecture"`    // [architecture] section
	SystemAnalysis SystemAnalysisTomlConfig `toml:"system_analysis"` // [system_analysis] section
	Dependencies   DependenciesTomlConfig   `toml:"dependencies"`    // [dependencies] section
	Clones         ClonesConfig             `toml:"clones"`          // [clones] section - unified flat structure
}

// ComplexityTomlConfig represents the [complexity] section
type ComplexityTomlConfig struct {
	LowThreshold    *int `toml:"low_threshold"`    // pointer to detect unset
	MediumThreshold *int `toml:"medium_threshold"` // pointer to detect unset
	MaxComplexity   *int `toml:"max_complexity"`   // pointer to detect unset
	MinComplexity   *int `toml:"min_complexity"`   // pointer to detect unset
}

// DeadCodeTomlConfig represents the [dead_code] section
type DeadCodeTomlConfig struct {
	Enabled                   *bool    `toml:"enabled"`
	MinSeverity               string   `toml:"min_severity"`
	ShowContext               *bool    `toml:"show_context"`
	ContextLines              *int     `toml:"context_lines"`
	SortBy                    string   `toml:"sort_by"`
	DetectAfterReturn         *bool    `toml:"detect_after_return"`
	DetectAfterBreak          *bool    `toml:"detect_after_break"`
	DetectAfterContinue       *bool    `toml:"detect_after_continue"`
	DetectAfterRaise          *bool    `toml:"detect_after_raise"`
	DetectUnreachableBranches *bool    `toml:"detect_unreachable_branches"`
	IgnorePatterns            []string `toml:"ignore_patterns"`
}

// OutputTomlConfig represents the [output] section
type OutputTomlConfig struct {
	Format        string `toml:"format"`
	ShowDetails   *bool  `toml:"show_details"`
	SortBy        string `toml:"sort_by"`
	MinComplexity *int   `toml:"min_complexity"`
	Directory     string `toml:"directory"`
}

// AnalysisTomlConfig represents the [analysis] section
type AnalysisTomlConfig struct {
	IncludePatterns []string `toml:"include_patterns"`
	ExcludePatterns []string `toml:"exclude_patterns"`
	Recursive       *bool    `toml:"recursive"`
	FollowSymlinks  *bool    `toml:"follow_symlinks"`
}

// CboTomlConfig represents the [cbo] section
type CboTomlConfig struct {
	LowThreshold    *int  `toml:"low_threshold"`
	MediumThreshold *int  `toml:"medium_threshold"`
	MinCbo          *int  `toml:"min_cbo"`
	MaxCbo          *int  `toml:"max_cbo"`
	ShowZeros       *bool `toml:"show_zeros"`
	IncludeBuiltins *bool `toml:"include_builtins"`
	IncludeImports  *bool `toml:"include_imports"`
}

// ArchitectureTomlConfig represents the [architecture] section
type ArchitectureTomlConfig struct {
	Enabled                         *bool                 `toml:"enabled"`
	ValidateLayers                  *bool                 `toml:"validate_layers"`
	ValidateCohesion                *bool                 `toml:"validate_cohesion"`
	ValidateResponsibility          *bool                 `toml:"validate_responsibility"`
	MinCohesion                     *float64              `toml:"min_cohesion"`
	MaxCoupling                     *int                  `toml:"max_coupling"`
	MaxResponsibilities             *int                  `toml:"max_responsibilities"`
	LayerViolationSeverity          string                `toml:"layer_violation_severity"`
	CohesionViolationSeverity       string                `toml:"cohesion_violation_severity"`
	ResponsibilityViolationSeverity string                `toml:"responsibility_violation_severity"`
	ShowAllViolations               *bool                 `toml:"show_all_violations"`
	GroupByType                     *bool                 `toml:"group_by_type"`
	IncludeSuggestions              *bool                 `toml:"include_suggestions"`
	MaxViolationsToShow             *int                  `toml:"max_violations_to_show"`
	CustomPatterns                  []string              `toml:"custom_patterns"`
	AllowedPatterns                 []string              `toml:"allowed_patterns"`
	ForbiddenPatterns               []string              `toml:"forbidden_patterns"`
	StrictMode                      *bool                 `toml:"strict_mode"`
	FailOnViolations                *bool                 `toml:"fail_on_violations"`
	Layers                          []LayerDefinitionToml `toml:"layers"`
	Rules                           []LayerRuleToml       `toml:"rules"`
}

// LayerDefinitionToml represents a layer definition in TOML
type LayerDefinitionToml struct {
	Name        string   `toml:"name"`
	Description string   `toml:"description"`
	Packages    []string `toml:"packages"`
}

// LayerRuleToml represents a layer rule in TOML
type LayerRuleToml struct {
	From  string   `toml:"from"`
	Allow []string `toml:"allow"`
	Deny  []string `toml:"deny"`
}

// SystemAnalysisTomlConfig represents the [system_analysis] section
type SystemAnalysisTomlConfig struct {
	Enabled               *bool `toml:"enabled"`
	EnableDependencies    *bool `toml:"enable_dependencies"`
	EnableArchitecture    *bool `toml:"enable_architecture"`
	UseComplexityData     *bool `toml:"use_complexity_data"`
	UseClonesData         *bool `toml:"use_clones_data"`
	UseDeadCodeData       *bool `toml:"use_dead_code_data"`
	GenerateUnifiedReport *bool `toml:"generate_unified_report"`
}

// DependenciesTomlConfig represents the [dependencies] section
type DependenciesTomlConfig struct {
	Enabled           *bool    `toml:"enabled"`
	IncludeStdLib     *bool    `toml:"include_stdlib"`
	IncludeThirdParty *bool    `toml:"include_third_party"`
	FollowRelative    *bool    `toml:"follow_relative"`
	DetectCycles      *bool    `toml:"detect_cycles"`
	CalculateMetrics  *bool    `toml:"calculate_metrics"`
	FindLongChains    *bool    `toml:"find_long_chains"`
	MinCoupling       *int     `toml:"min_coupling"`
	MaxCoupling       *int     `toml:"max_coupling"`
	MinInstability    *float64 `toml:"min_instability"`
	MaxDistance       *float64 `toml:"max_distance"`
	SortBy            string   `toml:"sort_by"`
	ShowMatrix        *bool    `toml:"show_matrix"`
	ShowMetrics       *bool    `toml:"show_metrics"`
	ShowChains        *bool    `toml:"show_chains"`
	GenerateDotGraph  *bool    `toml:"generate_dot_graph"`
	CycleReporting    string   `toml:"cycle_reporting"`
	MaxCyclesToShow   *int     `toml:"max_cycles_to_show"`
	ShowCyclePaths    *bool    `toml:"show_cycle_paths"`
}

// ClonesConfig represents the [clones] section (flat structure)
type ClonesConfig struct {
	// Analysis settings
	MinLines          int     `toml:"min_lines"`
	MinNodes          int     `toml:"min_nodes"`
	MaxEditDistance   float64 `toml:"max_edit_distance"`
	IgnoreLiterals    *bool   `toml:"ignore_literals"`    // pointer to detect unset
	IgnoreIdentifiers *bool   `toml:"ignore_identifiers"` // pointer to detect unset
	CostModelType     string  `toml:"cost_model_type"`

	// Thresholds
	Type1Threshold      float64 `toml:"type1_threshold"`
	Type2Threshold      float64 `toml:"type2_threshold"`
	Type3Threshold      float64 `toml:"type3_threshold"`
	Type4Threshold      float64 `toml:"type4_threshold"`
	SimilarityThreshold float64 `toml:"similarity_threshold"`

	// Filtering
	MinSimilarity     float64  `toml:"min_similarity"`
	MaxSimilarity     float64  `toml:"max_similarity"`
	EnabledCloneTypes []string `toml:"enabled_clone_types"`
	MaxResults        int      `toml:"max_results"`

	// Grouping
	GroupingMode      string  `toml:"grouping_mode"`
	GroupingThreshold float64 `toml:"grouping_threshold"`
	KCoreK            int     `toml:"k_core_k"`

	// LSH (flat structure with lsh_ prefix)
	LSHEnabled             string  `toml:"lsh_enabled"`
	LSHAutoThreshold       int     `toml:"lsh_auto_threshold"`
	LSHSimilarityThreshold float64 `toml:"lsh_similarity_threshold"`
	LSHBands               int     `toml:"lsh_bands"`
	LSHRows                int     `toml:"lsh_rows"`
	LSHHashes              int     `toml:"lsh_hashes"`

	// Performance
	MaxMemoryMB    int   `toml:"max_memory_mb"`
	BatchSize      int   `toml:"batch_size"`
	EnableBatching *bool `toml:"enable_batching"` // pointer to detect unset
	MaxGoroutines  int   `toml:"max_goroutines"`
	TimeoutSeconds int   `toml:"timeout_seconds"`

	// Input
	Paths           []string `toml:"paths"`
	Recursive       *bool    `toml:"recursive"` // pointer to detect unset
	IncludePatterns []string `toml:"include_patterns"`
	ExcludePatterns []string `toml:"exclude_patterns"`

	// Output
	Format      string `toml:"format"`
	ShowDetails *bool  `toml:"show_details"` // pointer to detect unset
	ShowContent *bool  `toml:"show_content"` // pointer to detect unset
	SortBy      string `toml:"sort_by"`
	GroupClones *bool  `toml:"group_clones"` // pointer to detect unset
}

// TomlConfigLoader handles TOML-only configuration loading
type TomlConfigLoader struct{}

// NewTomlConfigLoader creates a new TOML configuration loader
func NewTomlConfigLoader() *TomlConfigLoader {
	return &TomlConfigLoader{}
}

// LoadConfig loads configuration from TOML files with ruff-like priority:
// 1. .pyscn.toml (dedicated config file)
// 2. pyproject.toml (with [tool.pyscn] section)
// 3. defaults
func (l *TomlConfigLoader) LoadConfig(startDir string) (*PyscnConfig, error) {
	// Try .pyscn.toml first (highest priority)
	if config, err := l.loadFromPyscnToml(startDir); err == nil {
		return config, nil
	}

	// Try pyproject.toml as fallback
	if config, err := l.loadFromPyprojectToml(startDir); err == nil {
		return config, nil
	}

	// Return defaults if no config found
	return DefaultPyscnConfig(), nil
}

// loadFromPyprojectToml loads config from pyproject.toml
func (l *TomlConfigLoader) loadFromPyprojectToml(startDir string) (*PyscnConfig, error) {
	_, err := l.findPyprojectToml(startDir)
	if err != nil {
		return nil, err
	}

	return LoadPyprojectConfig(startDir)
}

// loadFromPyscnToml loads config from .pyscn.toml (dedicated config file)
func (l *TomlConfigLoader) loadFromPyscnToml(startDir string) (*PyscnConfig, error) {
	configPath, err := l.findPyscnToml(startDir)
	if err != nil {
		return nil, err
	}

	// Read and parse .pyscn.toml
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config PyscnTomlConfig
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Merge with defaults
	defaults := DefaultPyscnConfig()
	l.mergePyscnTomlConfigs(defaults, &config)

	return defaults, nil
}

// findPyprojectToml walks up the directory tree to find pyproject.toml
func (l *TomlConfigLoader) findPyprojectToml(startDir string) (string, error) {
	return findPyprojectToml(startDir) // Reuse existing function
}

// findPyscnToml walks up the directory tree to find .pyscn.toml
func (l *TomlConfigLoader) findPyscnToml(startDir string) (string, error) {
	dir := startDir
	for {
		configPath := filepath.Join(dir, ".pyscn.toml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	return "", os.ErrNotExist
}

// mergePyscnTomlConfigs merges .pyscn.toml config into defaults
// using pointer booleans to detect unset values
func (l *TomlConfigLoader) mergePyscnTomlConfigs(defaults *PyscnConfig, pyscnToml *PyscnTomlConfig) {
	// Merge from [complexity] section using shared merge logic
	mergeComplexitySection(defaults, &pyscnToml.Complexity)

	// Merge from [dead_code] section
	mergeDeadCodeSection(defaults, &pyscnToml.DeadCode)

	// Merge from [output] section
	mergeOutputSection(defaults, &pyscnToml.Output)

	// Merge from [analysis] section
	mergeAnalysisSection(defaults, &pyscnToml.Analysis)

	// Merge from [cbo] section
	mergeCboSection(defaults, &pyscnToml.Cbo)

	// Merge from [architecture] section
	mergeArchitectureSection(defaults, &pyscnToml.Architecture)

	// Merge from [system_analysis] section
	mergeSystemAnalysisSection(defaults, &pyscnToml.SystemAnalysis)

	// Merge from [dependencies] section
	mergeDependenciesSection(defaults, &pyscnToml.Dependencies)

	// Merge from [clones] section (unified flat structure)
	mergeClonesSection(defaults, &pyscnToml.Clones)
}

// mergeClonesSection is moved to pyproject_loader.go and is now shared
// between .pyscn.toml and pyproject.toml loaders

// GetSupportedConfigFiles returns the list of supported TOML config files
// in order of precedence
func (l *TomlConfigLoader) GetSupportedConfigFiles() []string {
	return []string{
		".pyscn.toml",    // dedicated config file (highest priority)
		"pyproject.toml", // with [tool.pyscn] section
	}
}
