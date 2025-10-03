package service

import (
	"fmt"
	"testing"

	"github.com/ludo-technologies/pyscn/internal/analyzer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractCouplingResult_VariesWithGraphs(t *testing.T) {
	tests := []struct {
		name          string
		setupGraph    func() *analyzer.DependencyGraph
		expectedCheck func(t *testing.T, metrics *analyzer.SystemMetrics)
	}{
		{
			name: "simple_graph_with_low_coupling",
			setupGraph: func() *analyzer.DependencyGraph {
				graph := analyzer.NewDependencyGraph("/test/project")

				// Create a simple linear dependency chain: A -> B -> C
				graph.AddModule("moduleA", "/test/moduleA.py")
				graph.AddModule("moduleB", "/test/moduleB.py")
				graph.AddModule("moduleC", "/test/moduleC.py")

				graph.AddDependency("moduleA", "moduleB", analyzer.DependencyEdgeImport, nil)
				graph.AddDependency("moduleB", "moduleC", analyzer.DependencyEdgeImport, nil)

				// Calculate metrics
				calculator := analyzer.NewCouplingMetricsCalculator(graph, analyzer.DefaultCouplingMetricsOptions())
				err := calculator.CalculateMetrics()
				require.NoError(t, err)

				return graph
			},
			expectedCheck: func(t *testing.T, metrics *analyzer.SystemMetrics) {
				assert.Equal(t, 3, metrics.TotalModules)
				assert.Equal(t, 2, metrics.TotalDependencies)
				assert.Equal(t, 0, metrics.CyclicDependencies)          // No cycles
				assert.InDelta(t, 0.667, metrics.DependencyRatio, 0.01) // 2/3
				assert.NotNil(t, metrics.RefactoringPriority)
			},
		},
		{
			name: "complex_graph_with_high_coupling",
			setupGraph: func() *analyzer.DependencyGraph {
				graph := analyzer.NewDependencyGraph("/test/project")

				// Create a highly coupled graph with cycles
				modules := []string{"core", "utils", "api", "db", "auth"}
				for _, m := range modules {
					graph.AddModule(m, "/test/"+m+".py")
				}

				// Create many dependencies including cycles
				graph.AddDependency("core", "utils", analyzer.DependencyEdgeImport, nil)
				graph.AddDependency("core", "api", analyzer.DependencyEdgeImport, nil)
				graph.AddDependency("core", "db", analyzer.DependencyEdgeImport, nil)
				graph.AddDependency("api", "auth", analyzer.DependencyEdgeImport, nil)
				graph.AddDependency("api", "db", analyzer.DependencyEdgeImport, nil)
				graph.AddDependency("auth", "db", analyzer.DependencyEdgeImport, nil)
				graph.AddDependency("auth", "utils", analyzer.DependencyEdgeImport, nil)
				graph.AddDependency("db", "utils", analyzer.DependencyEdgeImport, nil)
				graph.AddDependency("utils", "core", analyzer.DependencyEdgeImport, nil) // Creates cycle

				// Detect cycles
				detector := analyzer.NewCircularDependencyDetector(graph)
				detector.DetectCircularDependencies()

				// Calculate metrics
				calculator := analyzer.NewCouplingMetricsCalculator(graph, analyzer.DefaultCouplingMetricsOptions())
				err := calculator.CalculateMetrics()
				require.NoError(t, err)

				return graph
			},
			expectedCheck: func(t *testing.T, metrics *analyzer.SystemMetrics) {
				assert.Equal(t, 5, metrics.TotalModules)
				assert.Equal(t, 9, metrics.TotalDependencies)
				assert.Greater(t, metrics.CyclicDependencies, 0)      // Has cycles
				assert.InDelta(t, 1.8, metrics.DependencyRatio, 0.01) // 9/5
				assert.Greater(t, metrics.SystemComplexity, 0.0)
				assert.NotNil(t, metrics.RefactoringPriority)
			},
		},
		{
			name: "graph_with_packages",
			setupGraph: func() *analyzer.DependencyGraph {
				graph := analyzer.NewDependencyGraph("/test/project")

				// Create modules in different packages
				modA1 := graph.AddModule("packageA.module1", "/test/packageA/module1.py")
				modA1.Package = "packageA"
				modA2 := graph.AddModule("packageA.module2", "/test/packageA/module2.py")
				modA2.Package = "packageA"

				modB1 := graph.AddModule("packageB.module1", "/test/packageB/module1.py")
				modB1.Package = "packageB"
				modB2 := graph.AddModule("packageB.module2", "/test/packageB/module2.py")
				modB2.Package = "packageB"

				// Intra-package dependencies (good)
				graph.AddDependency("packageA.module1", "packageA.module2", analyzer.DependencyEdgeImport, nil)
				graph.AddDependency("packageB.module1", "packageB.module2", analyzer.DependencyEdgeImport, nil)

				// Inter-package dependency (less good)
				graph.AddDependency("packageA.module1", "packageB.module1", analyzer.DependencyEdgeImport, nil)

				// Calculate metrics
				calculator := analyzer.NewCouplingMetricsCalculator(graph, analyzer.DefaultCouplingMetricsOptions())
				err := calculator.CalculateMetrics()
				require.NoError(t, err)

				return graph
			},
			expectedCheck: func(t *testing.T, metrics *analyzer.SystemMetrics) {
				assert.Equal(t, 4, metrics.TotalModules)
				assert.Equal(t, 3, metrics.TotalDependencies)
				assert.Equal(t, 2, metrics.PackageCount)
				assert.Greater(t, metrics.ModularityIndex, 0.0)        // Should have some modularity
				assert.InDelta(t, 0.75, metrics.DependencyRatio, 0.01) // 3/4
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			service := &SystemAnalysisServiceImpl{}
			graph := tt.setupGraph()

			// Extract metrics
			result := service.extractCouplingResult(graph)

			// Verify metrics
			require.NotNil(t, result)
			tt.expectedCheck(t, result)
		})
	}
}

