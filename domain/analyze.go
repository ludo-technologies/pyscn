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

	// Calculate normalization factor for large projects
	normalizationFactor := 1.0
	if s.TotalFiles > 10 {
		// Scale factor increases with project size to reduce penalty impact
		// 20 files: 1.3x, 50 files: 1.7x, 100 files: 2.0x, 200 files: 2.3x
		normalizationFactor = 1.0 + math.Log10(float64(s.TotalFiles)/10.0)
	}

	// Complexity penalty (max 25 points)
	complexityPenalty := 0
	if s.AverageComplexity > 20 {
		complexityPenalty = 25
	} else if s.AverageComplexity > 10 {
		complexityPenalty = 15
	} else if s.AverageComplexity > 5 {
		complexityPenalty = 8
	}
	score -= complexityPenalty

	// Dead code penalty (max 25 points, normalized for project size)
	deadCodePenalty := 0
	if s.DeadCodeCount > 0 {
		rawPenalty := float64(s.DeadCodeCount) / normalizationFactor
		deadCodePenalty = int(math.Min(25, rawPenalty))
	}

	// Additional penalty for critical dead code (max 15 points, normalized)
	criticalPenalty := 0
	if s.CriticalDeadCode > 0 {
		rawCriticalPenalty := float64(s.CriticalDeadCode*3) / normalizationFactor
		criticalPenalty = int(math.Min(15, rawCriticalPenalty))
	}

	totalDeadCodePenalty := deadCodePenalty + criticalPenalty
	if totalDeadCodePenalty > 25 {
		totalDeadCodePenalty = 25
	}
	score -= totalDeadCodePenalty

	// Clone penalty (max 25 points, based on percentage)
	clonePenalty := 0
	if s.CodeDuplication > 40 {
		clonePenalty = 25
	} else if s.CodeDuplication > 25 {
		clonePenalty = 15
	} else if s.CodeDuplication > 10 {
		clonePenalty = 8
	}
	score -= clonePenalty

	// CBO penalty (max 25 points)
	cboPenalty := 0
	if s.CBOClasses > 0 {
		couplingRatio := float64(s.HighCouplingClasses) / float64(s.CBOClasses)
		if couplingRatio > 0.5 {
			cboPenalty = 20
		} else if couplingRatio > 0.3 {
			cboPenalty = 12
		} else if couplingRatio > 0.1 {
			cboPenalty = 6
		}
	}
	score -= cboPenalty

	// Set minimum score to 10 (never completely fail)
	if score < 10 {
		score = 10
	}

	s.HealthScore = score

	// Assign grade based on score (adjusted thresholds)
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
