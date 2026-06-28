package domain

import (
	"fmt"
	"io"
	"math"
	"time"
)

// AnalyzeOutputFormatter defines the interface for formatting unified analysis results
type AnalyzeOutputFormatter interface {
	Write(response *AnalyzeResponse, format OutputFormat, writer io.Writer) error
}

// AnalyzeExecutionConfig contains the resolved configuration that AnalyzeUseCase
// needs after config file discovery and loading.
type AnalyzeExecutionConfig struct {
	ConfigPath string

	IncludePatterns []string
	ExcludePatterns []string
	Recursive       bool

	ComplexityEnabled            bool
	ComplexityReportUnchanged    bool
	ComplexityMinComplexity      int
	ComplexityLowThreshold       int
	ComplexityMediumThreshold    int
	ComplexityMaxComplexity      int
	CognitiveComplexityThreshold int
	NestingDepthThreshold        int

	DeadCodeEnabled bool

	CloneLSHEnabled       string
	CloneLSHAutoThreshold int

	SystemEnabled             bool
	SystemAnalyzeDependencies bool
	SystemAnalyzeArchitecture bool

	CommunitiesEnabled bool
}

// AnalyzeConfigurationLoader resolves and loads configuration for AnalyzeUseCase.
type AnalyzeConfigurationLoader interface {
	LoadAnalyzeExecutionConfig(configPath string, targetPath string) (AnalyzeExecutionConfig, error)
}

// Health Score Calculation Constants
const (
	// Complexity thresholds and penalties
	ComplexityThresholdHigh   = 20
	ComplexityThresholdMedium = 10
	ComplexityThresholdLow    = 5
	ComplexityPenaltyHigh     = 20
	ComplexityPenaltyMedium   = 12
	ComplexityPenaltyLow      = 6

	// Code duplication thresholds and penalties
	// 0% = perfect, 30% = max penalty (using fragment ratio: clonedFragments/totalFragments)
	DuplicationThresholdHigh   = 30.0
	DuplicationThresholdMedium = 15.0
	DuplicationThresholdLow    = 0.0
	DuplicationPenaltyHigh     = 20
	DuplicationPenaltyMedium   = 12
	DuplicationPenaltyLow      = 6

	// CBO coupling scoring curve (used by calculateCouplingPenalty)
	// Penalty grows linearly with the weighted ratio of problematic classes
	// and saturates (reaches the max penalty) at CouplingSaturationRatio.
	CouplingMediumWeight    = 0.3  // Medium-risk classes count 0.3 vs High = 1.0
	CouplingSaturationRatio = 0.40 // weighted ratio at which the penalty maxes out

	// LCOM cohesion scoring curve (used by calculateCohesionPenalty)
	// Penalty grows linearly with the weighted ratio of low-cohesion classes
	// and saturates (reaches the max penalty) at CohesionSaturationRatio.
	CohesionMediumWeight    = 0.3  // Medium-risk classes count 0.3 vs High = 1.0
	CohesionSaturationRatio = 0.40 // weighted ratio at which the penalty maxes out

	// Maximum penalties
	MaxDeadCodePenalty = 20
	MaxCriticalPenalty = 10
	MaxCyclesPenalty   = 10 // Increased from 8 for stricter scoring
	MaxDepthPenalty    = 3  // Increased from 2 for stricter scoring
	MaxArchPenalty     = 12 // Increased from 8 for stricter scoring
	MaxMSDPenalty      = 3  // Increased from 2 for stricter scoring

	// Community detection scoring (opt-in; only applied when communities ran
	// with at least two detected communities). The risk score is a weighted
	// blend of the factors below; the health-score penalty is bounded at
	// MaxCommunityPenalty so disabling communities cannot move existing grades.
	MaxCommunityPenalty          = 10   // bounded contribution to the overall health score
	CommunityModularityTarget    = 0.30 // Q at or above which modularity risk is zero
	CommunityCrossEdgeSaturation = 0.50 // cross-community edge ratio at which that risk maxes out
	// Risk-factor weights (core factors sum to 1.0; optional factors are added
	// and the blend is renormalised over whatever factors are available).
	CommunityModularityWeight = 0.40
	CommunityCrossEdgeWeight  = 0.30
	CommunityBridgeWeight     = 0.30
	CommunityPackageWeight    = 0.25
	CommunityLayerWeight      = 0.25
	// Per-community risk_level thresholds, expressed as risk ratios (0..1).
	CommunityRiskHighRatio   = 0.60 // >= high
	CommunityRiskMediumRatio = 0.30 // >= medium, otherwise low

	// Score display scale - all categories normalized to this base
	MaxScoreBase = 20

	// Actual maximum penalty values for normalization
	MaxDependencyPenalty   = MaxCyclesPenalty + MaxDepthPenalty + MaxMSDPenalty // 16
	MaxArchitecturePenalty = MaxArchPenalty                                     // 12

	// Grade thresholds (stricter than before)
	GradeAThreshold = 90 // Increased from 85
	GradeBThreshold = 75 // Increased from 70
	GradeCThreshold = 60 // Increased from 55
	GradeDThreshold = 45 // Increased from 40

	// Score quality thresholds (aligned with grade thresholds)
	ScoreThresholdExcellent = 90 // Excellent: 90-100 (increased from 85)
	ScoreThresholdGood      = 75 // Good: 75-89 (increased from 70)
	ScoreThresholdFair      = 60 // Fair: 60-74 (increased from 55)
	// Poor: 0-59 (below ScoreThresholdFair)

	// Other constants
	MinimumScore                = 0 // Changed from 10 to allow truly low scores for severely problematic code
	HealthyThreshold            = 70
	FallbackComplexityThreshold = 10
	FallbackPenalty             = 5
)

