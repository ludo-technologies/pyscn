package domain

import (
	"context"
	"fmt"
	"io"
	"time"
)

// SystemAnalysisRequest represents a request for comprehensive system-level analysis
type SystemAnalysisRequest struct {
	// Input files or directories to analyze
	Paths []string

	// ProjectRoot overrides automatic project-root inference when set.
	ProjectRoot string

	// Output configuration
	OutputFormat OutputFormat
	OutputWriter io.Writer
	OutputPath   string // Path to save output file
	NoOpen       bool   // Don't auto-open HTML in browser

	// Analysis scope
	AnalyzeDependencies *bool // Enable dependency analysis
	AnalyzeArchitecture *bool // Enable architecture validation

	// Configuration
	ConfigPath      string
	Recursive       *bool
	IncludePatterns []string
	ExcludePatterns []string

	// Analysis options
	IncludeStdLib                   *bool             // Include standard library dependencies
	IncludeThirdParty               *bool             // Include third-party dependencies
	FollowRelative                  *bool             // Follow relative imports
	DetectCycles                    *bool             // Detect circular dependencies
	ValidateArchitecture            *bool             // Validate architecture rules
	ValidateCohesion                *bool             // Validate package cohesion
	ValidateResponsibility          *bool             // Validate single responsibility boundaries
	MinCohesion                     float64           // Minimum acceptable package cohesion
	MaxResponsibilities             int               // Maximum inferred responsibilities per module
	CohesionViolationSeverity       ViolationSeverity // Severity for package cohesion violations
	ResponsibilityViolationSeverity ViolationSeverity // Severity for SRP violations

	// Architecture rules (loaded from config or specified directly)
	ArchitectureRules *ArchitectureRules

	// Integration with other analyses
	ComplexityData map[string]int     // Module -> average complexity
	ClonesData     map[string]float64 // Module -> duplication ratio
	DeadCodeData   map[string]int     // Module -> dead code lines
}

// SystemAnalysisResponse represents the complete system analysis result
type SystemAnalysisResponse struct {
	// Core analysis results
	DependencyAnalysis   *DependencyAnalysisResult   `json:"dependency_analysis" yaml:"dependency_analysis"`     // Module dependency analysis
	ArchitectureAnalysis *ArchitectureAnalysisResult `json:"architecture_analysis" yaml:"architecture_analysis"` // Architecture validation results

	// Summary information
	Summary SystemAnalysisSummary `json:"summary" yaml:"summary"` // High-level summary

	// Issues and recommendations
	Issues          []SystemIssue          `json:"issues" yaml:"issues"`                   // Critical issues found
	Recommendations []SystemRecommendation `json:"recommendations" yaml:"recommendations"` // Improvement recommendations
	Warnings        []string               `json:"warnings" yaml:"warnings"`               // Analysis warnings
	Errors          []string               `json:"errors" yaml:"errors"`                   // Analysis errors
	Failures        []AnalysisFailure      `json:"failures,omitempty" yaml:"failures,omitempty"`

	// Metadata
	GeneratedAt time.Time   `json:"generated_at" yaml:"generated_at"` // When the analysis was generated
	Duration    int64       `json:"duration" yaml:"duration"`         // Analysis duration in milliseconds
	Version     string      `json:"version" yaml:"version"`           // Tool version
	Config      interface{} `json:"config" yaml:"config"`             // Configuration used for analysis
}

