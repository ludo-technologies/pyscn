package service

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigurationLoader(t *testing.T) {
	loader := NewConfigurationLoader()
	assert.NotNil(t, loader)
}

func TestConfigurationLoader_LoadDefaultConfig(t *testing.T) {
	loader := NewConfigurationLoader()

	req := loader.LoadDefaultConfig()
	require.NotNil(t, req)

	// Verify default values are set
	assert.Equal(t, domain.DefaultComplexityMinFilter, req.MinComplexity)
	assert.Equal(t, domain.DefaultComplexityLowThreshold, req.LowThreshold)
	assert.Equal(t, domain.DefaultComplexityMediumThreshold, req.MediumThreshold)
	require.NotNil(t, req.Enabled)
	require.NotNil(t, req.ReportUnchanged)
	assert.True(t, *req.Enabled)
	assert.True(t, *req.ReportUnchanged)
}

func TestConfigurationLoader_MergeConfig(t *testing.T) {
	loader := NewConfigurationLoader()

	base := &domain.ComplexityRequest{
		Paths:           []string{"base/path"},
		MinComplexity:   1,
		MaxComplexity:   0,
		LowThreshold:    10,
		MediumThreshold: 20,
		ShowDetails:     false,
		Recursive:       true,
	}

	override := &domain.ComplexityRequest{
		Paths:         []string{"override/path"},
		MinComplexity: 5,
		MaxComplexity: 15,
		ShowDetails:   true,
	}

	merged := loader.MergeConfig(base, override)

	assert.Equal(t, []string{"override/path"}, merged.Paths)
	assert.Equal(t, 5, merged.MinComplexity)
	assert.Equal(t, 15, merged.MaxComplexity)
	assert.True(t, merged.ShowDetails)
	// Base values should be preserved when not overridden
	assert.Equal(t, 10, merged.LowThreshold)
	assert.Equal(t, 20, merged.MediumThreshold)
}

func TestConfigurationLoader_MergeConfig_OutputFormat(t *testing.T) {
	loader := NewConfigurationLoader()

	base := &domain.ComplexityRequest{
		OutputFormat: domain.OutputFormatText,
	}

	override := &domain.ComplexityRequest{
		OutputFormat: domain.OutputFormatJSON,
	}

	merged := loader.MergeConfig(base, override)
	assert.Equal(t, domain.OutputFormatJSON, merged.OutputFormat)
}

func TestConfigurationLoader_MergeConfig_OutputWriter(t *testing.T) {
	loader := NewConfigurationLoader()

	base := &domain.ComplexityRequest{}
	var buf bytes.Buffer
	override := &domain.ComplexityRequest{
		OutputWriter: &buf,
	}

	merged := loader.MergeConfig(base, override)
	assert.Equal(t, &buf, merged.OutputWriter)
}

func TestConfigurationLoader_MergeConfig_Thresholds(t *testing.T) {
	loader := NewConfigurationLoader()

	base := &domain.ComplexityRequest{
		LowThreshold:    10,
		MediumThreshold: 20,
	}

	override := &domain.ComplexityRequest{
		LowThreshold:    15,
		MediumThreshold: 25,
	}

	merged := loader.MergeConfig(base, override)
	assert.Equal(t, 15, merged.LowThreshold)
	assert.Equal(t, 25, merged.MediumThreshold)
}

// TestConfigurationLoader_MergeConfig_ThresholdOverrideMatchesDefault is a
// regression test for issue #553: an explicit override that happens to equal
// the domain default must still win over the base value.
func TestConfigurationLoader_MergeConfig_ThresholdOverrideMatchesDefault(t *testing.T) {
	loader := NewConfigurationLoader()

	base := &domain.ComplexityRequest{
		LowThreshold:                 10,
		MediumThreshold:              20,
		CognitiveComplexityThreshold: 30,
		NestingDepthThreshold:        11,
	}

	override := &domain.ComplexityRequest{
		LowThreshold:                 domain.DefaultComplexityLowThreshold,       // 9
		MediumThreshold:              domain.DefaultComplexityMediumThreshold,    // 19
		CognitiveComplexityThreshold: domain.DefaultCognitiveComplexityThreshold, // 25
		NestingDepthThreshold:        domain.DefaultNestingDepthThreshold,        // 7
	}

	merged := loader.MergeConfig(base, override)
	assert.Equal(t, domain.DefaultComplexityLowThreshold, merged.LowThreshold,
		"explicit override matching default should win over base")
	assert.Equal(t, domain.DefaultComplexityMediumThreshold, merged.MediumThreshold,
		"explicit override matching default should win over base")
	assert.Equal(t, domain.DefaultCognitiveComplexityThreshold, merged.CognitiveComplexityThreshold,
		"explicit override matching default should win over base")
	assert.Equal(t, domain.DefaultNestingDepthThreshold, merged.NestingDepthThreshold,
		"explicit override matching default should win over base")
}

// TestConfigurationLoader_MergeConfig_ThresholdZeroPreservesBase verifies that
// a zero (unset) override leaves the base value intact.
func TestConfigurationLoader_MergeConfig_ThresholdZeroPreservesBase(t *testing.T) {
	loader := NewConfigurationLoader()

	base := &domain.ComplexityRequest{
		LowThreshold:                 10,
		MediumThreshold:              20,
		CognitiveComplexityThreshold: 30,
		NestingDepthThreshold:        11,
	}

	override := &domain.ComplexityRequest{}

	merged := loader.MergeConfig(base, override)
	assert.Equal(t, 10, merged.LowThreshold)
	assert.Equal(t, 20, merged.MediumThreshold)
	assert.Equal(t, 30, merged.CognitiveComplexityThreshold)
	assert.Equal(t, 11, merged.NestingDepthThreshold)
}