// AnalyzeResponse represents the combined results of all analyses
type AnalyzeResponse struct {
	// Analysis results
	Complexity  *ComplexityResponse      `json:"complexity,omitempty" yaml:"complexity,omitempty"`
	DeadCode    *DeadCodeResponse        `json:"dead_code,omitempty" yaml:"dead_code,omitempty"`
	Clone       *CloneResponse           `json:"clone,omitempty" yaml:"clone,omitempty"`
	CBO         *CBOResponse             `json:"cbo,omitempty" yaml:"cbo,omitempty"`
	LCOM        *LCOMResponse            `json:"lcom,omitempty" yaml:"lcom,omitempty"`
	System      *SystemAnalysisResponse  `json:"system,omitempty" yaml:"system,omitempty"`
	Communities *CommunityAnalysisResult `json:"community_analysis,omitempty" yaml:"community_analysis,omitempty"`
	MockData    *MockDataResponse        `json:"mock_data,omitempty" yaml:"mock_data,omitempty"`

	// Actionable suggestions derived from analysis results
	Suggestions []Suggestion `json:"suggestions,omitempty" yaml:"suggestions,omitempty"`

	// Overall summary
	Summary AnalyzeSummary `json:"summary" yaml:"summary"`

	// Metadata
	GeneratedAt time.Time `json:"generated_at" yaml:"generated_at"`
	Duration    int64     `json:"duration_ms" yaml:"duration_ms"`
	Version     string    `json:"version" yaml:"version"`
}