// SystemAnalysisSummary provides a high-level overview of system quality
type SystemAnalysisSummary struct {
	// System overview
	TotalModules      int    `json:"total_modules" yaml:"total_modules"`           // Total number of modules analyzed
	TotalPackages     int    `json:"total_packages" yaml:"total_packages"`         // Total number of packages
	TotalDependencies int    `json:"total_dependencies" yaml:"total_dependencies"` // Total dependency relationships
	ProjectRoot       string `json:"project_root" yaml:"project_root"`             // Project root directory
	ResolvedImports   int    `json:"resolved_imports" yaml:"resolved_imports"`     // Internal imports resolved to analyzed modules
	UnresolvedImports int    `json:"unresolved_imports" yaml:"unresolved_imports"` // Internal imports that could not be resolved

	// Quality scores (0-100, higher is better)
	OverallQualityScore  float64 `json:"overall_quality_score" yaml:"overall_quality_score"` // Composite quality score
	MaintainabilityScore float64 `json:"maintainability_score" yaml:"maintainability_score"` // Average maintainability index
	ArchitectureScore    float64 `json:"architecture_score" yaml:"architecture_score"`       // Architecture compliance score
	ModularityScore      float64 `json:"modularity_score" yaml:"modularity_score"`           // System modularity score
	TechnicalDebtHours   float64 `json:"technical_debt_hours" yaml:"technical_debt_hours"`   // Total estimated technical debt

	// Key metrics
	AverageCoupling        float64 `json:"average_coupling" yaml:"average_coupling"`               // Average module coupling
	AverageInstability     float64 `json:"average_instability" yaml:"average_instability"`         // Average instability
	CyclicDependencies     int     `json:"cyclic_dependencies" yaml:"cyclic_dependencies"`         // Number of modules in cycles
	ArchitectureViolations int     `json:"architecture_violations" yaml:"architecture_violations"` // Number of architecture rule violations
	HighRiskModules        int     `json:"high_risk_modules" yaml:"high_risk_modules"`             // Number of high-risk modules

	// Recommendations summary
	CriticalIssues           int `json:"critical_issues" yaml:"critical_issues"`                     // Number of critical issues requiring immediate attention
	RefactoringCandidates    int `json:"refactoring_candidates" yaml:"refactoring_candidates"`       // Number of modules needing refactoring
	ArchitectureImprovements int `json:"architecture_improvements" yaml:"architecture_improvements"` // Number of architecture improvements suggested
}

// DependencyAnalysisResult contains module dependency analysis results
type DependencyAnalysisResult struct {
	// Dependency graph information
	TotalModules      int      `json:"total_modules" yaml:"total_modules"`           // Total number of modules
	TotalDependencies int      `json:"total_dependencies" yaml:"total_dependencies"` // Total number of dependencies
	ResolvedImports   int      `json:"resolved_imports" yaml:"resolved_imports"`     // Internal imports resolved to analyzed modules
	UnresolvedImports int      `json:"unresolved_imports" yaml:"unresolved_imports"` // Internal imports that could not be resolved
	RootModules       []string `json:"root_modules" yaml:"root_modules"`             // Modules with no dependencies
	LeafModules       []string `json:"leaf_modules" yaml:"leaf_modules"`             // Modules with no dependents

	// Dependency metrics
	ModuleMetrics    map[string]*ModuleDependencyMetrics `json:"module_metrics" yaml:"module_metrics"`       // Per-module metrics
	DependencyMatrix map[string]map[string]bool          `json:"dependency_matrix" yaml:"dependency_matrix"` // Module -> dependencies

	// Circular dependency analysis
	CircularDependencies *CircularDependencyAnalysis `json:"circular_dependencies" yaml:"circular_dependencies"` // Circular dependency results

	// Coupling analysis
	CouplingAnalysis *CouplingAnalysis `json:"coupling_analysis" yaml:"coupling_analysis"` // Detailed coupling analysis

	// Dependency chains
	LongestChains []DependencyPath `json:"longest_chains" yaml:"longest_chains"` // Longest paths through the load-time SCC-condensed dependency graph
	MaxDepth      int              `json:"max_depth" yaml:"max_depth"`           // Edges along the chain LongestChains ranks first
}

