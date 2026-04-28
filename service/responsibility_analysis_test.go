package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/analyzer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeResponsibilityDetectsHubModule(t *testing.T) {
	service := NewSystemAnalysisService()
	graph := analyzer.NewDependencyGraph("/test/project")

	modules := []string{
		"app.core.hub",
		"app.api.views",
		"app.auth.policy",
		"app.billing.invoice",
		"app.db.repo",
		"app.reporting.export",
	}
	for _, module := range modules {
		graph.AddModule(module, "/test/project/"+module+".py")
	}

	for _, dependency := range modules[1:] {
		graph.AddDependency("app.core.hub", dependency, analyzer.DependencyEdgeImport, nil)
		graph.AddDependency(dependency, "app.core.hub", analyzer.DependencyEdgeImport, nil)
	}

	responsibility, cohesion, violations := service.analyzeResponsibility(graph, defaultResponsibilityOptions())

	require.NotNil(t, responsibility)
	require.NotNil(t, cohesion)
	require.Len(t, responsibility.SRPViolations, 1)
	assert.Equal(t, "app.core.hub", responsibility.SRPViolations[0].Module)
	assert.Equal(t, []string{"api", "auth", "billing", "db", "reporting"}, responsibility.SRPViolations[0].Responsibilities)
	assert.Equal(t, []string{"app.core.hub"}, responsibility.OverloadedModules)
	assert.Equal(t, domain.ViolationTypeResponsibility, violations[0].Type)
	assert.Equal(t, domain.ViolationSeverityError, violations[0].Severity)
}

func TestAnalyzeResponsibilityKeepsCohesivePackageClean(t *testing.T) {
	service := NewSystemAnalysisService()
	graph := analyzer.NewDependencyGraph("/test/project")

	graph.AddModule("app.orders.api", "/test/project/app/orders/api.py")
	graph.AddModule("app.orders.service", "/test/project/app/orders/service.py")
	graph.AddModule("app.orders.repo", "/test/project/app/orders/repo.py")
	graph.AddDependency("app.orders.api", "app.orders.service", analyzer.DependencyEdgeImport, nil)
	graph.AddDependency("app.orders.service", "app.orders.repo", analyzer.DependencyEdgeImport, nil)

	responsibility, cohesion, violations := service.analyzeResponsibility(graph, defaultResponsibilityOptions())

	require.NotNil(t, responsibility)
	require.NotNil(t, cohesion)
	assert.Empty(t, responsibility.SRPViolations)
	assert.Empty(t, responsibility.OverloadedModules)
	assert.Empty(t, violations)
	assert.Empty(t, cohesion.LowCohesionPackages)
	assert.InDelta(t, 1.0, cohesion.PackageCohesion["app.orders"], 0.01)
}

func TestAnalyzePackageCohesionFlagsScatteredPackage(t *testing.T) {
	graph := analyzer.NewDependencyGraph("/test/project")

	graph.AddModule("app.orders.api", "/test/project/app/orders/api.py")
	graph.AddModule("app.orders.worker", "/test/project/app/orders/worker.py")
	graph.AddModule("app.billing.invoice", "/test/project/app/billing/invoice.py")
	graph.AddModule("app.reporting.export", "/test/project/app/reporting/export.py")

	graph.AddDependency("app.orders.api", "app.billing.invoice", analyzer.DependencyEdgeImport, nil)
	graph.AddDependency("app.orders.worker", "app.reporting.export", analyzer.DependencyEdgeImport, nil)

	cohesion := analyzePackageCohesion(graph, defaultMinPackageCohesion)

	assert.Contains(t, cohesion.LowCohesionPackages, "app.orders")
	assert.Equal(t, 0.0, cohesion.PackageCohesion["app.orders"])
	assert.NotEmpty(t, cohesion.CohesionSuggestions["app.orders"])
}

func TestResponsibilityOptionsFromRequestUsesConfiguredThresholds(t *testing.T) {
	options := responsibilityOptionsFromRequest(domain.SystemAnalysisRequest{
		MinCohesion:                     0.75,
		MaxResponsibilities:             2,
		ResponsibilityViolationSeverity: domain.ViolationSeverityCritical,
	})

	assert.Equal(t, 0.75, options.minPackageCohesion)
	assert.Equal(t, 2, options.maxResponsibilities)
	assert.Equal(t, domain.ViolationSeverityCritical, options.severity)
}

func TestParseViolationSeverityFallsBackToWarning(t *testing.T) {
	assert.Equal(t, domain.ViolationSeverityInfo, parseViolationSeverity("info"))
	assert.Equal(t, domain.ViolationSeverityError, parseViolationSeverity("error"))
	assert.Equal(t, domain.ViolationSeverityCritical, parseViolationSeverity("critical"))
	assert.Equal(t, domain.ViolationSeverityWarning, parseViolationSeverity("unknown"))
}

func TestAnalyzeArchitectureRunsResponsibilityWithoutLayerRules(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "pyproject.toml"), []byte("[project]\nname = \"srp-repro\"\n"), 0o644))

	files := map[string]string{
		"app/__init__.py":        "",
		"app/red/__init__.py":    "",
		"app/blue/__init__.py":   "",
		"app/green/__init__.py":  "",
		"app/yellow/__init__.py": "",
		"app/orange/__init__.py": "",
		"app/red/hub.py":         "from ..blue.node import render\nfrom ..green.node import save\nfrom ..yellow.node import authorize\nfrom ..orange.node import export\n\ndef process(user, payload):\n    authorize(user)\n    result = render(payload)\n    save(result)\n    return export(result)\n",
		"app/blue/node.py":       "from ..red.hub import process\n\ndef render(payload):\n    return payload\n",
		"app/green/node.py":      "from ..red.hub import process\n\ndef save(row):\n    return row\n",
		"app/yellow/node.py":     "from ..red.hub import process\n\ndef authorize(user):\n    return bool(user)\n",
		"app/orange/node.py":     "from ..red.hub import process\n\ndef export(result):\n    return result\n",
	}

	paths := make([]string, 0, len(files))
	for name, content := range files {
		path := filepath.Join(root, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
		if filepath.Ext(path) == ".py" {
			paths = append(paths, path)
		}
	}

	service := NewSystemAnalysisService()
	result, err := service.AnalyzeArchitecture(context.Background(), domain.SystemAnalysisRequest{
		Paths:             paths,
		IncludeStdLib:     domain.BoolPtr(false),
		IncludeThirdParty: domain.BoolPtr(false),
		FollowRelative:    domain.BoolPtr(true),
	})

	require.NoError(t, err)
	require.NotNil(t, result.ResponsibilityAnalysis)
	require.Len(t, result.ResponsibilityAnalysis.SRPViolations, 1)
	assert.Equal(t, "app.red.hub", result.ResponsibilityAnalysis.SRPViolations[0].Module)
	assert.Equal(t, 1, result.TotalViolations)
	assert.Zero(t, result.LayerAnalysis.LayersAnalyzed)
}
