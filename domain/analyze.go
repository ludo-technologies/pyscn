package domain

import (
	"fmt"
	"math"
	"time"
)

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
	DuplicationThresholdHigh   = 20.0
	DuplicationThresholdMedium = 10.0
	DuplicationThresholdLow    = 3.0
	DuplicationPenaltyHigh     = 20
	DuplicationPenaltyMedium   = 12
	DuplicationPenaltyLow      = 6

	// CBO coupling thresholds and penalties
	CouplingRatioHigh     = 0.30 // 30% or more classes with high coupling
	CouplingRatioMedium   = 0.15 // 15-30% classes with high coupling
	CouplingRatioLow      = 0.05 // 5-15% classes with high coupling
	CouplingPenaltyHigh   = 20   // Aligned with other high penalties
	CouplingPenaltyMedium = 12   // Aligned with other medium penalties
	CouplingPenaltyLow    = 6    // Aligned with other low penalties

	// Maximum penalties
	MaxDeadCodePenalty = 20
	MaxCriticalPenalty = 10
	MaxCyclesPenalty   = 10 // Increased from 8 for stricter scoring
	MaxDepthPenalty    = 3  // Increased from 2 for stricter scoring
	MaxArchPenalty     = 12 // Increased from 8 for stricter scoring
	MaxMSDPenalty      = 3  // Increased from 2 for stricter scoring

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
	Complexity *ComplexityResponse     `json:"complexity,omitempty" yaml:"complexity,omitempty"`
	DeadCode   *DeadCodeResponse       `json:"dead_code,omitempty" yaml:"dead_code,omitempty"`
	Clone      *CloneResponse          `json:"clone,omitempty" yaml:"clone,omitempty"`
	CBO        *CBOResponse            `json:"cbo,omitempty" yaml:"cbo,omitempty"`
	System     *SystemAnalysisResponse `json:"system,omitempty" yaml:"system,omitempty"`

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

	// System-level (module dependencies & architecture) summary used for scoring
	DepsEnabled               bool    `json:"deps_enabled" yaml:"deps_enabled"`
	ArchEnabled               bool    `json:"arch_enabled" yaml:"arch_enabled"`
	DepsTotalModules          int     `json:"deps_total_modules" yaml:"deps_total_modules"`
	DepsModulesInCycles       int     `json:"deps_modules_in_cycles" yaml:"deps_modules_in_cycles"`
	DepsMaxDepth              int     `json:"deps_max_depth" yaml:"deps_max_depth"`
	DepsMainSequenceDeviation float64 `json:"deps_main_sequence_deviation" yaml:"deps_main_sequence_deviation"`
	ArchCompliance            float64 `json:"arch_compliance" yaml:"arch_compliance"`

	// Key metrics
	TotalFunctions      int     `json:"total_functions" yaml:"total_functions"`
	AverageComplexity   float64 `json:"average_complexity" yaml:"average_complexity"`
	HighComplexityCount int     `json:"high_complexity_count" yaml:"high_complexity_count"`

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

	// Overall health score (0-100)
	HealthScore int    `json:"health_score" yaml:"health_score"`
	Grade       string `json:"grade" yaml:"grade"` // A, B, C, D, F

	// Individual category scores (0-100)
	ComplexityScore   int `json:"complexity_score" yaml:"complexity_score"`
	DeadCodeScore     int `json:"dead_code_score" yaml:"dead_code_score"`
	DuplicationScore  int `json:"duplication_score" yaml:"duplication_score"`
	CouplingScore     int `json:"coupling_score" yaml:"coupling_score"`
	DependencyScore   int `json:"dependency_score" yaml:"dependency_score"`
	ArchitectureScore int `json:"architecture_score" yaml:"architecture_score"`
}

// Validate checks if the summary contains valid values
func (s *AnalyzeSummary) Validate() error {
	// Basic range checks
	if s.AverageComplexity < 0 {
		return fmt.Errorf("AverageComplexity cannot be negative: %f", s.AverageComplexity)
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
	if s.AverageComplexity <= 2.0 {
		return 0
	}

	penalty := (s.AverageComplexity - 2.0) / 13.0 * 20.0
	if penalty > 20.0 {
		penalty = 20.0
	}

	return int(math.Round(penalty))
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
// Uses continuous linear function starting from 1% duplication
func (s *AnalyzeSummary) calculateDuplicationPenalty() int {
	// Linear penalty: starts at 1%, reaches max (20) at 8%
	// Formula: penalty = (duplication - 1) / 7 * 20
	if s.CodeDuplication <= 1.0 {
		return 0
	}

	penalty := (s.CodeDuplication - 1.0) / 7.0 * 20.0
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
	// Weight: High Risk = 1.0, Medium Risk = 0.5
	weightedProblematicClasses := float64(s.HighCouplingClasses) + (0.5 * float64(s.MediumCouplingClasses))
	ratio := weightedProblematicClasses / float64(s.CBOClasses)

	// Linear penalty: starts at 0%, reaches max (20) at 12%
	// Formula: penalty = ratio / 0.12 * 20
	penalty := ratio / 0.12 * 20.0
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
		s.DependencyScore = 0
		s.ArchitectureScore = 0
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

	// Dependencies and Architecture need normalization since their max penalties differ from MaxScoreBase
	dependencyPenalty := s.calculateDependencyPenalty()
	normalizedDepPenalty := normalizeToScoreBase(dependencyPenalty, MaxDependencyPenalty)
	s.DependencyScore = penaltyToScore(normalizedDepPenalty, MaxScoreBase)
	score -= dependencyPenalty

	architecturePenalty := s.calculateArchitecturePenalty()
	normalizedArchPenalty := normalizeToScoreBase(architecturePenalty, MaxArchitecturePenalty)
	s.ArchitectureScore = penaltyToScore(normalizedArchPenalty, MaxScoreBase)
	score -= architecturePenalty

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