// ModuleDependencyMetrics contains dependency metrics for a single module
type ModuleDependencyMetrics struct {
	// Basic information
	ModuleName string `json:"module_name" yaml:"module_name"` // Module name
	Package    string `json:"package" yaml:"package"`         // Package name
	FilePath   string `json:"file_path" yaml:"file_path"`     // File path
	IsPackage  bool   `json:"is_package" yaml:"is_package"`   // True if this is a package

	// Size metrics
	LinesOfCode        int      `json:"lines_of_code" yaml:"lines_of_code"`               // Total lines of code
	FunctionCount      int      `json:"function_count" yaml:"function_count"`             // Number of functions
	ClassCount         int      `json:"class_count" yaml:"class_count"`                   // Number of classes
	AbstractClassCount int      `json:"abstract_class_count" yaml:"abstract_class_count"` // Number of abstract classes
	PublicInterface    []string `json:"public_interface" yaml:"public_interface"`         // Public names exported

	// Coupling metrics (Robert Martin's metrics)
	AfferentCoupling int     `json:"afferent_coupling" yaml:"afferent_coupling"` // Ca - modules that depend on this one
	EfferentCoupling int     `json:"efferent_coupling" yaml:"efferent_coupling"` // Ce - modules this one depends on
	Instability      float64 `json:"instability" yaml:"instability"`             // I = Ce / (Ca + Ce)
	Abstractness     float64 `json:"abstractness" yaml:"abstractness"`           // A - abstractness measure
	Distance         float64 `json:"distance" yaml:"distance"`                   // D - distance from main sequence

	// Quality metrics
	Maintainability float64   `json:"maintainability" yaml:"maintainability"` // Maintainability index (0-100)
	TechnicalDebt   float64   `json:"technical_debt" yaml:"technical_debt"`   // Estimated technical debt in hours
	RiskLevel       RiskLevel `json:"risk_level" yaml:"risk_level"`           // Overall risk assessment

	// Dependencies
	DirectDependencies     []string `json:"direct_dependencies" yaml:"direct_dependencies"`         // Modules this directly depends on
	TransitiveDependencies []string `json:"transitive_dependencies" yaml:"transitive_dependencies"` // All transitive dependencies
	Dependents             []string `json:"dependents" yaml:"dependents"`                           // Modules that depend on this one
}

// CircularDependencyAnalysis contains circular dependency analysis results
type CircularDependencyAnalysis struct {
	HasCircularDependencies  bool                 `json:"has_circular_dependencies" yaml:"has_circular_dependencies"`   // True if cycles exist
	TotalCycles              int                  `json:"total_cycles" yaml:"total_cycles"`                             // Number of circular dependencies
	TotalModulesInCycles     int                  `json:"total_modules_in_cycles" yaml:"total_modules_in_cycles"`       // Number of modules involved in cycles
	CircularDependencies     []CircularDependency `json:"circular_dependencies" yaml:"circular_dependencies"`           // All detected cycles
	CycleBreakingSuggestions []string             `json:"cycle_breaking_suggestions" yaml:"cycle_breaking_suggestions"` // Suggestions for breaking cycles
	CoreInfrastructure       []string             `json:"core_infrastructure" yaml:"core_infrastructure"`               // Modules in multiple cycles
}

// CircularDependency represents a circular dependency
type CircularDependency struct {
	Modules      []string         `json:"modules" yaml:"modules"`           // Modules in the cycle
	Dependencies []DependencyPath `json:"dependencies" yaml:"dependencies"` // Dependency paths forming the cycle
	Severity     CycleSeverity    `json:"severity" yaml:"severity"`         // Severity level
	Size         int              `json:"size" yaml:"size"`                 // Number of modules
	Description  string           `json:"description" yaml:"description"`   // Human-readable description
}

// DependencyPath represents a path of dependencies
type DependencyPath struct {
	From   string   `json:"from" yaml:"from"`     // Starting module
	To     string   `json:"to" yaml:"to"`         // Ending module
	Path   []string `json:"path" yaml:"path"`     // Complete path
	Length int      `json:"length" yaml:"length"` // Path length
}

// CycleSeverity represents severity of circular dependencies
type CycleSeverity string

