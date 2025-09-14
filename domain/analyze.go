package domain

import (
	"math"
	"time"
)

// AnalyzeResponse represents the combined results of all analyses
type AnalyzeResponse struct {
    // Analysis results
    Complexity *ComplexityResponse `json:"complexity,omitempty" yaml:"complexity,omitempty"`
    DeadCode   *DeadCodeResponse   `json:"dead_code,omitempty" yaml:"dead_code,omitempty"`
    Clone      *CloneResponse      `json:"clone,omitempty" yaml:"clone,omitempty"`
    CBO        *CBOResponse        `json:"cbo,omitempty" yaml:"cbo,omitempty"`
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
    DepsEnabled bool    `json:"deps_enabled" yaml:"deps_enabled"`
    ArchEnabled bool    `json:"arch_enabled" yaml:"arch_enabled"`
    DepsTotalModules        int     `json:"deps_total_modules" yaml:"deps_total_modules"`
    DepsModulesInCycles     int     `json:"deps_modules_in_cycles" yaml:"deps_modules_in_cycles"`
    DepsMaxDepth            int     `json:"deps_max_depth" yaml:"deps_max_depth"`
    DepsMainSequenceDeviation float64 `json:"deps_main_sequence_deviation" yaml:"deps_main_sequence_deviation"`
    ArchCompliance          float64 `json:"arch_compliance" yaml:"arch_compliance"`

	// Key metrics
	TotalFunctions      int     `json:"total_functions" yaml:"total_functions"`
	AverageComplexity   float64 `json:"average_complexity" yaml:"average_complexity"`
	HighComplexityCount int     `json:"high_complexity_count" yaml:"high_complexity_count"`

	DeadCodeCount    int `json:"dead_code_count" yaml:"dead_code_count"`
	CriticalDeadCode int `json:"critical_dead_code" yaml:"critical_dead_code"`

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

// CalculateHealthScore calculates an overall health score based on analysis results
func (s *AnalyzeSummary) CalculateHealthScore() {
    score := 100

    // Project size normalization (affects dead code penalties)
    normalizationFactor := 1.0
    if s.TotalFiles > 10 {
        normalizationFactor = 1.0 + math.Log10(float64(s.TotalFiles)/10.0)
    }

    // Complexity penalty (max 20)
    switch {
    case s.AverageComplexity > 20:
        score -= 20
    case s.AverageComplexity > 10:
        score -= 12
    case s.AverageComplexity > 5:
        score -= 6
    }

    // Dead code penalty (max 20, normalized)
    if s.DeadCodeCount > 0 || s.CriticalDeadCode > 0 {
        base := int(math.Min(20, float64(s.DeadCodeCount)/normalizationFactor))
        critical := int(math.Min(10, float64(3*s.CriticalDeadCode)/normalizationFactor))
        penalty := base + critical
        if penalty > 20 {
            penalty = 20
        }
        score -= penalty
    }

    // Clone penalty (max 20)
    switch {
    case s.CodeDuplication > 40:
        score -= 20
    case s.CodeDuplication > 25:
        score -= 12
    case s.CodeDuplication > 10:
        score -= 6
    }

    // CBO penalty (max 20) based on ratio of high-coupling classes
    if s.CBOClasses > 0 {
        ratio := float64(s.HighCouplingClasses) / float64(s.CBOClasses)
        switch {
        case ratio > 0.5:
            score -= 16
        case ratio > 0.3:
            score -= 10
        case ratio > 0.1:
            score -= 5
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
            score -= int(math.Round(8 * ratio))
        }

        // Depth penalty (max 2): excess over expected depth ~ O(log N)
        if s.DepsTotalModules > 0 {
            expected := int(math.Max(3, math.Ceil(math.Log2(float64(s.DepsTotalModules)+1))+1))
            excess := s.DepsMaxDepth - expected
            if excess < 0 {
                excess = 0
            }
            if excess > 2 {
                excess = 2
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
            score -= int(math.Round(msd * 2))
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
        score -= int(math.Round(8 * (1 - comp)))
    }

    // Minimum score floor
    if score < 10 {
        score = 10
    }
    s.HealthScore = score

    // Grade mapping
    switch {
    case score >= 85:
        s.Grade = "A"
    case score >= 70:
        s.Grade = "B"
    case score >= 55:
        s.Grade = "C"
    case score >= 40:
        s.Grade = "D"
    default:
        s.Grade = "F"
    }
}

// IsHealthy returns true if the codebase is considered healthy
func (s *AnalyzeSummary) IsHealthy() bool {
	return s.HealthScore >= 70
}

// HasIssues returns true if any issues were found
func (s *AnalyzeSummary) HasIssues() bool {
	return s.HighComplexityCount > 0 || s.DeadCodeCount > 0 || s.ClonePairs > 0 || s.HighCouplingClasses > 0
}
