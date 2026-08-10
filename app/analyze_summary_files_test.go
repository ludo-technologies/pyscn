package app

import (
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

// calculateSummary used to set both TotalFiles and AnalyzedFiles from
// FilesAnalyzed, which made the two always equal and hid every file that
// failed to parse from the health score (issue #690).
func TestCalculateSummaryCarriesTheFileShortfall(t *testing.T) {
	useCase := &AnalyzeUseCase{}
	response := &domain.AnalyzeResponse{
		Complexity: &domain.ComplexityResponse{
			Summary: domain.ComplexitySummary{
				FilesAnalyzed: 1,
				TotalFiles:    2,
				SkippedFiles:  1,
			},
		},
	}

	summary := domain.AnalyzeSummary{}
	useCase.calculateSummary(&summary, response)

	if summary.TotalFiles != 2 {
		t.Errorf("TotalFiles = %d, want 2 (parsed and unparsed alike)", summary.TotalFiles)
	}
	if summary.AnalyzedFiles != 1 {
		t.Errorf("AnalyzedFiles = %d, want 1", summary.AnalyzedFiles)
	}
	if summary.SkippedFiles != 1 {
		t.Errorf("SkippedFiles = %d, want 1", summary.SkippedFiles)
	}
}