const (
	CycleSeverityLow      CycleSeverity = "low"
	CycleSeverityMedium   CycleSeverity = "medium"
	CycleSeverityHigh     CycleSeverity = "high"
	CycleSeverityCritical CycleSeverity = "critical"
)

// CouplingAnalysis contains detailed coupling analysis
type CouplingAnalysis struct {
	// Overall coupling metrics
	AverageCoupling       float64     `json:"average_coupling" yaml:"average_coupling"`               // Average coupling across all modules
	CouplingDistribution  map[int]int `json:"coupling_distribution" yaml:"coupling_distribution"`     // Coupling value -> count
	HighlyCoupledModules  []string    `json:"highly_coupled_modules" yaml:"highly_coupled_modules"`   // Modules with high coupling
	LooselyCoupledModules []string    `json:"loosely_coupled_modules" yaml:"loosely_coupled_modules"` // Modules with low coupling

	// Instability analysis
	AverageInstability float64  `json:"average_instability" yaml:"average_instability"` // Average instability
	StableModules      []string `json:"stable_modules" yaml:"stable_modules"`           // Low instability modules
	InstableModules    []string `json:"instable_modules" yaml:"instable_modules"`       // High instability modules

	// Main sequence analysis
	MainSequenceDeviation float64  `json:"main_sequence_deviation" yaml:"main_sequence_deviation"` // Average distance from main sequence
	ZoneOfPain            []string `json:"zone_of_pain" yaml:"zone_of_pain"`                       // Stable + concrete modules
	ZoneOfUselessness     []string `json:"zone_of_uselessness" yaml:"zone_of_uselessness"`         // Unstable + abstract modules
	MainSequence          []string `json:"main_sequence" yaml:"main_sequence"`                     // Well-positioned modules
}

// ArchitectureAnalysisResult contains architecture validation results
type ArchitectureAnalysisResult struct {
	// Overall architecture compliance
	ComplianceScore    float64 `json:"compliance_score" yaml:"compliance_score"`       // Overall compliance score (0-1, where 1.0 = 100% compliant). Computed as 1 - WeightedViolations/TotalRules (clamped to [0,1]).
	TotalViolations    int     `json:"total_violations" yaml:"total_violations"`       // Raw number of violations (one per ArchitectureViolation entry).
	WeightedViolations int     `json:"weighted_violations" yaml:"weighted_violations"` // Severity-weighted violation count used as the ComplianceScore numerator: error * 5 + warning * 1.
	TotalRules         int     `json:"total_rules" yaml:"total_rules"`                 // Total number of rule invocations checked (ComplianceScore denominator).

	// Layer analysis
	LayerAnalysis          *LayerAnalysis          `json:"layer_analysis" yaml:"layer_analysis"`                   // Layer violation analysis
	CohesionAnalysis       *CohesionAnalysis       `json:"cohesion_analysis" yaml:"cohesion_analysis"`             // Package cohesion analysis
	ResponsibilityAnalysis *ResponsibilityAnalysis `json:"responsibility_analysis" yaml:"responsibility_analysis"` // SRP violation analysis

	// Detailed violations
	Violations        []ArchitectureViolation   `json:"violations" yaml:"violations"`                 // All architecture violations
	SeverityBreakdown map[ViolationSeverity]int `json:"severity_breakdown" yaml:"severity_breakdown"` // Violations by severity

	// Architecture recommendations
	Recommendations    []ArchitectureRecommendation `json:"recommendations" yaml:"recommendations"`         // Specific recommendations
	RefactoringTargets []string                     `json:"refactoring_targets" yaml:"refactoring_targets"` // Modules needing refactoring
}

