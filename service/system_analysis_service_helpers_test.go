package service

import (
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/analyzer"
	"github.com/ludo-technologies/pyscn/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildModuleLayerMap(t *testing.T) {
	service := NewSystemAnalysisService()
	graph := analyzer.NewDependencyGraph("/project")

	graph.AddModule("app.services.billing", "/project/app/services/billing.py")
	graph.AddModule("app.api.routes", "/project/app/api/routes.py")
	graph.AddModule("app.utilities.helpers", "/project/app/utilities/helpers.py")

	rules := &domain.ArchitectureRules{
		Layers: []domain.Layer{
			{Name: "application", Packages: []string{"app.services", "service"}},
			{Name: "presentation", Packages: []string{"app.api", "api"}},
		},
	}

	moduleToLayer := service.buildModuleLayerMap(graph, rules)

	assert.Equal(t, "application", moduleToLayer["app.services.billing"])
	assert.Equal(t, "presentation", moduleToLayer["app.api.routes"])
	assert.Equal(t, "unknown", moduleToLayer["app.utilities.helpers"])
}

func TestEvaluateLayerEdge(t *testing.T) {
	service := NewSystemAnalysisService()

	t.Run("strict mode warns on unknown layers", func(t *testing.T) {
		rules := &domain.ArchitectureRules{
			StrictMode: true,
			Rules: []domain.LayerRule{
				{From: "application", Allow: []string{"domain"}},
			},
		}

		violation := service.evaluateLayerEdge(rules, "app.services.billing", "app.domain.model", "unknown", "domain")
		require.NotNil(t, violation)
		assert.Equal(t, domain.ViolationSeverityWarning, violation.Severity)
		assert.Equal(t, "strict_mode", violation.Rule)
	})

	t.Run("strict mode warns when rule missing", func(t *testing.T) {
		rules := &domain.ArchitectureRules{
			StrictMode: true,
			Rules: []domain.LayerRule{
				{From: "application", Allow: []string{"domain"}},
			},
		}

		violation := service.evaluateLayerEdge(rules, "app.presentation.view", "app.domain.model", "presentation", "domain")
		require.NotNil(t, violation)
		assert.Equal(t, "no_rule", violation.Rule)
	})

	t.Run("deny rule triggers violation", func(t *testing.T) {
		rules := &domain.ArchitectureRules{
			Rules: []domain.LayerRule{
				{From: "application", Allow: []string{"domain"}, Deny: []string{"infrastructure"}},
			},
		}

		violation := service.evaluateLayerEdge(rules, "app.services.billing", "app.infrastructure.db", "application", "infrastructure")
		require.NotNil(t, violation)
		assert.Equal(t, domain.ViolationSeverityError, violation.Severity)
		assert.Equal(t, "application !> infrastructure", violation.Rule)
	})

	t.Run("allow list violation when target missing", func(t *testing.T) {
		rules := &domain.ArchitectureRules{
			Rules: []domain.LayerRule{
				{From: "domain", Allow: []string{"domain"}},
			},
		}

		violation := service.evaluateLayerEdge(rules, "app.domain.model", "app.application.service", "domain", "application")
		require.NotNil(t, violation)
		assert.Equal(t, "domain -> {domain}", violation.Rule)
	})

	t.Run("non-strict mode with missing rule returns nil", func(t *testing.T) {
		rules := &domain.ArchitectureRules{
			Rules: []domain.LayerRule{
				{From: "application", Allow: []string{"domain"}},
			},
		}

		assert.Nil(t, service.evaluateLayerEdge(rules, "app.presentation.view", "app.domain.model", "presentation", "domain"))
	})
}

