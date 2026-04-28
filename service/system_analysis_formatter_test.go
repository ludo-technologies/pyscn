package service

import (
	"strings"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemAnalysisFormatterIncludesResponsibilityAnalysisText(t *testing.T) {
	formatter := NewSystemAnalysisFormatter()
	response := systemResponseWithResponsibilityViolation()

	output, err := formatter.Format(response, domain.OutputFormatText)

	require.NoError(t, err)
	assert.Contains(t, output, "RESPONSIBILITY VIOLATIONS")
	assert.Contains(t, output, "app.core.hub")
	assert.Contains(t, output, "api, auth, db")
	assert.Contains(t, output, "LOW PACKAGE COHESION")
	assert.Contains(t, output, "app.orders")
}

func TestSystemAnalysisFormatterIncludesResponsibilityAnalysisHTML(t *testing.T) {
	formatter := NewSystemAnalysisFormatter()
	response := systemResponseWithResponsibilityViolation()

	output, err := formatter.Format(response, domain.OutputFormatHTML)

	require.NoError(t, err)
	assert.Contains(t, output, "Responsibility Violations")
	assert.Contains(t, output, "app.core.hub")
	assert.Contains(t, output, "api, auth, db")
	assert.Contains(t, output, "Package Cohesion")
	assert.Contains(t, output, "app.orders")
	assert.Contains(t, output, `badge bg-danger`)
	assert.False(t, strings.Contains(output, "<nil>"))
}

func systemResponseWithResponsibilityViolation() *domain.SystemAnalysisResponse {
	return &domain.SystemAnalysisResponse{
		ArchitectureAnalysis: &domain.ArchitectureAnalysisResult{
			ComplianceScore: 0.75,
			TotalViolations: 1,
			TotalRules:      1,
			LayerAnalysis: &domain.LayerAnalysis{
				LayersAnalyzed:    1,
				LayerViolations:   []domain.LayerViolation{},
				LayerCoupling:     map[string]map[string]int{},
				LayerCohesion:     map[string]float64{},
				ProblematicLayers: []string{},
			},
			CohesionAnalysis: &domain.CohesionAnalysis{
				PackageCohesion:     map[string]float64{"app.orders": 0.25},
				LowCohesionPackages: []string{"app.orders"},
				CohesionSuggestions: map[string]string{"app.orders": "Split unrelated order concerns"},
			},
			ResponsibilityAnalysis: &domain.ResponsibilityAnalysis{
				SRPViolations: []domain.SRPViolation{
					{
						Module:           "app.core.hub",
						Responsibilities: []string{"api", "auth", "db"},
						Severity:         domain.ViolationSeverityError,
						Suggestion:       "Split hub module by concern",
					},
				},
				ModuleResponsibilities: map[string][]string{
					"app.core.hub": {"api", "auth", "db"},
				},
				OverloadedModules: []string{"app.core.hub"},
			},
			Violations: []domain.ArchitectureViolation{
				{
					Type:     domain.ViolationTypeResponsibility,
					Severity: domain.ViolationSeverityError,
					Module:   "app.core.hub",
				},
			},
			SeverityBreakdown:  map[domain.ViolationSeverity]int{domain.ViolationSeverityError: 1},
			Recommendations:    []domain.ArchitectureRecommendation{},
			RefactoringTargets: []string{"app.core.hub"},
		},
	}
}