// AnalyzeSummary provides an overall summary of all analyses
type AnalyzeSummary struct {
	// File statistics
	TotalFiles    int `json:"total_files" yaml:"total_files"`
	AnalyzedFiles int `json:"analyzed_files" yaml:"analyzed_files"`
	SkippedFiles  int `json:"skipped_files" yaml:"skipped_files"`

	// Analysis status
	ComplexityEnabled bool `json:"complexity_enabled" yaml:"complexity_enabled"`
	DeadCodeEnabled   bool `json:"dead_code_enabled" yaml:"dead_code_enabled"`
	CloneEnabled      bool `json:"clone_enabled" yaml:"clone_enabled"`
	CBOEnabled        bool `json:"cbo_enabled" yaml:"cbo_enabled"`
	MockDataEnabled   bool `json:"mock_data_enabled" yaml:"mock_data_enabled"`

	// System-level (module dependencies & architecture) summary used for scoring
	DepsEnabled               bool    `json:"deps_enabled" yaml:"deps_enabled"`
	ArchEnabled               bool    `json:"arch_enabled" yaml:"arch_enabled"`
	CommunitiesEnabled        bool    `json:"communities_enabled" yaml:"communities_enabled"`
	DepsTotalModules          int     `json:"deps_total_modules" yaml:"deps_total_modules"`
	DepsModulesInCycles       int     `json:"deps_modules_in_cycles" yaml:"deps_modules_in_cycles"`
	DepsMaxDepth              int     `json:"deps_max_depth" yaml:"deps_max_depth"`
	DepsMainSequenceDeviation float64 `json:"deps_main_sequence_deviation" yaml:"deps_main_sequence_deviation"`
	ArchCompliance            float64 `json:"arch_compliance" yaml:"arch_compliance"`

	// Community detection metrics used for scoring (populated when CommunitiesEnabled).
	CommunityCount            int      `json:"community_count" yaml:"community_count"`
	CommunityModularity       float64  `json:"community_modularity" yaml:"community_modularity"`
	CommunityBridgeModules    int      `json:"community_bridge_modules" yaml:"community_bridge_modules"`
	CommunityInternalEdges    int      `json:"community_internal_edges" yaml:"community_internal_edges"`
	CommunityCrossEdges       int      `json:"community_cross_edges" yaml:"community_cross_edges"`
	CommunityPackageAlignment *float64 `json:"community_package_alignment,omitempty" yaml:"community_package_alignment,omitempty"`
	CommunityLayerAlignment   *float64 `json:"community_layer_alignment,omitempty" yaml:"community_layer_alignment,omitempty"`

	// Key metrics
	TotalFunctions             int     `json:"total_functions" yaml:"total_functions"`
	AverageComplexity          float64 `json:"average_complexity" yaml:"average_complexity"`
	AverageCognitiveComplexity float64 `json:"average_cognitive_complexity" yaml:"average_cognitive_complexity"`
	AverageNestingDepth        float64 `json:"average_nesting_depth" yaml:"average_nesting_depth"`
	HighComplexityCount        int     `json:"high_complexity_count" yaml:"high_complexity_count"`

	DeadCodeCount    int `json:"dead_code_count" yaml:"dead_code_count"`
	CriticalDeadCode int `json:"critical_dead_code" yaml:"critical_dead_code"`
	WarningDeadCode  int `json:"warning_dead_code" yaml:"warning_dead_code"`
	InfoDeadCode     int `json:"info_dead_code" yaml:"info_dead_code"`

	TotalClones     int     `json:"total_clones" yaml:"total_clones"`
	ClonePairs      int     `json:"clone_pairs" yaml:"clone_pairs"`
	CloneGroups     int     `json:"clone_groups" yaml:"clone_groups"`
	CodeDuplication float64 `json:"code_duplication_percentage" yaml:"code_duplication_percentage"`

	CBOClasses            int     `json:"cbo_classes" yaml:"cbo_classes"`
	HighCouplingClasses   int     `json:"high_coupling_classes" yaml:"high_coupling_classes"`     // CBO > 7 (High Risk)
	MediumCouplingClasses int     `json:"medium_coupling_classes" yaml:"medium_coupling_classes"` // 3 < CBO ≤ 7 (Medium Risk)
	AverageCoupling       float64 `json:"average_coupling" yaml:"average_coupling"`

	LCOMEnabled       bool    `json:"lcom_enabled" yaml:"lcom_enabled"`
	LCOMClasses       int     `json:"lcom_classes" yaml:"lcom_classes"`
	HighLCOMClasses   int     `json:"high_lcom_classes" yaml:"high_lcom_classes"`     // LCOM4 > 5 (High Risk)
	MediumLCOMClasses int     `json:"medium_lcom_classes" yaml:"medium_lcom_classes"` // 2 < LCOM4 ≤ 5 (Medium Risk)
	AverageLCOM       float64 `json:"average_lcom" yaml:"average_lcom"`

	MockDataCount        int `json:"mock_data_count" yaml:"mock_data_count"`
	MockDataErrorCount   int `json:"mock_data_error_count" yaml:"mock_data_error_count"`
	MockDataWarningCount int `json:"mock_data_warning_count" yaml:"mock_data_warning_count"`
	MockDataInfoCount    int `json:"mock_data_info_count" yaml:"mock_data_info_count"`

	// Overall health score (0-100)
	HealthScore int    `json:"health_score" yaml:"health_score"`
	Grade       string `json:"grade" yaml:"grade"` // A, B, C, D, F

	// Individual category scores (0-100)
	ComplexityScore   int `json:"complexity_score" yaml:"complexity_score"`
	DeadCodeScore     int `json:"dead_code_score" yaml:"dead_code_score"`
	DuplicationScore  int `json:"duplication_score" yaml:"duplication_score"`
	CouplingScore     int `json:"coupling_score" yaml:"coupling_score"`
	CohesionScore     int `json:"cohesion_score" yaml:"cohesion_score"`
	DependencyScore   int `json:"dependency_score" yaml:"dependency_score"`
	ArchitectureScore int `json:"architecture_score" yaml:"architecture_score"`
	CommunityScore    int `json:"community_score" yaml:"community_score"`

	// CommunityRiskScore is a system-level 0-100 risk signal (higher = worse).
	// It is the inverse of CommunityScore and only meaningful when communities ran.
	CommunityRiskScore int `json:"community_risk_score" yaml:"community_risk_score"`
}