func TestAutoDetectArchitecture(t *testing.T) {
	service := NewSystemAnalysisService()
	graph := analyzer.NewDependencyGraph("/project")

	graph.AddModule("app.api.users.router", "/project/app/api/users/router.py")
	graph.AddModule("app.services.user_service", "/project/app/services/user_service.py")
	graph.AddModule("app.domain.user_model", "/project/app/domain/user_model.py")
	graph.AddModule("app.infrastructure.db.client", "/project/app/infrastructure/db/client.py")

	rules := service.autoDetectArchitecture(graph)
	require.NotNil(t, rules)
	assert.True(t, rules.StrictMode)
	require.Greater(t, len(rules.Rules), 0)

	layerPackages := make(map[string][]string)
	for _, layer := range rules.Layers {
		layerPackages[layer.Name] = layer.Packages
	}

	require.Contains(t, layerPackages, "presentation")
	require.Contains(t, layerPackages, "application")
	require.Contains(t, layerPackages, "domain")
	require.Contains(t, layerPackages, "infrastructure")

	assert.Contains(t, layerPackages["presentation"], "app.api")
	assert.Contains(t, layerPackages["application"], "app.services")
	assert.Contains(t, layerPackages["domain"], "app.domain")
	assert.Contains(t, layerPackages["infrastructure"], "app")

	// Auto-detection should return nil when no layer patterns match
	graph = analyzer.NewDependencyGraph("/project")
	graph.AddModule("app.misc.utilities", "/project/app/misc/utilities.py")

	assert.Nil(t, service.autoDetectArchitecture(graph))
}

func TestAutoDetectArchitecture_FlatUnderscoreModules(t *testing.T) {
	service := NewSystemAnalysisService()
	graph := analyzer.NewDependencyGraph("/project")

	graph.AddModule("user_service", "/project/user_service.py")
	graph.AddModule("user_repository", "/project/user_repository.py")
	graph.AddModule("api_v1", "/project/api_v1.py")

	rules := service.autoDetectArchitecture(graph)
	require.NotNil(t, rules)

	layerPackages := make(map[string][]string)
	for _, layer := range rules.Layers {
		layerPackages[layer.Name] = layer.Packages
	}

	require.Contains(t, layerPackages, "presentation")
	require.Contains(t, layerPackages, "application")
	require.Contains(t, layerPackages, "infrastructure")
	assert.Contains(t, layerPackages["presentation"], "api_v1")
	assert.Contains(t, layerPackages["application"], "user_service")
	assert.Contains(t, layerPackages["infrastructure"], "user_repository")
}

func TestDependencyMatrixAndLongestChains(t *testing.T) {
	service := NewSystemAnalysisService()
	graph := analyzer.NewDependencyGraph("/project")

	graph.AddModule("moduleA", "/project/moduleA.py")
	graph.AddModule("moduleB", "/project/moduleB.py")
	graph.AddModule("moduleC", "/project/moduleC.py")
	graph.AddModule("moduleD", "/project/moduleD.py")

	graph.AddDependency("moduleA", "moduleB", analyzer.DependencyEdgeImport, nil)
	graph.AddDependency("moduleB", "moduleC", analyzer.DependencyEdgeImport, nil)
	graph.AddDependency("moduleC", "moduleD", analyzer.DependencyEdgeImport, nil)
	graph.AddDependency("moduleA", "moduleD", analyzer.DependencyEdgeImport, nil)

	matrix := service.buildDependencyMatrix(graph)
	require.Contains(t, matrix, "moduleA")
	require.True(t, matrix["moduleA"]["moduleB"])
	require.True(t, matrix["moduleA"]["moduleD"])
	require.False(t, matrix["moduleB"]["moduleA"])

	chains := service.findLongestChains(graph, 5)
	require.NotEmpty(t, chains)
	assert.Equal(t, 4, chains[0].Length)
	assert.Equal(t, []string{"moduleA", "moduleB", "moduleC", "moduleD"}, chains[0].Path)
	assert.LessOrEqual(t, len(chains), 5)
}

func TestConvertCouplingResults(t *testing.T) {
	service := NewSystemAnalysisService()

	assert.Nil(t, service.convertCouplingResults(nil))

	highCoupling := &analyzer.SystemMetrics{
		AverageFanIn:          0.5,
		AverageFanOut:         0.4,
		AverageInstability:    0.6,
		MainSequenceDeviation: 0.3,
		RefactoringPriority:   []string{"moduleA", "moduleB"},
	}

	result := service.convertCouplingResults(highCoupling)
	require.NotNil(t, result)
	assert.Equal(t, 0.9, result.AverageCoupling)
	assert.Equal(t, []string{"moduleA", "moduleB"}, result.HighlyCoupledModules)

	lowCoupling := &analyzer.SystemMetrics{
		AverageFanIn:        0.2,
		AverageFanOut:       0.2,
		RefactoringPriority: []string{"moduleA"},
	}

	result = service.convertCouplingResults(lowCoupling)
	require.NotNil(t, result)
	assert.Empty(t, result.HighlyCoupledModules)
}