// LayerAnalysis contains layer architecture validation results
type LayerAnalysis struct {
	LayersAnalyzed    int                       `json:"layers_analyzed" yaml:"layers_analyzed"`       // Number of layers analyzed
	LayerViolations   []LayerViolation          `json:"layer_violations" yaml:"layer_violations"`     // Layer rule violations
	LayerCoupling     map[string]map[string]int `json:"layer_coupling" yaml:"layer_coupling"`         // Layer -> Layer -> dependency count
	LayerCohesion     map[string]float64        `json:"layer_cohesion" yaml:"layer_cohesion"`         // Layer -> cohesion score
	ProblematicLayers []string                  `json:"problematic_layers" yaml:"problematic_layers"` // Layers with violations
}

// LayerViolation represents a layer architecture rule violation
type LayerViolation struct {
	FromModule  string            `json:"from_module" yaml:"from_module"` // Module causing violation
	ToModule    string            `json:"to_module" yaml:"to_module"`     // Target module
	FromLayer   string            `json:"from_layer" yaml:"from_layer"`   // Source layer
	ToLayer     string            `json:"to_layer" yaml:"to_layer"`       // Target layer
	Rule        string            `json:"rule" yaml:"rule"`               // Rule that was violated
	Severity    ViolationSeverity `json:"severity" yaml:"severity"`       // Severity of violation
	Description string            `json:"description" yaml:"description"` // Description of violation
	Suggestion  string            `json:"suggestion" yaml:"suggestion"`   // Suggested fix
}

// CohesionAnalysis contains package cohesion analysis
type CohesionAnalysis struct {
	PackageCohesion     map[string]float64 `json:"package_cohesion" yaml:"package_cohesion"`           // Package -> cohesion score
	LowCohesionPackages []string           `json:"low_cohesion_packages" yaml:"low_cohesion_packages"` // Packages with low cohesion
	CohesionSuggestions map[string]string  `json:"cohesion_suggestions" yaml:"cohesion_suggestions"`   // Package -> suggestion
}

// ResponsibilityAnalysis contains Single Responsibility Principle analysis
type ResponsibilityAnalysis struct {
	SRPViolations          []SRPViolation      `json:"srp_violations" yaml:"srp_violations"`                   // SRP violations detected
	ModuleResponsibilities map[string][]string `json:"module_responsibilities" yaml:"module_responsibilities"` // Module -> responsibilities
	OverloadedModules      []string            `json:"overloaded_modules" yaml:"overloaded_modules"`           // Modules with too many responsibilities
}

// SRPViolation represents a Single Responsibility Principle violation
type SRPViolation struct {
	Module           string            `json:"module" yaml:"module"`                     // Module with violation
	Responsibilities []string          `json:"responsibilities" yaml:"responsibilities"` // Multiple responsibilities detected
	Severity         ViolationSeverity `json:"severity" yaml:"severity"`                 // Severity level
	Suggestion       string            `json:"suggestion" yaml:"suggestion"`             // Refactoring suggestion
}

// ArchitectureViolation represents an architecture rule violation
type ArchitectureViolation struct {
	Type        ViolationType     `json:"type" yaml:"type"`               // Type of violation
	Severity    ViolationSeverity `json:"severity" yaml:"severity"`       // Severity level
	Module      string            `json:"module" yaml:"module"`           // Module involved
	Target      string            `json:"target" yaml:"target"`           // Target of violation (if applicable)
	Rule        string            `json:"rule" yaml:"rule"`               // Rule that was violated
	Description string            `json:"description" yaml:"description"` // Human-readable description
	Suggestion  string            `json:"suggestion" yaml:"suggestion"`   // Suggested remediation
	Location    *SourceLocation   `json:"location" yaml:"location"`       // Location in code (if available)
}

// ViolationType represents the type of architecture violation
type ViolationType string

const (
	ViolationTypeLayer          ViolationType = "layer"          // Layer dependency violation
	ViolationTypeCycle          ViolationType = "cycle"          // Circular dependency
	ViolationTypeCoupling       ViolationType = "coupling"       // Excessive coupling
	ViolationTypeResponsibility ViolationType = "responsibility" // SRP violation
	ViolationTypeCohesion       ViolationType = "cohesion"       // Low cohesion
)

