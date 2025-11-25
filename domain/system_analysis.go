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
	IncludeStdLib        *bool // Include standard library dependencies
	IncludeThirdParty    *bool // Include third-party dependencies
	FollowRelative       *bool // Follow relative imports
	DetectCycles         *bool // Detect circular dependencies
	ValidateArchitecture *bool // Validate architecture rules

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
	DependencyAnalysis   *DependencyAnalysisResult   // Module dependency analysis
	ArchitectureAnalysis *ArchitectureAnalysisResult // Architecture validation results

	// Summary information
	Summary SystemAnalysisSummary // High-level summary

	// Issues and recommendations
	Issues          []SystemIssue          // Critical issues found
	Recommendations []SystemRecommendation // Improvement recommendations
	Warnings        []string               // Analysis warnings
	Errors          []string               // Analysis errors

	// Metadata
	GeneratedAt time.Time   // When the analysis was generated
	Duration    int64       // Analysis duration in milliseconds
	Version     string      // Tool version
	Config      interface{} // Configuration used for analysis
}

// SystemAnalysisSummary provides a high-level overview of system quality
type SystemAnalysisSummary struct {
	// System overview
	TotalModules      int    // Total number of modules analyzed
	TotalPackages     int    // Total number of packages
	TotalDependencies int    // Total dependency relationships
	ProjectRoot       string // Project root directory

	// Quality scores (0-100, higher is better)
	OverallQualityScore  float64 // Composite quality score
	MaintainabilityScore float64 // Average maintainability index
	ArchitectureScore    float64 // Architecture compliance score
	ModularityScore      float64 // System modularity score
	TechnicalDebtHours   float64 // Total estimated technical debt

	// Key metrics
	AverageCoupling        float64 // Average module coupling
	AverageInstability     float64 // Average instability
	CyclicDependencies     int     // Number of modules in cycles
	ArchitectureViolations int     // Number of architecture rule violations
	HighRiskModules        int     // Number of high-risk modules

	// Recommendations summary
	CriticalIssues           int // Number of critical issues requiring immediate attention
	RefactoringCandidates    int // Number of modules needing refactoring
	ArchitectureImprovements int // Number of architecture improvements suggested
}

// DependencyAnalysisResult contains module dependency analysis results
type DependencyAnalysisResult struct {
	// Dependency graph information
	TotalModules      int      // Total number of modules
	TotalDependencies int      // Total number of dependencies
	RootModules       []string // Modules with no dependencies
	LeafModules       []string // Modules with no dependents

	// Dependency metrics
	ModuleMetrics    map[string]*ModuleDependencyMetrics // Per-module metrics
	DependencyMatrix map[string]map[string]bool          // Module -> dependencies

	// Circular dependency analysis
	CircularDependencies *CircularDependencyAnalysis // Circular dependency results

	// Coupling analysis
	CouplingAnalysis *CouplingAnalysis // Detailed coupling analysis

	// Dependency chains
	LongestChains []DependencyPath // Longest dependency chains
	MaxDepth      int              // Maximum dependency depth
}

// ModuleDependencyMetrics contains dependency metrics for a single module
type ModuleDependencyMetrics struct {
	// Basic information
	ModuleName string // Module name
	Package    string // Package name
	FilePath   string // File path
	IsPackage  bool   // True if this is a package

	// Size metrics
	LinesOfCode     int      // Total lines of code
	FunctionCount   int      // Number of functions
	ClassCount      int      // Number of classes
	PublicInterface []string // Public names exported

	// Coupling metrics (Robert Martin's metrics)
	AfferentCoupling int     // Ca - modules that depend on this one
	EfferentCoupling int     // Ce - modules this one depends on
	Instability      float64 // I = Ce / (Ca + Ce)
	Abstractness     float64 // A - abstractness measure
	Distance         float64 // D - distance from main sequence

	// Quality metrics
	Maintainability float64   // Maintainability index (0-100)
	TechnicalDebt   float64   // Estimated technical debt in hours
	RiskLevel       RiskLevel // Overall risk assessment

	// Dependencies
	DirectDependencies     []string // Modules this directly depends on
	TransitiveDependencies []string // All transitive dependencies
	Dependents             []string // Modules that depend on this one
}

// CircularDependencyAnalysis contains circular dependency analysis results
type CircularDependencyAnalysis struct {
	HasCircularDependencies  bool                 // True if cycles exist
	TotalCycles              int                  // Number of circular dependencies
	TotalModulesInCycles     int                  // Number of modules involved in cycles
	CircularDependencies     []CircularDependency // All detected cycles
	CycleBreakingSuggestions []string             // Suggestions for breaking cycles
	CoreInfrastructure       []string             // Modules in multiple cycles
}

// CircularDependency represents a circular dependency
type CircularDependency struct {
	Modules      []string         // Modules in the cycle
	Dependencies []DependencyPath // Dependency paths forming the cycle
	Severity     CycleSeverity    // Severity level
	Size         int              // Number of modules
	Description  string           // Human-readable description
}

// DependencyPath represents a path of dependencies
type DependencyPath struct {
	From   string   // Starting module
	To     string   // Ending module
	Path   []string // Complete path
	Length int      // Path length
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
	AverageCoupling       float64     // Average coupling across all modules
	CouplingDistribution  map[int]int // Coupling value -> count
	HighlyCoupledModules  []string    // Modules with high coupling
	LooselyCoupledModules []string    // Modules with low coupling

	// Instability analysis
	AverageInstability float64  // Average instability
	StableModules      []string // Low instability modules
	InstableModules    []string // High instability modules

	// Main sequence analysis
	MainSequenceDeviation float64  // Average distance from main sequence
	ZoneOfPain            []string // Stable + concrete modules
	ZoneOfUselessness     []string // Unstable + abstract modules
	MainSequence          []string // Well-positioned modules
}