// Validate checks if the summary contains valid values
func (s *AnalyzeSummary) Validate() error {
	// Basic range checks
	if s.AverageComplexity < 0 {
		return fmt.Errorf("AverageComplexity cannot be negative: %f", s.AverageComplexity)
	}

	if s.AverageCognitiveComplexity < 0 {
		return fmt.Errorf("AverageCognitiveComplexity cannot be negative: %f", s.AverageCognitiveComplexity)
	}

	if s.AverageNestingDepth < 0 {
		return fmt.Errorf("AverageNestingDepth cannot be negative: %f", s.AverageNestingDepth)
	}

	if s.CodeDuplication < 0 || s.CodeDuplication > 100 {
		return fmt.Errorf("CodeDuplication must be 0-100: %f", s.CodeDuplication)
	}

	// Architecture compliance check (when enabled)
	if s.ArchEnabled {
		if s.ArchCompliance < 0 || s.ArchCompliance > 1 {
			return fmt.Errorf("ArchCompliance must be 0-1, got %f", s.ArchCompliance)
		}
	}

	// Dependency metrics check (when enabled)
	if s.DepsEnabled {
		if s.DepsMainSequenceDeviation < 0 || s.DepsMainSequenceDeviation > 1 {
			return fmt.Errorf("DepsMainSequenceDeviation must be 0-1, got %f", s.DepsMainSequenceDeviation)
		}

		if s.DepsTotalModules > 0 && s.DepsModulesInCycles > s.DepsTotalModules {
			return fmt.Errorf("DepsModulesInCycles (%d) cannot exceed DepsTotalModules (%d)",
				s.DepsModulesInCycles, s.DepsTotalModules)
		}
	}

	// Community metric checks (when enabled)
	if s.CommunitiesEnabled {
		if s.CommunityModularity < -1 || s.CommunityModularity > 1 {
			return fmt.Errorf("CommunityModularity must be -1..1, got %f", s.CommunityModularity)
		}
		if s.CommunityPackageAlignment != nil && (*s.CommunityPackageAlignment < 0 || *s.CommunityPackageAlignment > 1) {
			return fmt.Errorf("CommunityPackageAlignment must be 0-1, got %f", *s.CommunityPackageAlignment)
		}
		if s.CommunityLayerAlignment != nil && (*s.CommunityLayerAlignment < 0 || *s.CommunityLayerAlignment > 1) {
			return fmt.Errorf("CommunityLayerAlignment must be 0-1, got %f", *s.CommunityLayerAlignment)
		}
	}

	// LCOM checks
	if s.LCOMClasses > 0 {
		if s.HighLCOMClasses > s.LCOMClasses {
			return fmt.Errorf("HighLCOMClasses (%d) cannot exceed LCOMClasses (%d)",
				s.HighLCOMClasses, s.LCOMClasses)
		}
		if s.MediumLCOMClasses > s.LCOMClasses {
			return fmt.Errorf("MediumLCOMClasses (%d) cannot exceed LCOMClasses (%d)",
				s.MediumLCOMClasses, s.LCOMClasses)
		}
		if (s.HighLCOMClasses + s.MediumLCOMClasses) > s.LCOMClasses {
			return fmt.Errorf("HighLCOMClasses + MediumLCOMClasses (%d) cannot exceed LCOMClasses (%d)",
				s.HighLCOMClasses+s.MediumLCOMClasses, s.LCOMClasses)
		}
	}

	// CBO checks
	if s.CBOClasses > 0 {
		if s.HighCouplingClasses > s.CBOClasses {
			return fmt.Errorf("HighCouplingClasses (%d) cannot exceed CBOClasses (%d)",
				s.HighCouplingClasses, s.CBOClasses)
		}
		if s.MediumCouplingClasses > s.CBOClasses {
			return fmt.Errorf("MediumCouplingClasses (%d) cannot exceed CBOClasses (%d)",
				s.MediumCouplingClasses, s.CBOClasses)
		}
		if (s.HighCouplingClasses + s.MediumCouplingClasses) > s.CBOClasses {
			return fmt.Errorf("HighCouplingClasses + MediumCouplingClasses (%d) cannot exceed CBOClasses (%d)",
				s.HighCouplingClasses+s.MediumCouplingClasses, s.CBOClasses)
		}
	}

	return nil
}

