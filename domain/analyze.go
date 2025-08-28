package domain

import (
	"time"
)

// AnalyzeResponse represents the combined results of all analyses
type AnalyzeResponse struct {
	// Analysis results
	Complexity *ComplexityResponse `json:"complexity,omitempty" yaml:"complexity,omitempty"`
	DeadCode   *DeadCodeResponse   `json:"dead_code,omitempty" yaml:"dead_code,omitempty"`
	Clone      *CloneResponse      `json:"clone,omitempty" yaml:"clone,omitempty"`

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
	TotalFiles     int `json:"total_files" yaml:"total_files"`
	AnalyzedFiles  int `json:"analyzed_files" yaml:"analyzed_files"`
	SkippedFiles   int `json:"skipped_files" yaml:"skipped_files"`

	// Analysis status
	ComplexityEnabled bool `json:"complexity_enabled" yaml:"complexity_enabled"`
	DeadCodeEnabled   bool `json:"dead_code_enabled" yaml:"dead_code_enabled"`
	CloneEnabled      bool `json:"clone_enabled" yaml:"clone_enabled"`

	// Key metrics
	TotalFunctions      int     `json:"total_functions" yaml:"total_functions"`
	AverageComplexity   float64 `json:"average_complexity" yaml:"average_complexity"`
	HighComplexityCount int     `json:"high_complexity_count" yaml:"high_complexity_count"`
	
	DeadCodeCount       int     `json:"dead_code_count" yaml:"dead_code_count"`
	CriticalDeadCode    int     `json:"critical_dead_code" yaml:"critical_dead_code"`
	
	ClonePairs          int     `json:"clone_pairs" yaml:"clone_pairs"`
	CloneGroups         int     `json:"clone_groups" yaml:"clone_groups"`
	CodeDuplication     float64 `json:"code_duplication_percentage" yaml:"code_duplication_percentage"`

	// Overall health score (0-100)
	HealthScore int    `json:"health_score" yaml:"health_score"`
	Grade       string `json:"grade" yaml:"grade"` // A, B, C, D, F
}

// CalculateHealthScore calculates an overall health score based on analysis results
func (s *AnalyzeSummary) CalculateHealthScore() {
	score := 100

	// Deduct points for high complexity
	if s.AverageComplexity > 20 {
		score -= 30
	} else if s.AverageComplexity > 10 {
		score -= 20
	} else if s.AverageComplexity > 5 {
		score -= 10
	}

	// Deduct points for dead code
	if s.DeadCodeCount > 0 {
		deduction := (s.DeadCodeCount * 2)
		if deduction > 20 {
			deduction = 20
		}
		score -= deduction
	}

	// Deduct points for critical dead code
	score -= s.CriticalDeadCode * 5

	// Deduct points for code duplication
	if s.CodeDuplication > 30 {
		score -= 25
	} else if s.CodeDuplication > 20 {
		score -= 15
	} else if s.CodeDuplication > 10 {
		score -= 10
	}

	// Ensure score is within bounds
	if score < 0 {
		score = 0
	}

	s.HealthScore = score

	// Assign grade based on score
	switch {
	case score >= 90:
		s.Grade = "A"
	case score >= 80:
		s.Grade = "B"
	case score >= 70:
		s.Grade = "C"
	case score >= 60:
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
	return s.HighComplexityCount > 0 || s.DeadCodeCount > 0 || s.ClonePairs > 0
}