// ViolationSeverity represents the severity of a violation
type ViolationSeverity string

const (
	ViolationSeverityInfo     ViolationSeverity = "info"
	ViolationSeverityWarning  ViolationSeverity = "warning"
	ViolationSeverityError    ViolationSeverity = "error"
	ViolationSeverityCritical ViolationSeverity = "critical"
)

// ArchitectureRecommendation represents a specific architecture improvement recommendation
type ArchitectureRecommendation struct {
	Type        RecommendationType     `json:"type" yaml:"type"`               // Type of recommendation
	Priority    RecommendationPriority `json:"priority" yaml:"priority"`       // Priority level
	Title       string                 `json:"title" yaml:"title"`             // Short title
	Description string                 `json:"description" yaml:"description"` // Detailed description
	Benefits    []string               `json:"benefits" yaml:"benefits"`       // Expected benefits
	Effort      EstimatedEffort        `json:"effort" yaml:"effort"`           // Estimated effort
	Modules     []string               `json:"modules" yaml:"modules"`         // Affected modules
	Steps       []string               `json:"steps" yaml:"steps"`             // Implementation steps
}

// RecommendationType represents the type of recommendation
type RecommendationType string

const (
	RecommendationTypeRefactor    RecommendationType = "refactor"    // Code refactoring
	RecommendationTypeRestructure RecommendationType = "restructure" // Architectural restructuring
	RecommendationTypeExtract     RecommendationType = "extract"     // Extract module/package
	RecommendationTypeMerge       RecommendationType = "merge"       // Merge modules
	RecommendationTypeInterface   RecommendationType = "interface"   // Add abstraction
)

// RecommendationPriority represents priority level
type RecommendationPriority string

const (
	RecommendationPriorityLow      RecommendationPriority = "low"
	RecommendationPriorityMedium   RecommendationPriority = "medium"
	RecommendationPriorityHigh     RecommendationPriority = "high"
	RecommendationPriorityCritical RecommendationPriority = "critical"
)

// EstimatedEffort represents estimated implementation effort
type EstimatedEffort string

const (
	EstimatedEffortLow    EstimatedEffort = "low"    // < 4 hours
	EstimatedEffortMedium EstimatedEffort = "medium" // 4-16 hours
	EstimatedEffortHigh   EstimatedEffort = "high"   // 16-40 hours
	EstimatedEffortLarge  EstimatedEffort = "large"  // > 40 hours
)

// SystemIssue represents a critical system-level issue
type SystemIssue struct {
	Type        IssueType     `json:"type" yaml:"type"`               // Type of issue
	Severity    IssueSeverity `json:"severity" yaml:"severity"`       // Severity level
	Title       string        `json:"title" yaml:"title"`             // Issue title
	Description string        `json:"description" yaml:"description"` // Detailed description
	Impact      string        `json:"impact" yaml:"impact"`           // Impact description
	Modules     []string      `json:"modules" yaml:"modules"`         // Affected modules
	Suggestion  string        `json:"suggestion" yaml:"suggestion"`   // Remediation suggestion
}

// IssueType represents the type of system issue
type IssueType string

const (
	IssueTypeCircularDependency    IssueType = "circular_dependency"
	IssueTypeExcessiveCoupling     IssueType = "excessive_coupling"
	IssueTypeArchitectureViolation IssueType = "architecture_violation"
	IssueTypePoorModularity        IssueType = "poor_modularity"
)

// IssueSeverity represents issue severity
type IssueSeverity string

const (
	IssueSeverityLow      IssueSeverity = "low"
	IssueSeverityMedium   IssueSeverity = "medium"
	IssueSeverityHigh     IssueSeverity = "high"
	IssueSeverityCritical IssueSeverity = "critical"
)

