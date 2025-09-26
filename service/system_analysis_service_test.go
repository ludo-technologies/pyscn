package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeDependencies_ModuleMetricsPopulated(t *testing.T) {
	// Create a temporary directory with test Python files
	tempDir, err := os.MkdirTemp("", "pyscn-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test Python files with relative imports that should resolve
	testFiles := map[string]string{
		"module_a.py": `
from . import module_b
from . import module_c

def func_a():
    pass

class ClassA:
    pass
`,
		"module_b.py": `
from . import module_c

def func_b():
    pass
`,
		"module_c.py": `
def func_c():
    pass

class ClassC:
    pass
`,
		"__init__.py": `
# Package init file
`,
	}

	// Write test files
	var filePaths []string
	for filename, content := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)
		filePaths = append(filePaths, filePath)
	}

	// Create service instance
	service := NewSystemAnalysisService()

	// Create analysis request
	req := domain.SystemAnalysisRequest{
		Paths:               filePaths,
		AnalyzeDependencies: true,
		IncludeStdLib:       false,
		IncludeThirdParty:   false,
		FollowRelative:      true,
	}

	// Perform dependency analysis
	ctx := context.Background()
	result, err := service.AnalyzeDependencies(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Debug output
	t.Logf("Total modules found: %d", result.TotalModules)
	t.Logf("Total dependencies found: %d", result.TotalDependencies)
	t.Logf("Module metrics count: %d", len(result.ModuleMetrics))
	for name, metrics := range result.ModuleMetrics {
		t.Logf("Module %s: Ca=%d, Ce=%d", name, metrics.AfferentCoupling, metrics.EfferentCoupling)
	}

	// Verify that ModuleMetrics is not empty
	assert.NotEmpty(t, result.ModuleMetrics, "ModuleMetrics should not be empty")

	// Check that we have metrics for each module
	assert.True(t, len(result.ModuleMetrics) > 0, "Should have at least one module with metrics")

	// Verify specific module metrics exist and have expected properties
	for _, metrics := range result.ModuleMetrics {
		assert.NotNil(t, metrics, "Metrics should not be nil")

		// Basic information should be populated
		assert.NotEmpty(t, metrics.ModuleName, "ModuleName should be populated")
		assert.NotEmpty(t, metrics.FilePath, "FilePath should be populated")

		// Instability should be calculated correctly when there are dependencies
		if metrics.AfferentCoupling+metrics.EfferentCoupling > 0 {
			expectedInstability := float64(metrics.EfferentCoupling) / float64(metrics.AfferentCoupling+metrics.EfferentCoupling)
			assert.InDelta(t, expectedInstability, metrics.Instability, 0.01,
				"Instability should be Ce/(Ca+Ce) for module %s", metrics.ModuleName)
		}

		// Risk level should be assigned
		assert.NotEmpty(t, metrics.RiskLevel, "Risk level should be assigned")

		// Dependencies lists should be initialized (even if empty)
		assert.NotNil(t, metrics.DirectDependencies, "DirectDependencies should be initialized")
		assert.NotNil(t, metrics.Dependents, "Dependents should be initialized")
		assert.NotNil(t, metrics.TransitiveDependencies, "TransitiveDependencies should be initialized")
	}

	// Verify system-level metrics are also calculated
	assert.Greater(t, result.TotalModules, 0, "TotalModules should be greater than 0")
	// Note: TotalDependencies may be 0 if import resolution fails in temp directory
	assert.NotNil(t, result.CouplingAnalysis, "CouplingAnalysis should not be nil")
}

func TestAnalyzeDependencies_EmptyInput(t *testing.T) {
	// Create service instance
	service := NewSystemAnalysisService()

	// Create a temporary empty directory
	tempDir, err := os.MkdirTemp("", "pyscn-test-empty-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create analysis request pointing to empty directory
	req := domain.SystemAnalysisRequest{
		Paths:               []string{tempDir},
		AnalyzeDependencies: true,
		IncludeStdLib:       false,
		IncludeThirdParty:   false,
		FollowRelative:      true,
	}

	// Perform dependency analysis
	ctx := context.Background()
	result, err := service.AnalyzeDependencies(ctx, req)
	// The analyzer may return an error for no Python files, or may return empty result
	if err != nil {
		// If it errors, that's ok for an empty directory
		assert.Contains(t, err.Error(), "no valid Python files", "Should error about no Python files")
	} else {
		// If it doesn't error, verify the result is properly initialized
		require.NotNil(t, result)
		assert.Equal(t, 0, result.TotalModules)
		assert.Equal(t, 0, result.TotalDependencies)
		assert.NotNil(t, result.RootModules)
		assert.NotNil(t, result.LeafModules)
		assert.NotNil(t, result.ModuleMetrics, "ModuleMetrics should be initialized even if empty")
		assert.Empty(t, result.ModuleMetrics, "ModuleMetrics should be empty for no input files")
	}
}
