package domain_test

import (
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

// A file that fails to parse yields no functions, no dead code and no clones,
// so before issue #690 a project half of which did not compile scored the same
// as one that fully did. Every case below shares the same clean metrics and
// differs only in how much of the project was actually analyzed.
func TestCalculateHealthScorePenalizesUnanalyzedFiles(t *testing.T) {
	cleanSummary := func(total, skipped int) domain.AnalyzeSummary {
		return domain.AnalyzeSummary{
			AverageComplexity: 2.0,
			CodeDuplication:   0.0,
			ArchEnabled:       true,
			ArchCompliance:    1.0,
			TotalFiles:        total,
			AnalyzedFiles:     total - skipped,
			SkippedFiles:      skipped,
		}
	}

	tests := []struct {
		name          string
		totalFiles    int
		skippedFiles  int
		expectedScore int
		expectedGrade string
	}{
		{
			name:          "everything parsed keeps a perfect score",
			totalFiles:    2,
			skippedFiles:  0,
			expectedScore: 100,
			expectedGrade: "A",
		},
		{
			// The floor applies: one broken file in a large tree is a small
			// fraction but still forfeits the A.
			name:          "one file of fifty forfeits the top grade",
			totalFiles:    50,
			skippedFiles:  1,
			expectedScore: 100 - domain.MinParseErrorPenalty,
			expectedGrade: "B",
		},
		{
			name:          "half the project unparsed costs half the maximum",
			totalFiles:    2,
			skippedFiles:  1,
			expectedScore: 100 - domain.MaxParseErrorPenalty/2,
			expectedGrade: "C",
		},
		{
			name:          "nothing parsed cannot rank above the lowest grade",
			totalFiles:    3,
			skippedFiles:  3,
			expectedScore: 100 - domain.MaxParseErrorPenalty,
			expectedGrade: "F",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := cleanSummary(tt.totalFiles, tt.skippedFiles)
			if err := summary.CalculateHealthScore(); err != nil {
				t.Fatalf("CalculateHealthScore() error: %v", err)
			}

			if summary.HealthScore != tt.expectedScore {
				t.Errorf("HealthScore = %d, want %d", summary.HealthScore, tt.expectedScore)
			}
			if summary.Grade != tt.expectedGrade {
				t.Errorf("Grade = %q, want %q", summary.Grade, tt.expectedGrade)
			}
		})
	}
}

// The per-category scores stay clean because the skipped file contributes
// nothing to them; only the overall score records the shortfall. Asserting
// this keeps the penalty from being mistaken for a complexity regression.
func TestParseErrorPenaltyLeavesCategoryScoresUntouched(t *testing.T) {
	summary := domain.AnalyzeSummary{
		AverageComplexity: 2.0,
		TotalFiles:        2,
		AnalyzedFiles:     1,
		SkippedFiles:      1,
	}

	if err := summary.CalculateHealthScore(); err != nil {
		t.Fatalf("CalculateHealthScore() error: %v", err)
	}

	if summary.ComplexityScore != 100 {
		t.Errorf("ComplexityScore = %d, want 100", summary.ComplexityScore)
	}
	if summary.HealthScore >= 100 {
		t.Errorf("HealthScore = %d, want a penalised score below 100", summary.HealthScore)
	}
}

func TestValidateRejectsImpossibleSkippedCount(t *testing.T) {
	summary := domain.AnalyzeSummary{
		TotalFiles:   1,
		SkippedFiles: 2,
	}

	if err := summary.Validate(); err == nil {
		t.Fatal("expected SkippedFiles exceeding TotalFiles to be rejected")
	}
}
