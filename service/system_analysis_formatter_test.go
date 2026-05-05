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

func TestSystemAnalysisFormatterIncludesMainSequenceZones(t *testing.T) {
	formatter := NewSystemAnalysisFormatter()
	response := &domain.SystemAnalysisResponse{
		DependencyAnalysis: &domain.DependencyAnalysisResult{
			TotalModules:      3,
			TotalDependencies: 2,
			CouplingAnalysis: &domain.CouplingAnalysis{
				AverageCoupling:       1.5,
				AverageInstability:    0.4,
				MainSequenceDeviation: 0.2,
				ZoneOfPain:            []string{"domain.core"},
				ZoneOfUselessness:     []string{"unused.contracts"},
				MainSequence:          []string{"balanced.service"},
			},
		},
	}

	textOutput, err := formatter.Format(response, domain.OutputFormatText)
	require.NoError(t, err)
	assert.Contains(t, textOutput, "Zone of Pain")
	assert.Contains(t, textOutput, "domain.core")
	assert.Contains(t, textOutput, "Zone of Uselessness")
	assert.Contains(t, textOutput, "unused.contracts")
	assert.Contains(t, textOutput, "Main Sequence")
	assert.Contains(t, textOutput, "balanced.service")

	htmlOutput, err := formatter.Format(response, domain.OutputFormatHTML)
	require.NoError(t, err)
	assert.Contains(t, htmlOutput, "Zone of Pain")
	assert.Contains(t, htmlOutput, "domain.core")
	assert.Contains(t, htmlOutput, "Zone of Uselessness")
	assert.Contains(t, htmlOutput, "unused.contracts")

	csvOutput, err := formatter.Format(response, domain.OutputFormatCSV)
	require.NoError(t, err)
	assert.Contains(t, csvOutput, "Zone of Pain,domain.core")
	assert.Contains(t, csvOutput, "Zone of Uselessness,unused.contracts")
	assert.Contains(t, csvOutput, "Main Sequence,balanced.service")
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