// TestConfigurationLoader_MergeConfig_MinComplexityOverrideMatchesDefault is a
// regression test: an explicit MinComplexity override equal to the default (1)
// must still win over the base value. The old guard (!= 1) dropped it.
func TestConfigurationLoader_MergeConfig_MinComplexityOverrideMatchesDefault(t *testing.T) {
	loader := NewConfigurationLoader()

	base := &domain.ComplexityRequest{MinComplexity: 5}
	override := &domain.ComplexityRequest{MinComplexity: domain.DefaultComplexityMinFilter}

	merged := loader.MergeConfig(base, override)
	assert.Equal(t, domain.DefaultComplexityMinFilter, merged.MinComplexity,
		"explicit MinComplexity override matching default should win over base")
}

// TestConfigurationLoader_MergeConfig_SortByOverrideMatchesDefault is a
// regression test: an explicit SortBy override equal to the default
// (complexity) must still win over the base value. The old guard
// (!= "complexity") dropped it.
func TestConfigurationLoader_MergeConfig_SortByOverrideMatchesDefault(t *testing.T) {
	loader := NewConfigurationLoader()

	base := &domain.ComplexityRequest{SortBy: domain.SortByName}
	override := &domain.ComplexityRequest{SortBy: domain.SortByComplexity}

	merged := loader.MergeConfig(base, override)
	assert.Equal(t, domain.SortByComplexity, merged.SortBy,
		"explicit SortBy override matching default should win over base")
}

// TestConfigurationLoader_MergeConfig_ZeroOverridePreservesBase verifies that
// unset (zero-value) override fields leave the base values intact.
func TestConfigurationLoader_MergeConfig_ZeroOverridePreservesBase(t *testing.T) {
	loader := NewConfigurationLoader()

	base := &domain.ComplexityRequest{
		MinComplexity: 5,
		SortBy:        domain.SortByName,
	}
	override := &domain.ComplexityRequest{}

	merged := loader.MergeConfig(base, override)
	assert.Equal(t, 5, merged.MinComplexity)
	assert.Equal(t, domain.SortByName, merged.SortBy)
}

func TestConfigurationLoader_MergeConfig_Patterns(t *testing.T) {
	loader := NewConfigurationLoader()

	base := &domain.ComplexityRequest{
		IncludePatterns: []string{"*.py"},
		ExcludePatterns: []string{"test_*.py"},
	}

	override := &domain.ComplexityRequest{
		IncludePatterns: []string{"**/*.py"},
		ExcludePatterns: []string{"*_test.py"},
	}

	merged := loader.MergeConfig(base, override)
	assert.Equal(t, []string{"**/*.py"}, merged.IncludePatterns)
	assert.Equal(t, []string{"*_test.py"}, merged.ExcludePatterns)
}

func TestConfigurationLoader_LoadConfig_ValidFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".pyscn.toml")

	configContent := `
	[complexity]
	enabled = false
	report_unchanged = false
	low_threshold = 10
	medium_threshold = 15
	`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	loader := NewConfigurationLoader()
	req, err := loader.LoadConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, req)
	require.NotNil(t, req.Enabled)
	require.NotNil(t, req.ReportUnchanged)
	assert.False(t, *req.Enabled)
	assert.False(t, *req.ReportUnchanged)
}

func TestConfigurationLoader_LoadConfig_MinComplexityPrecedence(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".pyscn.toml")

	configContent := `
	[complexity]
	min_complexity = 3

	[output]
	min_complexity = 1
	`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	loader := NewConfigurationLoader()
	req, err := loader.LoadConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, req)
	assert.Equal(t, 1, req.MinComplexity)
}

func TestConfigurationLoader_FindDefaultConfigFile(t *testing.T) {
	loader := NewConfigurationLoader()

	// In a directory without config file
	result := loader.FindDefaultConfigFile()
	// Result depends on current working directory
	// Just verify it returns a string (empty or path)
	assert.IsType(t, "", result)
}

func TestConfigurationLoader_ValidateConfig(t *testing.T) {
	loader := NewConfigurationLoader()

	tests := []struct {
		name      string
		config    *domain.ComplexityRequest
		expectErr bool
	}{
		{
			name: "valid config",
			config: &domain.ComplexityRequest{
				MinComplexity:   1,
				MaxComplexity:   20,
				LowThreshold:    10,
				MediumThreshold: 15,
			},
			expectErr: false,
		},
		{
			name: "min greater than max",
			config: &domain.ComplexityRequest{
				MinComplexity: 20,
				MaxComplexity: 10,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loader.ValidateConfig(tt.config)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigurationLoader_GetDefaultThresholds(t *testing.T) {
	loader := NewConfigurationLoader()

	low, medium := loader.GetDefaultThresholds()

	assert.Equal(t, domain.DefaultComplexityLowThreshold, low)
	assert.Equal(t, domain.DefaultComplexityMediumThreshold, medium)
}

func TestConfigurationLoader_CreateConfigTemplate(t *testing.T) {
	loader := NewConfigurationLoader()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "template.toml")

	err := loader.CreateConfigTemplate(configPath)
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(configPath)
	assert.NoError(t, err)
}