// SystemRecommendation represents a system-level improvement recommendation
type SystemRecommendation struct {
	Category    RecommendationCategory `json:"category" yaml:"category"`       // Category of recommendation
	Priority    RecommendationPriority `json:"priority" yaml:"priority"`       // Priority level
	Title       string                 `json:"title" yaml:"title"`             // Recommendation title
	Description string                 `json:"description" yaml:"description"` // Detailed description
	Rationale   string                 `json:"rationale" yaml:"rationale"`     // Why this is recommended
	Benefits    []string               `json:"benefits" yaml:"benefits"`       // Expected benefits
	Steps       []string               `json:"steps" yaml:"steps"`             // Implementation steps
	Resources   []string               `json:"resources" yaml:"resources"`     // Additional resources
	Effort      EstimatedEffort        `json:"effort" yaml:"effort"`           // Estimated effort
}

// RecommendationCategory represents recommendation category
type RecommendationCategory string

const (
	RecommendationCategoryArchitecture  RecommendationCategory = "architecture"
	RecommendationCategoryRefactoring   RecommendationCategory = "refactoring"
	RecommendationCategoryTesting       RecommendationCategory = "testing"
	RecommendationCategoryDocumentation RecommendationCategory = "documentation"
	RecommendationCategoryProcess       RecommendationCategory = "process"
)

// ArchitectureRules defines architecture validation rules
type ArchitectureRules struct {
	// Style is an optional preset name: "layered", "hexagonal", "clean", "mvc".
	// When non-empty and Layers/Rules are empty, the service loads the matching
	// preset's layers and rules. Empty/"layered" preserves the legacy behavior.
	Style string `json:"style" yaml:"style"`

	// Layer rules
	Layers []Layer     `json:"layers" yaml:"layers"`
	Rules  []LayerRule `json:"rules" yaml:"rules"`

	// Package rules
	PackageRules []PackageRule `json:"package_rules" yaml:"package_rules"`

	// Custom rules
	CustomRules []CustomRule `json:"custom_rules" yaml:"custom_rules"`

	// Neutral prefixes to strip from module names before layer matching
	NeutralPrefixes []string `json:"neutral_prefixes" yaml:"neutral_prefixes"`

	// Global settings
	StrictMode        bool     `json:"strict_mode" yaml:"strict_mode"`
	AllowedPatterns   []string `json:"allowed_patterns" yaml:"allowed_patterns"`
	ForbiddenPatterns []string `json:"forbidden_patterns" yaml:"forbidden_patterns"`
}

// Layer defines an architectural layer
type Layer struct {
	Name        string   `json:"name" yaml:"name"`
	Packages    []string `json:"packages" yaml:"packages"`
	Description string   `json:"description" yaml:"description"`
}

// LayerRule defines a dependency rule between layers
type LayerRule struct {
	From  string   `json:"from" yaml:"from"`
	Allow []string `json:"allow" yaml:"allow"`
	Deny  []string `json:"deny" yaml:"deny"`
	// Warn lists target layers that are permitted but discouraged: a dependency
	// on one of these emits a warning instead of an error. Used e.g. by the MVC
	// preset for view -> model direct access.
	Warn []string `json:"warn" yaml:"warn"`
}

// PackageRule defines rules for packages
type PackageRule struct {
	Package             string   `json:"package" yaml:"package"`
	MaxSize             int      `json:"max_size" yaml:"max_size"`
	MaxCoupling         int      `json:"max_coupling" yaml:"max_coupling"`
	MinCohesion         float64  `json:"min_cohesion" yaml:"min_cohesion"`
	AllowedDependencies []string `json:"allowed_dependencies" yaml:"allowed_dependencies"`
}

// CustomRule defines custom validation rules
type CustomRule struct {
	Name        string            `json:"name" yaml:"name"`
	Pattern     string            `json:"pattern" yaml:"pattern"`
	Description string            `json:"description" yaml:"description"`
	Severity    ViolationSeverity `json:"severity" yaml:"severity"`
}

// Service interfaces

