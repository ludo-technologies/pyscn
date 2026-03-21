package integration

import (
	"bytes"
	"context"
	"testing"

	"github.com/ludo-technologies/pyscn/app"
	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fastapiLayersDir = "../testdata/python/fastapi_layers"

func newArchitectureUseCase() *app.SystemAnalysisUseCase {
	return app.NewSystemAnalysisUseCase(
		service.NewSystemAnalysisService(),
		service.NewFileReader(),
		service.NewSystemAnalysisFormatter(),
		service.NewSystemAnalysisConfigurationLoader(),
	)
}

func analyzeArchitecture(t *testing.T, dir string) *domain.ArchitectureAnalysisResult {
	t.Helper()
	uc := newArchitectureUseCase()
	var buf bytes.Buffer
	result, err := uc.AnalyzeArchitectureOnly(context.Background(), domain.SystemAnalysisRequest{
		Paths:        []string{dir},
		ConfigPath:   dir,
		OutputFormat: domain.OutputFormatJSON,
		OutputWriter: &buf,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	return result
}

// TestArchitecture_TOMLConfigLoading verifies that layers and rules defined in
// .pyscn.toml are correctly parsed and propagated to the analysis request.
// Regression test for PR #356 (TOML layers/rules silently ignored).
func TestArchitecture_TOMLConfigLoading(t *testing.T) {
	configLoader := service.NewSystemAnalysisConfigurationLoader()
	cfg, err := configLoader.LoadConfig(fastapiLayersDir)
	require.NoError(t, err)
	require.NotNil(t, cfg.ArchitectureRules)

	// Verify layers
	require.Len(t, cfg.ArchitectureRules.Layers, 3)

	layersByName := make(map[string]domain.Layer)
	for _, l := range cfg.ArchitectureRules.Layers {
		layersByName[l.Name] = l
	}

	presentation := layersByName["presentation"]
	assert.ElementsMatch(t, []string{"routers", "pages"}, presentation.Packages)

	domainLayer := layersByName["domain"]
	assert.ElementsMatch(t, []string{"domain"}, domainLayer.Packages)

	infra := layersByName["infrastructure"]
	assert.ElementsMatch(t, []string{"repositories"}, infra.Packages)

	// Verify rules
	require.Len(t, cfg.ArchitectureRules.Rules, 3)

	rulesByFrom := make(map[string]domain.LayerRule)
	for _, r := range cfg.ArchitectureRules.Rules {
		rulesByFrom[r.From] = r
	}

	assert.ElementsMatch(t, []string{"presentation", "domain", "infrastructure"},
		rulesByFrom["presentation"].Allow)
	assert.ElementsMatch(t, []string{"domain"},
		rulesByFrom["domain"].Allow)
	assert.ElementsMatch(t, []string{"infrastructure", "domain"},
		rulesByFrom["infrastructure"].Allow)
}

// TestArchitecture_LayerClassificationAndCoupling verifies that modules are
// assigned to the correct layers and the dependency directions are as expected.
// Regression test for discussion #352 (routers misclassified as "domain").
func TestArchitecture_LayerClassificationAndCoupling(t *testing.T) {
	result := analyzeArchitecture(t, fastapiLayersDir)
	require.NotNil(t, result.LayerAnalysis)

	coupling := result.LayerAnalysis.LayerCoupling

	// presentation -> domain: routers and pages import domain models
	require.Contains(t, coupling, "presentation")
	assert.Contains(t, coupling["presentation"], "domain")

	// presentation -> infrastructure: item_router imports item_repo
	assert.Contains(t, coupling["presentation"], "infrastructure")

	// infrastructure -> domain: repos import domain models
	require.Contains(t, coupling, "infrastructure")
	assert.Contains(t, coupling["infrastructure"], "domain")

	// domain must NOT have cross-layer outgoing dependencies to presentation
	// (except for the intentional violation fixture in domain/services/)
	// The violation fixture adds domain -> presentation, so we check that
	// it IS detected rather than asserting it doesn't exist.
	if domainDeps, ok := coupling["domain"]; ok {
		if _, hasPresentationDep := domainDeps["presentation"]; hasPresentationDep {
			// This is expected due to domain/services/user_service.py importing pages
			t.Logf("domain -> presentation coupling detected (expected from violation fixture)")
		}
	}
}

// TestArchitecture_ViolationDetection verifies that a domain -> presentation
// dependency is flagged as a violation. The fixture domain/services/user_service.py
// intentionally imports from the presentation layer.
func TestArchitecture_ViolationDetection(t *testing.T) {
	result := analyzeArchitecture(t, fastapiLayersDir)
	require.NotNil(t, result.LayerAnalysis)

	// There must be at least one violation: domain -> presentation
	assert.Greater(t, result.TotalViolations, 0,
		"should detect domain -> presentation violation")
	assert.Less(t, result.ComplianceScore, 1.0,
		"compliance should be less than 100%% with violations")

	// Find the specific domain -> presentation violation
	var found bool
	for _, v := range result.LayerAnalysis.LayerViolations {
		if v.FromLayer == "domain" && v.ToLayer == "presentation" {
			found = true
			assert.Contains(t, v.FromModule, "user_service",
				"violation should originate from domain/services/user_service")
			t.Logf("violation detected: %s (%s) -> %s (%s)", v.FromModule, v.FromLayer, v.ToModule, v.ToLayer)
			break
		}
	}
	assert.True(t, found, "should find a domain -> presentation layer violation")
}

// TestArchitecture_AllowedDepsNotFlagged verifies that dependencies permitted by
// the rules (presentation -> domain, presentation -> infrastructure,
// infrastructure -> domain) are NOT flagged as violations.
func TestArchitecture_AllowedDepsNotFlagged(t *testing.T) {
	result := analyzeArchitecture(t, fastapiLayersDir)
	require.NotNil(t, result.LayerAnalysis)

	for _, v := range result.LayerAnalysis.LayerViolations {
		// presentation -> anything is allowed per config
		if v.FromLayer == "presentation" {
			t.Errorf("unexpected violation from presentation: %s (%s) -> %s (%s)",
				v.FromModule, v.FromLayer, v.ToModule, v.ToLayer)
		}
		// infrastructure -> domain is allowed per config
		if v.FromLayer == "infrastructure" && v.ToLayer == "domain" {
			t.Errorf("unexpected violation: infrastructure -> domain: %s -> %s",
				v.FromModule, v.ToModule)
		}
	}
}