func TestConvertCircularResults(t *testing.T) {
	service := NewSystemAnalysisService()

	circular := &analyzer.CircularDependencyResult{
		HasCircularDependencies: true,
		TotalCycles:             2,
		TotalModulesInCycles:    4,
		CircularDependencies: []*analyzer.CircularDependency{
			{
				Modules:     []string{"moduleA", "moduleB"},
				Description: "cycle between A and B",
				Severity:    analyzer.CycleSeverityLow,
				Size:        2,
			},
			{
				Modules:     []string{"moduleB", "moduleC", "moduleD"},
				Description: "three module cycle",
				Severity:    analyzer.CycleSeverityMedium,
				Size:        3,
			},
		},
	}

	result := service.convertCircularResults(circular)
	require.NotNil(t, result)
	assert.True(t, result.HasCircularDependencies)
	assert.Equal(t, 2, result.TotalCycles)
	assert.Equal(t, 4, result.TotalModulesInCycles)
	require.Len(t, result.CircularDependencies, 2)
	assert.Equal(t, []string{"moduleA", "moduleB"}, result.CircularDependencies[0].Modules)
	assert.NotEmpty(t, result.CycleBreakingSuggestions)
	require.Contains(t, result.CoreInfrastructure, "moduleB")
}

func TestGenerateArchitectureRecommendations(t *testing.T) {
	service := NewSystemAnalysisService()

	var violations []domain.ArchitectureViolation
	for i := 0; i < 11; i++ {
		violations = append(violations, domain.ArchitectureViolation{
			Module:   "app.services.billing",
			Severity: domain.ViolationSeverityError,
		})
	}
	violations = append(violations, domain.ArchitectureViolation{
		Module:   "app.services.payments",
		Severity: domain.ViolationSeverityError,
	})

	layerCohesion := map[string]float64{
		"application": 0.3,
	}
	problematicLayers := []string{"application"}

	recommendations := service.generateArchitectureRecommendations(violations, layerCohesion, problematicLayers, 0.5)
	require.Len(t, recommendations, 3)

	assert.Equal(t, domain.RecommendationTypeRestructure, recommendations[0].Type)
	assert.Equal(t, domain.RecommendationPriorityCritical, recommendations[0].Priority)

	assert.Equal(t, domain.RecommendationTypeRefactor, recommendations[1].Type)
	assert.Contains(t, recommendations[1].Modules, "app.services.billing")

	assert.Equal(t, domain.RecommendationTypeRestructure, recommendations[2].Type)
	assert.Contains(t, recommendations[2].Title, "application")
}

func TestIdentifyArchitectureRefactoringTargets(t *testing.T) {
	service := NewSystemAnalysisService()

	violations := []domain.ArchitectureViolation{
		{Module: "moduleA"},
		{Module: "moduleA"},
		{Module: "moduleB"},
		{Module: "moduleC"},
		{Module: "moduleC"},
		{Module: "moduleC"},
	}

	moduleToLayer := map[string]string{
		"moduleA": "application",
		"moduleB": "domain",
		"moduleC": "infrastructure",
	}

	targets := service.identifyArchitectureRefactoringTargets(violations, moduleToLayer)
	require.Equal(t, []string{"moduleC", "moduleA", "moduleB"}, targets)
}

func TestBuildModuleLayerMap_AmbiguousPackages(t *testing.T) {
	svc := NewSystemAnalysisService()
	graph := analyzer.NewDependencyGraph("/project")

	graph.AddModule("domain.routers", "/project/domain/routers/__init__.py")
	graph.AddModule("routers.main", "/project/routers/main.py")
	graph.AddModule("app.services.billing", "/project/app/services/billing.py")

	rules := &domain.ArchitectureRules{
		Layers: []domain.Layer{
			{Name: "domain", Packages: []string{"domain"}},
			{Name: "presentation", Packages: []string{"routers"}},
			{Name: "application", Packages: []string{"app.services"}},
		},
	}

	moduleToLayer := svc.buildModuleLayerMap(graph, rules)

	assert.Equal(t, "domain", moduleToLayer["domain.routers"],
		"domain.routers should be classified as 'domain' (prefix match wins)")
	assert.Equal(t, "presentation", moduleToLayer["routers.main"],
		"routers.main should be classified as 'presentation'")
	assert.Equal(t, "application", moduleToLayer["app.services.billing"],
		"app.services.billing should be classified as 'application'")
}