// SystemAnalysisService defines the core business logic for system analysis
type SystemAnalysisService interface {
	// Analyze performs comprehensive system analysis
	Analyze(ctx context.Context, req SystemAnalysisRequest) (*SystemAnalysisResponse, error)

	// AnalyzeDependencies performs dependency analysis only
	AnalyzeDependencies(ctx context.Context, req SystemAnalysisRequest) (*DependencyAnalysisResult, error)

	// AnalyzeArchitecture performs architecture validation only
	AnalyzeArchitecture(ctx context.Context, req SystemAnalysisRequest) (*ArchitectureAnalysisResult, error)
}

// SystemAnalysisConfigurationLoader defines configuration loading interface
type SystemAnalysisConfigurationLoader interface {
	// LoadConfig loads configuration from the specified path
	LoadConfig(path string) (*SystemAnalysisRequest, error)

	// LoadDefaultConfig discovers configuration from targetPath (the analyzed
	// path) and falls back to built-in defaults when none is found
	LoadDefaultConfig(targetPath string) *SystemAnalysisRequest

	// MergeConfig merges CLI flags with configuration file
	MergeConfig(base *SystemAnalysisRequest, override *SystemAnalysisRequest) *SystemAnalysisRequest
}

// SystemAnalysisOutputFormatter defines formatting interface
type SystemAnalysisOutputFormatter interface {
	// Format formats the analysis response according to the specified format
	Format(response *SystemAnalysisResponse, format OutputFormat) (string, error)

	// Write writes the formatted output to the writer
	Write(response *SystemAnalysisResponse, format OutputFormat, writer io.Writer) error
}

// DefaultSystemAnalysisRequest returns a SystemAnalysisRequest with default values
func DefaultSystemAnalysisRequest() *SystemAnalysisRequest {
	return &SystemAnalysisRequest{
		OutputFormat:                    OutputFormatText,
		AnalyzeDependencies:             BoolPtr(true),
		AnalyzeArchitecture:             BoolPtr(true),
		Recursive:                       BoolPtr(true),
		IncludeStdLib:                   BoolPtr(false),
		IncludeThirdParty:               BoolPtr(true),
		FollowRelative:                  BoolPtr(true),
		DetectCycles:                    BoolPtr(true),
		ValidateArchitecture:            BoolPtr(true),
		ValidateCohesion:                BoolPtr(true),
		ValidateResponsibility:          BoolPtr(true),
		MinCohesion:                     DefaultArchitectureMinCohesion,
		MaxResponsibilities:             DefaultArchitectureMaxResponsibilities,
		CohesionViolationSeverity:       ViolationSeverityWarning,
		ResponsibilityViolationSeverity: ViolationSeverityWarning,
		IncludePatterns:                 DefaultPythonModuleIncludePatterns(),
		ExcludePatterns:                 DefaultAnalysisExcludePatterns(),
		ComplexityData:                  make(map[string]int),
		ClonesData:                      make(map[string]float64),
		DeadCodeData:                    make(map[string]int),
	}
}

// SourceLocation represents a location in source code
type SourceLocation struct {
	FilePath  string `json:"file_path" yaml:"file_path"`
	StartLine int    `json:"start_line" yaml:"start_line"`
	EndLine   int    `json:"end_line" yaml:"end_line"`
	StartCol  int    `json:"start_col" yaml:"start_col"`
	EndCol    int    `json:"end_col" yaml:"end_col"`
}

// String returns string representation of SourceLocation
func (sl *SourceLocation) String() string {
	if sl.StartCol > 0 && sl.EndCol > 0 {
		return fmt.Sprintf("%s:%d:%d-%d:%d", sl.FilePath, sl.StartLine, sl.StartCol, sl.EndLine, sl.EndCol)
	}
	return fmt.Sprintf("%s:%d-%d", sl.FilePath, sl.StartLine, sl.EndLine)
}

// LineCount returns the number of lines in this location
func (sl *SourceLocation) LineCount() int {
	return sl.EndLine - sl.StartLine + 1
}