// calculateComplexityPenalty calculates the penalty for complexity (max 20)
// Uses continuous linear function starting from avg complexity of 2
func (s *AnalyzeSummary) calculateComplexityPenalty() int {
	// Linear penalty: starts at avg=2, reaches max (20) at avg=15
	// Formula: penalty = (avg - 2) / 13 * 20
	mccabePenalty := calculateLinearPenalty(s.AverageComplexity, 2.0, 15.0)
	cognitivePenalty := calculateLinearPenalty(s.AverageCognitiveComplexity, 15.0, float64(DefaultCognitiveComplexityThreshold))
	nestingPenalty := calculateLinearPenalty(s.AverageNestingDepth, 3.0, float64(DefaultNestingDepthThreshold))

	return maxInt(mccabePenalty, cognitivePenalty, nestingPenalty)
}

func calculateLinearPenalty(value, start, saturation float64) int {
	if value <= start {
		return 0
	}
	if saturation <= start {
		return MaxScoreBase
	}

	penalty := (value - start) / (saturation - start) * float64(MaxScoreBase)
	if penalty > float64(MaxScoreBase) {
		penalty = float64(MaxScoreBase)
	}

	return int(math.Round(penalty))
}

func maxInt(values ...int) int {
	max := 0
	for _, value := range values {
		if value > max {
			max = value
		}
	}
	return max
}

// calculateDeadCodePenalty calculates the penalty for dead code (max 20)
// Uses weighted counting: Critical=1.0, Warning=0.5, Info=0.2
func (s *AnalyzeSummary) calculateDeadCodePenalty(normalizationFactor float64) int {
	// Calculate weighted dead code count
	// Critical issues have full weight, warning half, info minimal
	weightedDeadCode := float64(s.CriticalDeadCode)*1.0 +
		float64(s.WarningDeadCode)*0.5 +
		float64(s.InfoDeadCode)*0.2

	if weightedDeadCode <= 0 {
		return 0
	}

	penalty := int(math.Min(float64(MaxDeadCodePenalty), weightedDeadCode/normalizationFactor))
	return penalty
}

// calculateDuplicationPenalty calculates the penalty for code duplication (max 20)
// Uses continuous linear function based on defined thresholds
func (s *AnalyzeSummary) calculateDuplicationPenalty() int {
	// Linear penalty: 0% = 0 penalty, 30% = max penalty (20)
	if s.CodeDuplication <= DuplicationThresholdLow {
		return 0
	}

	// Formula: penalty = (duplication - low) / (high - low) * 20
	penaltyRange := DuplicationThresholdHigh - DuplicationThresholdLow // 30%
	penalty := (s.CodeDuplication - DuplicationThresholdLow) / penaltyRange * 20.0
	if penalty > 20.0 {
		penalty = 20.0
	}

	return int(math.Round(penalty))
}

// calculateCouplingPenalty calculates the penalty for class coupling (max 20)
// Uses continuous linear function based on weighted ratio of problematic classes
func (s *AnalyzeSummary) calculateCouplingPenalty() int {
	if s.CBOClasses == 0 {
		return 0
	}

	// Calculate combined problematic classes ratio
	// Weight: High Risk = 1.0, Medium Risk = CouplingMediumWeight
	weightedProblematicClasses := float64(s.HighCouplingClasses) + (CouplingMediumWeight * float64(s.MediumCouplingClasses))
	ratio := weightedProblematicClasses / float64(s.CBOClasses)

	// Linear penalty: starts at 0%, reaches max (20) at CouplingSaturationRatio
	// Formula: penalty = ratio / CouplingSaturationRatio * 20
	penalty := ratio / CouplingSaturationRatio * 20.0
	if penalty > 20.0 {
		penalty = 20.0
	}

	return int(math.Round(penalty))
}