func TestPyscnConfigToSystemAnalysisRequest_PropagatesLayers(t *testing.T) {
	loader := NewSystemAnalysisConfigurationLoader()

	cfg := &config.PyscnConfig{
		ArchitectureLayers: []config.LayerDefinition{
			{Name: "api", Packages: []string{"myapp.api"}, Description: "API layer"},
			{Name: "domain", Packages: []string{"myapp.domain"}, Description: "Domain layer"},
		},
		ArchitectureRules: []config.LayerRule{
			{From: "api", Allow: []string{"domain"}, Deny: []string{"infrastructure"}},
		},
	}

	request := loader.pyscnConfigToSystemAnalysisRequest(cfg)

	require.NotNil(t, request.ArchitectureRules, "ArchitectureRules should be set")
	require.Len(t, request.ArchitectureRules.Layers, 2, "should have 2 layers")

	assert.Equal(t, "api", request.ArchitectureRules.Layers[0].Name)
	assert.Equal(t, []string{"myapp.api"}, request.ArchitectureRules.Layers[0].Packages)
	assert.Equal(t, "API layer", request.ArchitectureRules.Layers[0].Description)

	assert.Equal(t, "domain", request.ArchitectureRules.Layers[1].Name)
	assert.Equal(t, []string{"myapp.domain"}, request.ArchitectureRules.Layers[1].Packages)

	require.Len(t, request.ArchitectureRules.Rules, 1, "should have 1 rule")
	assert.Equal(t, "api", request.ArchitectureRules.Rules[0].From)
	assert.Equal(t, []string{"domain"}, request.ArchitectureRules.Rules[0].Allow)
	assert.Equal(t, []string{"infrastructure"}, request.ArchitectureRules.Rules[0].Deny)
}

func TestPyscnConfigToSystemAnalysisRequest_PropagatesLayersWithStrictMode(t *testing.T) {
	loader := NewSystemAnalysisConfigurationLoader()
	strictMode := true

	cfg := &config.PyscnConfig{
		ArchitectureStrictMode: &strictMode,
		ArchitectureLayers: []config.LayerDefinition{
			{Name: "api", Packages: []string{"api"}},
		},
	}

	request := loader.pyscnConfigToSystemAnalysisRequest(cfg)

	require.NotNil(t, request.ArchitectureRules)
	assert.True(t, request.ArchitectureRules.StrictMode)
	require.Len(t, request.ArchitectureRules.Layers, 1)
	assert.Equal(t, "api", request.ArchitectureRules.Layers[0].Name)
}

func TestPyscnConfigToSystemAnalysisRequest_LayersOnlyPreservesDefaultRules(t *testing.T) {
	loader := NewSystemAnalysisConfigurationLoader()

	cfg := &config.PyscnConfig{
		ArchitectureLayers: []config.LayerDefinition{
			{Name: "api", Packages: []string{"myapp.api"}},
			{Name: "domain", Packages: []string{"myapp.domain"}},
		},
	}

	request := loader.pyscnConfigToSystemAnalysisRequest(cfg)

	require.NotNil(t, request.ArchitectureRules, "ArchitectureRules should be set")
	require.Len(t, request.ArchitectureRules.Layers, 2, "should have user-provided layers")
	// Rules should be empty at this stage — auto-detection fills them at analysis time
	assert.Empty(t, request.ArchitectureRules.Rules, "rules should be empty in config; auto-detect fills them later")
}