// ArchitectureAnalysisResult contains architecture validation results
type ArchitectureAnalysisResult struct {
	// Overall architecture compliance
	ComplianceScore float64 // Overall compliance score (0-1, where 1.0 = 100% compliant)
	TotalViolations int     // Total number of violations
	TotalRules      int     // Total number of rules checked

	// Layer analysis
	LayerAnalysis          *LayerAnalysis          // Layer violation analysis
	CohesionAnalysis       *CohesionAnalysis       // Package cohesion analysis
	ResponsibilityAnalysis *ResponsibilityAnalysis // SRP violation analysis

	// Detailed violations
	Violations        []ArchitectureViolation   // All architecture violations
	SeverityBreakdown map[ViolationSeverity]int // Violations by severity

	// Architecture recommendations
	Recommendations    []ArchitectureRecommendation // Specific recommendations
	RefactoringTargets []string                     // Modules needing refactoring
}

// LayerAnalysis contains layer architecture validation results
type LayerAnalysis struct {
	LayersAnalyzed    int                       // Number of layers analyzed
	LayerViolations   []LayerViolation          // Layer rule violations
	LayerCoupling     map[string]map[string]int // Layer -> Layer -> dependency count
	LayerCohesion     map[string]float64        // Layer -> cohesion score
	ProblematicLayers []string                  // Layers with violations
}

// LayerViolation represents a layer architecture rule violation
type LayerViolation struct {
	FromModule  string            // Module causing violation
	ToModule    string            // Target module
	FromLayer   string            // Source layer
	ToLayer     string            // Target layer
	Rule        string            // Rule that was violated
	Severity    ViolationSeverity // Severity of violation
	Description string            // Description of violation
	Suggestion  string            // Suggested fix
}

// CohesionAnalysis contains package cohesion analysis
type CohesionAnalysis struct {
	PackageCohesion     map[string]float64 // Package -> cohesion score
	LowCohesionPackages []string           // Packages with low cohesion
	CohesionSuggestions map[string]string  // Package -> suggestion
}

// ResponsibilityAnalysis contains Single Responsibility Principle analysis
type ResponsibilityAnalysis struct {
	SRPViolations          []SRPViolation      // SRP violations detected
	ModuleResponsibilities map[string][]string // Module -> responsibilities
	OverloadedModules      []string            // Modules with too many responsibilities
}

// SRPViolation represents a Single Responsibility Principle violation
type SRPViolation struct {
	Module           string            // Module with violation
	Responsibilities []string          // Multiple responsibilities detected
	Severity         ViolationSeverity // Severity level
	Suggestion       string            // Refactoring suggestion
}

// ArchitectureViolation represents an architecture rule violation
type ArchitectureViolation struct {
	Type        ViolationType     // Type of violation
	Severity    ViolationSeverity // Severity level
	Module      string            // Module involved
	Target      string            // Target of violation (if applicable)
	Rule        string            // Rule that was violated
	Description string            // Human-readable description
	Suggestion  string            // Suggested remediation
	Location    *SourceLocation   // Location in code (if available)
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
	Type        RecommendationType     // Type of recommendation
	Priority    RecommendationPriority // Priority level
	Title       string                 // Short title
	Description string                 // Detailed description
	Benefits    []string               // Expected benefits
	Effort      EstimatedEffort        // Estimated effort
	Modules     []string               // Affected modules
	Steps       []string               // Implementation steps
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
	Type        IssueType     // Type of issue
	Severity    IssueSeverity // Severity level
	Title       string        // Issue title
	Description string        // Detailed description
	Impact      string        // Impact description
	Modules     []string      // Affected modules
	Suggestion  string        // Remediation suggestion
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
	Category    RecommendationCategory // Category of recommendation
	Priority    RecommendationPriority // Priority level
	Title       string                 // Recommendation title
	Description string                 // Detailed description
	Rationale   string                 // Why this is recommended
	Benefits    []string               // Expected benefits
	Steps       []string               // Implementation steps
	Resources   []string               // Additional resources
	Effort      EstimatedEffort        // Estimated effort
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
	// Layer rules
	Layers []Layer     `json:"layers" yaml:"layers"`
	Rules  []LayerRule `json:"rules" yaml:"rules"`

	// Package rules
	PackageRules []PackageRule `json:"package_rules" yaml:"package_rules"`

	// Custom rules
	CustomRules []CustomRule `json:"custom_rules" yaml:"custom_rules"`

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

	// LoadDefaultConfig loads the default configuration
	LoadDefaultConfig() *SystemAnalysisRequest

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
		OutputFormat:         OutputFormatText,
		AnalyzeDependencies:  BoolPtr(true),
		AnalyzeArchitecture:  BoolPtr(true),
		Recursive:            BoolPtr(true),
		IncludeStdLib:        BoolPtr(false),
		IncludeThirdParty:    BoolPtr(true),
		FollowRelative:       BoolPtr(true),
		DetectCycles:         BoolPtr(true),
		ValidateArchitecture: BoolPtr(true),
		IncludePatterns:      []string{"**/*.py"},
		ExcludePatterns:      []string{"test_*.py", "*_test.py"},
		ComplexityData:       make(map[string]int),
		ClonesData:           make(map[string]float64),
		DeadCodeData:         make(map[string]int),
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
