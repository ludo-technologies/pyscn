package service

import (
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/analyzer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractModuleMetrics(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (*SystemAnalysisServiceImpl, *analyzer.DependencyGraph)
		validate func(t *testing.T, result map[string]*domain.ModuleDependencyMetrics)
	}{
		{
			name: "extract_metrics_with_node_details",
			setup: func() (*SystemAnalysisServiceImpl, *analyzer.DependencyGraph) {
				service := &SystemAnalysisServiceImpl{}
				graph := analyzer.NewDependencyGraph("/test/project")

				// Add a module with full details
				module := graph.AddModule("test.module", "/test/project/test/module.py")
				module.Package = "test"
				module.LineCount = 150
				module.FunctionCount = 10
				module.ClassCount = 3
				module.PublicNames = []string{"TestClass", "test_function", "CONSTANT"}

				// Add dependencies
				graph.AddModule("test.other", "/test/project/test/other.py")
				graph.AddDependency("test.module", "test.other", analyzer.DependencyEdgeImport, nil)

				return service, graph
			},
			validate: func(t *testing.T, result map[string]*domain.ModuleDependencyMetrics) {
				require.Contains(t, result, "test.module")
				metrics := result["test.module"]

				// Basic information
				assert.Equal(t, "test.module", metrics.ModuleName)
				assert.Equal(t, "/test/project/test/module.py", metrics.FilePath)
				assert.Equal(t, "test", metrics.Package)
				assert.False(t, metrics.IsPackage)

				// Size metrics from node
				assert.Equal(t, 150, metrics.LinesOfCode)
				assert.Equal(t, 10, metrics.FunctionCount)
				assert.Equal(t, 3, metrics.ClassCount)
				assert.Equal(t, []string{"TestClass", "test_function", "CONSTANT"}, metrics.PublicInterface)

				// Dependencies
				assert.Contains(t, metrics.DirectDependencies, "test.other")
			},
		},
		{
			name: "extract_metrics_for_package_init",
			setup: func() (*SystemAnalysisServiceImpl, *analyzer.DependencyGraph) {
				service := &SystemAnalysisServiceImpl{}
				graph := analyzer.NewDependencyGraph("/test/project")

				// Add a package __init__.py
				module := graph.AddModule("mypackage", "/test/project/mypackage/__init__.py")
				module.Package = "mypackage"
				module.IsPackage = true
				module.LineCount = 50
				module.FunctionCount = 2
				module.ClassCount = 0
				module.PublicNames = []string{"initialize", "VERSION"}

				return service, graph
			},
			validate: func(t *testing.T, result map[string]*domain.ModuleDependencyMetrics) {
				require.Contains(t, result, "mypackage")
				metrics := result["mypackage"]

				assert.True(t, metrics.IsPackage)
				assert.Equal(t, "mypackage", metrics.Package)
				assert.Equal(t, 50, metrics.LinesOfCode)
				assert.Equal(t, 2, metrics.FunctionCount)
				assert.Equal(t, 0, metrics.ClassCount)
				assert.Equal(t, []string{"initialize", "VERSION"}, metrics.PublicInterface)
			},
		},
		{
			name: "extract_metrics_with_analyzer_metrics",
			setup: func() (*SystemAnalysisServiceImpl, *analyzer.DependencyGraph) {
				service := &SystemAnalysisServiceImpl{}
				graph := analyzer.NewDependencyGraph("/test/project")

				// Add module with both node data and analyzer metrics
				module := graph.AddModule("analyzed.module", "/test/project/analyzed/module.py")
				module.Package = "analyzed"
				module.LineCount = 200
				module.FunctionCount = 15
				module.ClassCount = 5
				module.PublicNames = []string{"AnalyzedClass", "process", "validate"}
				module.InDegree = 3
				module.OutDegree = 2

				// Add analyzer metrics
				graph.ModuleMetrics = make(map[string]*analyzer.ModuleMetrics)
				graph.ModuleMetrics["analyzed.module"] = &analyzer.ModuleMetrics{
					AfferentCoupling:     3,
					EfferentCoupling:     2,
					Instability:          0.4,
					Abstractness:         0.3,
					Distance:             0.5, // Medium risk threshold
					LinesOfCode:          200,
					PublicInterface:      3,
					CyclomaticComplexity: 10,
				}

				return service, graph
			},
			validate: func(t *testing.T, result map[string]*domain.ModuleDependencyMetrics) {
				require.Contains(t, result, "analyzed.module")
				metrics := result["analyzed.module"]

				// Node data should still be populated
				assert.Equal(t, "analyzed", metrics.Package)
				assert.Equal(t, 200, metrics.LinesOfCode)
				assert.Equal(t, 15, metrics.FunctionCount)
				assert.Equal(t, 5, metrics.ClassCount)
				assert.Equal(t, []string{"AnalyzedClass", "process", "validate"}, metrics.PublicInterface)

				// Analyzer metrics should be used
				assert.Equal(t, 3, metrics.AfferentCoupling)
				assert.Equal(t, 2, metrics.EfferentCoupling)
				assert.InDelta(t, 0.4, metrics.Instability, 0.01)
				assert.InDelta(t, 0.3, metrics.Abstractness, 0.01)
				assert.InDelta(t, 0.5, metrics.Distance, 0.01)
				assert.Equal(t, domain.RiskLevelMedium, metrics.RiskLevel) // Distance=0.5 triggers medium risk
			},
		},
		{
			name: "extract_metrics_with_high_risk",
			setup: func() (*SystemAnalysisServiceImpl, *analyzer.DependencyGraph) {
				service := &SystemAnalysisServiceImpl{}
				graph := analyzer.NewDependencyGraph("/test/project")

				module := graph.AddModule("risky.module", "/test/project/risky/module.py")
				module.Package = "risky"
				module.LineCount = 500
				module.FunctionCount = 30
				module.ClassCount = 10
				module.PublicNames = []string{}

				// Add high-risk analyzer metrics
				graph.ModuleMetrics = make(map[string]*analyzer.ModuleMetrics)
				graph.ModuleMetrics["risky.module"] = &analyzer.ModuleMetrics{
					Distance: 0.8, // High distance from main sequence
				}

				return service, graph
			},
			validate: func(t *testing.T, result map[string]*domain.ModuleDependencyMetrics) {
				require.Contains(t, result, "risky.module")
				metrics := result["risky.module"]

				assert.Equal(t, domain.RiskLevelHigh, metrics.RiskLevel)
				assert.Equal(t, 500, metrics.LinesOfCode)
				assert.Empty(t, metrics.PublicInterface)
			},
		},
		{
			name: "extract_metrics_fallback_without_analyzer_metrics",
			setup: func() (*SystemAnalysisServiceImpl, *analyzer.DependencyGraph) {
				service := &SystemAnalysisServiceImpl{}
				graph := analyzer.NewDependencyGraph("/test/project")

				module := graph.AddModule("simple.module", "/test/project/simple/module.py")
				module.Package = "simple"
				module.LineCount = 100
				module.FunctionCount = 5
				module.ClassCount = 2
				module.PublicNames = []string{"SimpleClass"}
				module.InDegree = 2
				module.OutDegree = 3

				// No analyzer metrics

				return service, graph
			},
			validate: func(t *testing.T, result map[string]*domain.ModuleDependencyMetrics) {
				require.Contains(t, result, "simple.module")
				metrics := result["simple.module"]

				// Node data should be populated
				assert.Equal(t, "simple", metrics.Package)
				assert.Equal(t, 100, metrics.LinesOfCode)
				assert.Equal(t, 5, metrics.FunctionCount)
				assert.Equal(t, 2, metrics.ClassCount)
				assert.Equal(t, []string{"SimpleClass"}, metrics.PublicInterface)

				// Fallback metrics from node degrees
				assert.Equal(t, 2, metrics.AfferentCoupling)
				assert.Equal(t, 3, metrics.EfferentCoupling)
				assert.InDelta(t, 0.6, metrics.Instability, 0.01) // 3 / (2 + 3)
				assert.Equal(t, domain.RiskLevelLow, metrics.RiskLevel)
			},
		},
		{
			name: "extract_metrics_empty_module",
			setup: func() (*SystemAnalysisServiceImpl, *analyzer.DependencyGraph) {
				service := &SystemAnalysisServiceImpl{}
				graph := analyzer.NewDependencyGraph("/test/project")

				// Add an empty module
				module := graph.AddModule("empty.module", "/test/project/empty/module.py")
				module.Package = "empty"
				// All counts remain at zero
				// PublicNames remains empty

				return service, graph
			},
			validate: func(t *testing.T, result map[string]*domain.ModuleDependencyMetrics) {
				require.Contains(t, result, "empty.module")
				metrics := result["empty.module"]

				assert.Equal(t, "empty", metrics.Package)
				assert.Equal(t, 0, metrics.LinesOfCode)
				assert.Equal(t, 0, metrics.FunctionCount)
				assert.Equal(t, 0, metrics.ClassCount)
				assert.Empty(t, metrics.PublicInterface)
				assert.Equal(t, 0, metrics.AfferentCoupling)
				assert.Equal(t, 0, metrics.EfferentCoupling)
				assert.Equal(t, 0.0, metrics.Instability)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, graph := tt.setup()
			result := service.extractModuleMetrics(graph)
			tt.validate(t, result)
		})
	}
}