// calculateCohesionPenalty calculates the penalty for class cohesion (max 20)
// Uses continuous linear function based on weighted ratio of low-cohesion classes
func (s *AnalyzeSummary) calculateCohesionPenalty() int {
	if s.LCOMClasses == 0 {
		return 0
	}

	// Calculate combined problematic classes ratio
	// Weight: High Risk (LCOM4 > 5) = 1.0, Medium Risk (LCOM4 3-5) = CohesionMediumWeight
	weightedProblematicClasses := float64(s.HighLCOMClasses) + (CohesionMediumWeight * float64(s.MediumLCOMClasses))
	ratio := weightedProblematicClasses / float64(s.LCOMClasses)

	// Linear penalty: starts at 0%, reaches max (20) at CohesionSaturationRatio
	// Formula: penalty = ratio / CohesionSaturationRatio * 20
	penalty := ratio / CohesionSaturationRatio * 20.0
	if penalty > 20.0 {
		penalty = 20.0
	}

	return int(math.Round(penalty))
}

// calculateDependencyPenalty calculates the penalty for module dependencies (max 16: cycles=10, depth=3, MSD=3)
func (s *AnalyzeSummary) calculateDependencyPenalty() int {
	if !s.DepsEnabled {
		return 0
	}

	penalty := 0

	// Cycles penalty (max 10): proportion of modules in cycles
	if s.DepsTotalModules > 0 {
		ratio := float64(s.DepsModulesInCycles) / float64(s.DepsTotalModules)
		if ratio < 0 {
			ratio = 0
		}
		if ratio > 1 {
			ratio = 1
		}
		penalty += int(math.Round(float64(MaxCyclesPenalty) * ratio))
	}

	// Depth penalty (max 3): excess over expected depth ~ O(log N)
	if s.DepsTotalModules > 0 {
		expected := int(math.Max(3, math.Ceil(math.Log2(float64(s.DepsTotalModules)+1))+1))
		excess := s.DepsMaxDepth - expected
		if excess < 0 {
			excess = 0
		}
		if excess > MaxDepthPenalty {
			excess = MaxDepthPenalty
		}
		penalty += excess
	}

	// Main sequence deviation penalty (max 3)
	if s.DepsMainSequenceDeviation > 0 {
		msd := s.DepsMainSequenceDeviation
		if msd < 0 {
			msd = 0
		}
		if msd > 1 {
			msd = 1
		}
		penalty += int(math.Round(msd * float64(MaxMSDPenalty)))
	}

	return penalty
}

// calculateArchitecturePenalty calculates the penalty for architecture compliance (max 12)
func (s *AnalyzeSummary) calculateArchitecturePenalty() int {
	if !s.ArchEnabled {
		return 0
	}

	comp := s.ArchCompliance
	if comp < 0 {
		comp = 0
	}
	if comp > 1 {
		comp = 1
	}
	return int(math.Round(float64(MaxArchPenalty) * (1 - comp)))
}

// clamp01 bounds a value to the [0, 1] interval.
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// communityRiskInputs collects the raw signals used to derive the community
// risk score. It is shared by the AnalyzeSummary path (health scoring) and the
// CommunityAnalysisResult path (standalone community command) so both produce
// identical numbers.
type communityRiskInputs struct {
	communityCount   int
	modularity       float64
	bridgeModules    int
	internalEdges    int
	crossEdges       int
	packageAlignment *float64 // nil when package metadata is unavailable
	layerAlignment   *float64 // nil when architecture layers are not configured
}