func TestLoadDefaultRulesForLayers_FiltersToMatchingNames(t *testing.T) {
	svc := NewSystemAnalysisService()

	// Only "domain" matches a built-in rule name; "api" does not.
	layers := []domain.Layer{
		{Name: "domain", Packages: []string{"myapp.domain"}},
		{Name: "api", Packages: []string{"myapp.api"}},
	}

	rules := svc.loadDefaultRulesForLayers(layers)

	// Should only include rules whose From matches "domain" or "api".
	// "api" is not a built-in layer name, so it should be filtered out.
	for _, r := range rules {
		assert.True(t, r.From == "domain" || r.From == "api",
			"rule From=%q should match a user-defined layer name", r.From)
	}

	// "domain" is a built-in name, so we expect at least one rule for it
	hasDomainRule := false
	for _, r := range rules {
		if r.From == "domain" {
			hasDomainRule = true
			break
		}
	}
	assert.True(t, hasDomainRule, "should have default rules for 'domain' layer")
}

func TestLoadDefaultRulesForLayers_CustomNamesReturnEmpty(t *testing.T) {
	svc := NewSystemAnalysisService()

	// All custom names, none match built-in rule names
	layers := []domain.Layer{
		{Name: "my_service", Packages: []string{"svc"}},
		{Name: "my_api", Packages: []string{"api"}},
	}

	rules := svc.loadDefaultRulesForLayers(layers)
	assert.Empty(t, rules, "custom layer names should not match any built-in rules")
}

func TestPyscnConfigToSystemAnalysisRequest_RulesOnlyDoesNotBlockAutoDetect(t *testing.T) {
	loader := NewSystemAnalysisConfigurationLoader()

	cfg := &config.PyscnConfig{
		ArchitectureRules: []config.LayerRule{
			{From: "api", Allow: []string{"domain"}},
		},
	}

	request := loader.pyscnConfigToSystemAnalysisRequest(cfg)

	require.NotNil(t, request.ArchitectureRules, "ArchitectureRules should be set")
	require.Len(t, request.ArchitectureRules.Rules, 1, "should have user-provided rules")
	// Layers should be empty at this stage — auto-detection fills them at analysis time
	assert.Empty(t, request.ArchitectureRules.Layers, "layers should be empty in config; auto-detect fills them later")
}

func TestResolveArchitectureRules_DoesNotMutateOriginal(t *testing.T) {
	svc := NewSystemAnalysisService()
	graph := analyzer.NewDependencyGraph("/project")
	graph.AddModule("app.api.users.router", "/project/app/api/users/router.py")
	graph.AddModule("app.services.user_service", "/project/app/services/user_service.py")
	graph.AddModule("app.domain.user_model", "/project/app/domain/user_model.py")

	original := &domain.ArchitectureRules{
		Layers: []domain.Layer{
			{Name: "domain", Packages: []string{"domain"}},
		},
	}
	origRulesLen := len(original.Rules)
	origLayersLen := len(original.Layers)

	resolved := svc.resolveArchitectureRules(graph, original)

	// resolved may have filled in rules, but original must be untouched
	assert.Equal(t, origRulesLen, len(original.Rules), "original Rules must not be mutated")
	assert.Equal(t, origLayersLen, len(original.Layers), "original Layers must not be mutated")
	assert.NotSame(t, original, resolved, "resolved should be a new object")
}

func TestMergeLayerRules_UserOverridesDefaultForSameFrom(t *testing.T) {
	svc := NewSystemAnalysisService()

	base := []domain.LayerRule{
		{From: "domain", Allow: []string{"domain"}},
		{From: "application", Allow: []string{"domain", "infrastructure"}},
	}
	overrides := []domain.LayerRule{
		{From: "domain", Allow: []string{"domain", "application"}, Deny: []string{"infrastructure"}},
	}

	merged := svc.mergeLayerRules(base, overrides)

	// "domain" rule should come from overrides, "application" from base
	require.Len(t, merged, 2)

	rulesByFrom := make(map[string]domain.LayerRule)
	for _, r := range merged {
		rulesByFrom[r.From] = r
	}

	assert.Equal(t, []string{"domain", "application"}, rulesByFrom["domain"].Allow,
		"user override for 'domain' should replace base")
	assert.Equal(t, []string{"infrastructure"}, rulesByFrom["domain"].Deny,
		"user override for 'domain' should include Deny")
	assert.Equal(t, []string{"domain", "infrastructure"}, rulesByFrom["application"].Allow,
		"base rule for 'application' should be preserved")
}