func TestExtractModuleMetrics_MultipleModules(t *testing.T) {
	service := &SystemAnalysisServiceImpl{}
	graph := analyzer.NewDependencyGraph("/test/project")

	// Create a small dependency network
	moduleA := graph.AddModule("package.moduleA", "/test/project/package/moduleA.py")
	moduleA.Package = "package"
	moduleA.LineCount = 100
	moduleA.FunctionCount = 5
	moduleA.ClassCount = 2
	moduleA.PublicNames = []string{"ClassA", "funcA"}

	moduleB := graph.AddModule("package.moduleB", "/test/project/package/moduleB.py")
	moduleB.Package = "package"
	moduleB.LineCount = 150
	moduleB.FunctionCount = 8
	moduleB.ClassCount = 3
	moduleB.PublicNames = []string{"ClassB", "funcB", "helperB"}

	moduleC := graph.AddModule("package.moduleC", "/test/project/package/moduleC.py")
	moduleC.Package = "package"
	moduleC.LineCount = 75
	moduleC.FunctionCount = 3
	moduleC.ClassCount = 1
	moduleC.PublicNames = []string{"ClassC"}

	// Add dependencies: A -> B, B -> C, C -> A (cycle)
	graph.AddDependency("package.moduleA", "package.moduleB", analyzer.DependencyEdgeImport, nil)
	graph.AddDependency("package.moduleB", "package.moduleC", analyzer.DependencyEdgeImport, nil)
	graph.AddDependency("package.moduleC", "package.moduleA", analyzer.DependencyEdgeImport, nil)

	result := service.extractModuleMetrics(graph)

	// Verify all modules are present
	assert.Len(t, result, 3)
	assert.Contains(t, result, "package.moduleA")
	assert.Contains(t, result, "package.moduleB")
	assert.Contains(t, result, "package.moduleC")

	// Verify each module has correct data
	metricsA := result["package.moduleA"]
	assert.Equal(t, "package", metricsA.Package)
	assert.Equal(t, 100, metricsA.LinesOfCode)
	assert.Equal(t, 5, metricsA.FunctionCount)
	assert.Equal(t, 2, metricsA.ClassCount)
	assert.Equal(t, []string{"ClassA", "funcA"}, metricsA.PublicInterface)

	metricsB := result["package.moduleB"]
	assert.Equal(t, "package", metricsB.Package)
	assert.Equal(t, 150, metricsB.LinesOfCode)
	assert.Equal(t, 8, metricsB.FunctionCount)
	assert.Equal(t, 3, metricsB.ClassCount)
	assert.Equal(t, []string{"ClassB", "funcB", "helperB"}, metricsB.PublicInterface)

	metricsC := result["package.moduleC"]
	assert.Equal(t, "package", metricsC.Package)
	assert.Equal(t, 75, metricsC.LinesOfCode)
	assert.Equal(t, 3, metricsC.FunctionCount)
	assert.Equal(t, 1, metricsC.ClassCount)
	assert.Equal(t, []string{"ClassC"}, metricsC.PublicInterface)
}

