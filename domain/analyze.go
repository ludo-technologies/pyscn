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
	DuplicationThresholdHigh   = 40.0
	DuplicationThresholdMedium = 25.0
	DuplicationThresholdLow    = 10.0
	DuplicationPenaltyHigh     = 20
	DuplicationPenaltyMedium   = 12
	DuplicationPenaltyLow      = 6

	// CBO coupling thresholds and penalties
	CouplingRatioHigh     = 0.5
	CouplingRatioMedium   = 0.3
	CouplingRatioLow      = 0.1
	CouplingPenaltyHigh   = 16
	CouplingPenaltyMedium = 10
	CouplingPenaltyLow    = 5

	// Maximum penalties
	MaxDeadCodePenalty = 20
	MaxCriticalPenalty = 10
	MaxCyclesPenalty   = 8
	MaxDepthPenalty    = 2
	MaxArchPenalty     = 8
	MaxMSDPenalty      = 2

	// Grade thresholds
	GradeAThreshold = 85
	GradeBThreshold = 70
	GradeCThreshold = 55
	GradeDThreshold = 40

	// Other constants
	MinimumScore                = 10
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

	TotalClones     int     `json:"total_clones" yaml:"total_clones"`
	ClonePairs      int     `json:"clone_pairs" yaml:"clone_pairs"`
	CloneGroups     int     `json:"clone_groups" yaml:"clone_groups"`
	CodeDuplication float64 `json:"code_duplication_percentage" yaml:"code_duplication_percentage"`

	CBOClasses          int     `json:"cbo_classes" yaml:"cbo_classes"`
	HighCouplingClasses int     `json:"high_coupling_classes" yaml:"high_coupling_classes"`
	AverageCoupling     float64 `json:"average_coupling" yaml:"average_coupling"`

	// Overall health score (0-100)
	HealthScore int    `json:"health_score" yaml:"health_score"`
	Grade       string `json:"grade" yaml:"grade"` // A, B, C, D, F
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
	if s.CBOClasses > 0 && s.HighCouplingClasses > s.CBOClasses {
		return fmt.Errorf("HighCouplingClasses (%d) cannot exceed CBOClasses (%d)",
			s.HighCouplingClasses, s.CBOClasses)
	}

	return nil
}

// CalculateHealthScore calculates an overall health score based on analysis results
func (s *AnalyzeSummary) CalculateHealthScore() error {
	// Validate input values first
	if err := s.Validate(); err != nil {
		// Set default values on error
		s.HealthScore = 0
		s.Grade = "N/A"
		return fmt.Errorf("invalid summary data: %w", err)
	}
	score := 100

	// Project size normalization (affects dead code penalties)
	normalizationFactor := 1.0
	if s.TotalFiles > 10 {
		normalizationFactor = 1.0 + math.Log10(float64(s.TotalFiles)/10.0)
	}

	// Complexity penalty
	switch {
	case s.AverageComplexity > float64(ComplexityThresholdHigh):
		score -= ComplexityPenaltyHigh
	case s.AverageComplexity > float64(ComplexityThresholdMedium):
		score -= ComplexityPenaltyMedium
	case s.AverageComplexity > float64(ComplexityThresholdLow):
		score -= ComplexityPenaltyLow
	}

	// Dead code penalty (normalized)
	if s.DeadCodeCount > 0 || s.CriticalDeadCode > 0 {
		base := int(math.Min(float64(MaxDeadCodePenalty), float64(s.DeadCodeCount)/normalizationFactor))
		critical := int(math.Min(float64(MaxCriticalPenalty), float64(3*s.CriticalDeadCode)/normalizationFactor))
		penalty := base + critical
		if penalty > MaxDeadCodePenalty {
			penalty = MaxDeadCodePenalty
		}
		score -= penalty
	}

	// Clone penalty
	switch {
	case s.CodeDuplication > DuplicationThresholdHigh:
		score -= DuplicationPenaltyHigh
	case s.CodeDuplication > DuplicationThresholdMedium:
		score -= DuplicationPenaltyMedium
	case s.CodeDuplication > DuplicationThresholdLow:
		score -= DuplicationPenaltyLow
	}

	// CBO penalty based on ratio of high-coupling classes
	if s.CBOClasses > 0 {
		ratio := float64(s.HighCouplingClasses) / float64(s.CBOClasses)
		switch {
		case ratio > CouplingRatioHigh:
			score -= CouplingPenaltyHigh
		case ratio > CouplingRatioMedium:
			score -= CouplingPenaltyMedium
		case ratio > CouplingRatioLow:
			score -= CouplingPenaltyLow
		}
	}

	// Module Dependencies & Architecture (max 20 total)
	if s.DepsEnabled {
		// Cycles penalty (max 8): proportion of modules in cycles
		if s.DepsTotalModules > 0 {
			ratio := float64(s.DepsModulesInCycles) / float64(s.DepsTotalModules)
			if ratio < 0 {
				ratio = 0
			}
			if ratio > 1 {
				ratio = 1
			}
			score -= int(math.Round(float64(MaxCyclesPenalty) * ratio))
		}

		// Depth penalty (max 2): excess over expected depth ~ O(log N)
		if s.DepsTotalModules > 0 {
			expected := int(math.Max(3, math.Ceil(math.Log2(float64(s.DepsTotalModules)+1))+1))
			excess := s.DepsMaxDepth - expected
			if excess < 0 {
				excess = 0
			}
			if excess > MaxDepthPenalty {
				excess = MaxDepthPenalty
			}
			score -= excess
		}

		// Main sequence deviation penalty (max 2)
		if s.DepsMainSequenceDeviation > 0 {
			msd := s.DepsMainSequenceDeviation
			if msd < 0 {
				msd = 0
			}
			if msd > 1 {
				msd = 1
			}
			score -= int(math.Round(msd * float64(MaxMSDPenalty)))
		}
	}

	// Architecture compliance penalty (max 8)
	if s.ArchEnabled {
		comp := s.ArchCompliance
		if comp < 0 {
			comp = 0
		}
		if comp > 1 {
			comp = 1
		}
		score -= int(math.Round(float64(MaxArchPenalty) * (1 - comp)))
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