// computeCommunityRiskRatio blends the community risk factors into a single
// 0..1 ratio (0 = healthy, 1 = worst). Optional factors (package/layer
// alignment) are only included when available, and the weighted average is
// renormalised over whichever factors contributed. See docs/ANALYZE_SCORING.md.
func computeCommunityRiskRatio(in communityRiskInputs) float64 {
	var weightedSum, totalWeight float64

	// Low modularity Q: risk rises as Q falls below the "good separation" target.
	modularityRisk := clamp01((CommunityModularityTarget - in.modularity) / CommunityModularityTarget)
	weightedSum += CommunityModularityWeight * modularityRisk
	totalWeight += CommunityModularityWeight

	// Cross-community edge ratio: how tangled the partitions are. This also
	// captures the aggregate external_dependency_ratio at the system level.
	if denom := in.internalEdges + in.crossEdges; denom > 0 {
		crossRatio := float64(in.crossEdges) / float64(denom)
		crossRisk := clamp01(crossRatio / CommunityCrossEdgeSaturation)
		weightedSum += CommunityCrossEdgeWeight * crossRisk
		totalWeight += CommunityCrossEdgeWeight
	}

	// Bridge modules: count relative to the number of communities, saturating at
	// roughly one bridge module per community.
	if in.communityCount > 0 {
		bridgeRisk := clamp01(float64(in.bridgeModules) / float64(in.communityCount))
		weightedSum += CommunityBridgeWeight * bridgeRisk
		totalWeight += CommunityBridgeWeight
	}

	// Low package alignment (when available).
	if in.packageAlignment != nil {
		weightedSum += CommunityPackageWeight * clamp01(1-*in.packageAlignment)
		totalWeight += CommunityPackageWeight
	}

	// Low layer alignment (when available).
	if in.layerAlignment != nil {
		weightedSum += CommunityLayerWeight * clamp01(1-*in.layerAlignment)
		totalWeight += CommunityLayerWeight
	}

	if totalWeight == 0 {
		return 0
	}
	return weightedSum / totalWeight
}

// communityRiskRatio returns the system community risk ratio (0..1) for the
// summary and whether community scoring applies. Scoring is skipped unless
// communities ran and at least two communities were detected (a single
// community has no meaningful modular structure to score).
func (s *AnalyzeSummary) communityRiskRatio() (float64, bool) {
	if !s.CommunitiesEnabled || s.CommunityCount < 2 {
		return 0, false
	}
	return computeCommunityRiskRatio(communityRiskInputs{
		communityCount:   s.CommunityCount,
		modularity:       s.CommunityModularity,
		bridgeModules:    s.CommunityBridgeModules,
		internalEdges:    s.CommunityInternalEdges,
		crossEdges:       s.CommunityCrossEdges,
		packageAlignment: s.CommunityPackageAlignment,
		layerAlignment:   s.CommunityLayerAlignment,
	}), true
}

// normalizeToScoreBase normalizes a penalty value to the MaxScoreBase scale (0-20)
// This ensures all category scores use a consistent display scale
func normalizeToScoreBase(penalty int, maxPenalty int) int {
	if maxPenalty == 0 {
		return 0
	}
	normalized := int(math.Round(float64(penalty) / float64(maxPenalty) * float64(MaxScoreBase)))
	if normalized < 0 {
		normalized = 0
	}
	if normalized > MaxScoreBase {
		normalized = MaxScoreBase
	}
	return normalized
}

// penaltyToScore converts a penalty value to a 0-100 score
func penaltyToScore(penalty int, maxPenalty int) int {
	if maxPenalty == 0 {
		return 100
	}
	score := 100 - int(math.Round(float64(penalty)*100.0/float64(maxPenalty)))
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score
}