func TestIsTestModule(t *testing.T) {
	service := &SystemAnalysisServiceImpl{}

	testCases := []struct {
		module   string
		expected bool
	}{
		// Test modules - should return true
		{"tests.test_model", true},
		{"test_model", true},
		{"model_test", true},
		{"tests", true},
		{"test", true},
		{"app.testing.fixtures", true},
		{"conftest", true},
		{"app.tests.unit.test_service", true},
		{"test.unit.test_controller", true},
		{"tests.integration.test_api", true},
		{"app.test_user", true},
		{"api_test", true},

		// Non-test modules - should return false
		{"app.domain.models", false},
		{"app.models.user_model", false},
		{"app.services.user_service", false},
		{"app.controllers.api", false},
		{"domain.entities", false},
		{"infrastructure.repository", false},
		{"contest", false},              // Not a test - "conftest" is special, but "contest" is not
		{"app.contestant.model", false}, // Contains "test" but not a test module
	}

	for _, tc := range testCases {
		t.Run(tc.module, func(t *testing.T) {
			result := service.isTestModule(tc.module)
			assert.Equal(t, tc.expected, result, "isTestModule(%q) = %v, expected %v", tc.module, result, tc.expected)
		})
	}
}

func TestDetectLayerFromModule_ExcludesTestModules(t *testing.T) {
	service := &SystemAnalysisServiceImpl{}

	// Standard layer patterns
	patterns := map[string][]string{
		"domain":         {"models", "model", "entities", "entity", "domain", "schemas", "schema"},
		"application":    {"services", "service", "use_cases", "usecase"},
		"infrastructure": {"repositories", "repository", "db", "database"},
		"presentation":   {"controllers", "controller", "api", "views", "router"},
	}

	testCases := []struct {
		module   string
		expected string
	}{
		// Test modules should NOT be classified as any layer (return "")
		{"tests.test_model", ""},
		{"tests.test_entity", ""},
		{"tests.test_schema", ""},
		{"test.unit.test_service", ""},
		{"app.testing.fixtures", ""},
		{"conftest", ""},
		{"tests.integration.test_api", ""},
		{"test_controller", ""},
		{"repository_test", ""},

		// Valid domain modules should still be detected
		{"app.domain.models", "domain"},
		{"app.models.user_model", "domain"},
		{"domain.entities", "domain"},
		{"app.schemas.user_schema", "domain"},

		// Valid application modules should still be detected
		{"app.services.user_service", "application"},
		{"app.use_cases.create_user", "application"},

		// Valid infrastructure modules should still be detected
		{"app.repositories.user_repository", "infrastructure"},
		{"infrastructure.db", "infrastructure"},

		// Valid presentation modules should still be detected
		{"app.api.v1.router", "presentation"},
		{"app.controllers.user_controller", "presentation"},
	}

	for _, tc := range testCases {
		t.Run(tc.module, func(t *testing.T) {
			result := service.detectLayerFromModule(tc.module, patterns)
			assert.Equal(t, tc.expected, result, "detectLayerFromModule(%q) = %q, expected %q", tc.module, result, tc.expected)
		})
	}
}