func TestExtractCouplingResult_UsesCalculatedMetrics(t *testing.T) {
	service := &SystemAnalysisServiceImpl{}
	graph := analyzer.NewDependencyGraph("/test/project")

	// Add some modules
	graph.AddModule("module1", "/test/module1.py")
	graph.AddModule("module2", "/test/module2.py")
	graph.AddModule("module3", "/test/module3.py")

	// Add dependencies
	graph.AddDependency("module1", "module2", analyzer.DependencyEdgeImport, nil)
	graph.AddDependency("module2", "module3", analyzer.DependencyEdgeImport, nil)

	// Calculate metrics using CouplingMetricsCalculator
	calculator := analyzer.NewCouplingMetricsCalculator(graph, analyzer.DefaultCouplingMetricsOptions())
	err := calculator.CalculateMetrics()
	require.NoError(t, err)

	// Extract results
	result := service.extractCouplingResult(graph)

	// Verify it uses the calculated SystemMetrics
	assert.Equal(t, graph.SystemMetrics, result)
	assert.NotNil(t, result)
	assert.Equal(t, 3, result.TotalModules)
	assert.Equal(t, 2, result.TotalDependencies)

	// Verify module metrics were used in calculation
	assert.NotNil(t, graph.ModuleMetrics)
	assert.Len(t, graph.ModuleMetrics, 3)
}

func TestExtractCouplingResult_DifferentGraphsProduceDifferentMetrics(t *testing.T) {
	service := &SystemAnalysisServiceImpl{}

	// Graph 1: Low complexity
	graph1 := analyzer.NewDependencyGraph("/test/project1")
	graph1.AddModule("simple", "/test/simple.py")
	calculator1 := analyzer.NewCouplingMetricsCalculator(graph1, analyzer.DefaultCouplingMetricsOptions())
	err := calculator1.CalculateMetrics()
	require.NoError(t, err)
	metrics1 := service.extractCouplingResult(graph1)

	// Graph 2: Higher complexity
	graph2 := analyzer.NewDependencyGraph("/test/project2")
	for i := 0; i < 10; i++ {
		moduleName := fmt.Sprintf("module%d", i)
		graph2.AddModule(moduleName, fmt.Sprintf("/test/%s.py", moduleName))
		if i > 0 {
			// Create chain of dependencies
			graph2.AddDependency(fmt.Sprintf("module%d", i-1), moduleName, analyzer.DependencyEdgeImport, nil)
		}
	}
	calculator2 := analyzer.NewCouplingMetricsCalculator(graph2, analyzer.DefaultCouplingMetricsOptions())
	err = calculator2.CalculateMetrics()
	require.NoError(t, err)
	metrics2 := service.extractCouplingResult(graph2)

	// Verify metrics are different
	assert.NotEqual(t, metrics1.TotalModules, metrics2.TotalModules)
	assert.NotEqual(t, metrics1.TotalDependencies, metrics2.TotalDependencies)
	assert.NotEqual(t, metrics1.DependencyRatio, metrics2.DependencyRatio)
	assert.NotEqual(t, metrics1.SystemComplexity, metrics2.SystemComplexity)

	// Graph 2 should have higher complexity
	assert.Greater(t, metrics2.TotalModules, metrics1.TotalModules)
	assert.Greater(t, metrics2.TotalDependencies, metrics1.TotalDependencies)
	assert.Greater(t, metrics2.SystemComplexity, metrics1.SystemComplexity)
	assert.GreaterOrEqual(t, metrics2.MaxDependencyDepth, metrics1.MaxDependencyDepth)
}