// CalculateHealthScore calculates an overall health score based on analysis results
func (s *AnalyzeSummary) CalculateHealthScore() error {
	// Validate input values first
	if err := s.Validate(); err != nil {
		// Set default values on error
		s.HealthScore = 0
		s.Grade = "N/A"
		s.ComplexityScore = 0
		s.DeadCodeScore = 0
		s.DuplicationScore = 0
		s.CouplingScore = 0
		s.CohesionScore = 0
		s.DependencyScore = 0
		s.ArchitectureScore = 0
		s.CommunityScore = 0
		s.CommunityRiskScore = 0
		return fmt.Errorf("invalid summary data: %w", err)
	}
	score := 100

	// Project size normalization (affects dead code penalties)
	normalizationFactor := 1.0
	if s.TotalFiles > 10 {
		normalizationFactor = 1.0 + math.Log10(float64(s.TotalFiles)/10.0)
	}

	// Calculate penalties and corresponding scores
	// Individual scores are normalized to a consistent 20-point scale for display consistency

	complexityPenalty := s.calculateComplexityPenalty()
	s.ComplexityScore = penaltyToScore(complexityPenalty, MaxScoreBase)
	score -= complexityPenalty

	deadCodePenalty := s.calculateDeadCodePenalty(normalizationFactor)
	s.DeadCodeScore = penaltyToScore(deadCodePenalty, MaxScoreBase)
	score -= deadCodePenalty

	duplicationPenalty := s.calculateDuplicationPenalty()
	s.DuplicationScore = penaltyToScore(duplicationPenalty, MaxScoreBase)
	score -= duplicationPenalty

	couplingPenalty := s.calculateCouplingPenalty()
	s.CouplingScore = penaltyToScore(couplingPenalty, MaxScoreBase)
	score -= couplingPenalty

	cohesionPenalty := s.calculateCohesionPenalty()
	s.CohesionScore = penaltyToScore(cohesionPenalty, MaxScoreBase)
	score -= cohesionPenalty

	// Dependencies and Architecture need normalization since their max penalties differ from MaxScoreBase
	dependencyPenalty := s.calculateDependencyPenalty()
	normalizedDepPenalty := normalizeToScoreBase(dependencyPenalty, MaxDependencyPenalty)
	s.DependencyScore = penaltyToScore(normalizedDepPenalty, MaxScoreBase)
	score -= dependencyPenalty

	architecturePenalty := s.calculateArchitecturePenalty()
	// Use compliance directly as score (98% compliance = 98 points)
	s.ArchitectureScore = int(math.Round(s.ArchCompliance * 100))
	score -= architecturePenalty

	// Community detection: only penalises when communities ran with >= 2
	// communities. Disabled or trivial cases score 100 / risk 0 so existing
	// grades are unaffected (backward compatible).
	if communityRatio, scored := s.communityRiskRatio(); scored {
		s.CommunityRiskScore = int(math.Round(communityRatio * 100))
		s.CommunityScore = 100 - s.CommunityRiskScore
		score -= int(math.Round(communityRatio * float64(MaxCommunityPenalty)))
	} else {
		s.CommunityScore = 100
		s.CommunityRiskScore = 0
	}

	// Minimum score floor
	if score < MinimumScore {
		score = MinimumScore
	}
	s.HealthScore = score

	// Grade mapping
	switch {
	case score >= GradeAThreshold:
		s.Grade = "A"
	case score >= GradeBThreshold:
		s.Grade = "B"
	case score >= GradeCThreshold:
		s.Grade = "C"
	case score >= GradeDThreshold:
		s.Grade = "D"
	default:
		s.Grade = "F"
	}

	return nil
}

// CalculateFallbackScore provides a simple fallback health score calculation
// Used when validation fails to provide a basic score based on available metrics
func (s *AnalyzeSummary) CalculateFallbackScore() int {
	score := 100

	// Complexity penalty
	if s.AverageComplexity > float64(FallbackComplexityThreshold) {
		score -= FallbackComplexityThreshold
	}

	// Dead code penalty
	if s.DeadCodeCount > 0 {
		score -= FallbackPenalty
	}

	// High complexity penalty
	if s.HighComplexityCount > 0 {
		score -= FallbackPenalty
	}

	// Low cohesion penalty
	if s.HighLCOMClasses > 0 {
		score -= FallbackPenalty
	}

	if score < MinimumScore {
		score = MinimumScore
	}

	return score
}

// GetGradeFromScore maps a health score to a letter grade
func GetGradeFromScore(score int) string {
	switch {
	case score >= GradeAThreshold:
		return "A"
	case score >= GradeBThreshold:
		return "B"
	case score >= GradeCThreshold:
		return "C"
	case score >= GradeDThreshold:
		return "D"
	default:
		return "F"
	}
}

// IsHealthy returns true if the codebase is considered healthy
func (s *AnalyzeSummary) IsHealthy() bool {
	return s.HealthScore >= HealthyThreshold
}

// HasIssues returns true if any issues were found
func (s *AnalyzeSummary) HasIssues() bool {
	return s.HighComplexityCount > 0 || s.DeadCodeCount > 0 || s.ClonePairs > 0 || s.HighCouplingClasses > 0
}
