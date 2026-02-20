package config

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/pelletier/go-toml/v2"
)

// defaultConfigTmpl contains the embedded default configuration template
//
//go:embed default_config.toml.tmpl
var defaultConfigTmpl string

// DefaultConfigValues holds all values used to render the default config template.
// All values are sourced from the domain package to ensure a single source of truth.
type DefaultConfigValues struct {
	// Complexity
	ComplexityLowThreshold         int
	ComplexityMediumThreshold      int
	ComplexityLowThresholdPlus1    int
	ComplexityMediumThresholdPlus1 int
	ComplexityMinFilter            int
	ComplexityMaxLimit             int

	// Dead Code
	DeadCodeMinSeverity  string
	DeadCodeContextLines int
	DeadCodeSortBy       string

	// Clone thresholds
	Type1Threshold       float64
	Type2Threshold       float64
	Type3Threshold       float64
	Type4Threshold       float64
	SimilarityThreshold  float64
	GroupingThreshold    float64
	CloneMinLines        int
	CloneMinNodes        int
	CloneMaxEditDistance  float64

	// LSH
	LSHAutoThreshold      int
	LSHSimilarityThreshold float64
	LSHBands              int
	LSHRows               int
	LSHHashes             int

	// Performance
	MaxMemoryMB    int
	BatchSize      int
	MaxGoroutines  int
	TimeoutSeconds int

	// CBO
	CBOLowThreshold         int
	CBOMediumThreshold      int
	CBOLowThresholdPlus1    int
}

// newDefaultConfigValues creates a DefaultConfigValues populated from domain constants.
func newDefaultConfigValues() DefaultConfigValues {
	return DefaultConfigValues{
		// Complexity
		ComplexityLowThreshold:         domain.DefaultComplexityLowThreshold,
		ComplexityMediumThreshold:      domain.DefaultComplexityMediumThreshold,
		ComplexityLowThresholdPlus1:    domain.DefaultComplexityLowThreshold + 1,
		ComplexityMediumThresholdPlus1: domain.DefaultComplexityMediumThreshold + 1,
		ComplexityMinFilter:            domain.DefaultComplexityMinFilter,
		ComplexityMaxLimit:             domain.DefaultComplexityMaxLimit,

		// Dead Code
		DeadCodeMinSeverity:  domain.DefaultDeadCodeMinSeverity,
		DeadCodeContextLines: domain.DefaultDeadCodeContextLines,
		DeadCodeSortBy:       domain.DefaultDeadCodeSortBy,

		// Clone thresholds
		Type1Threshold:      domain.DefaultType1CloneThreshold,
		Type2Threshold:      domain.DefaultType2CloneThreshold,
		Type3Threshold:      domain.DefaultType3CloneThreshold,
		Type4Threshold:      domain.DefaultType4CloneThreshold,
		SimilarityThreshold: domain.DefaultCloneSimilarityThreshold,
		GroupingThreshold:   domain.DefaultCloneGroupingThreshold,
		CloneMinLines:       domain.DefaultCloneMinLines,
		CloneMinNodes:       domain.DefaultCloneMinNodes,
		CloneMaxEditDistance: domain.DefaultCloneMaxEditDistance,

		// LSH
		LSHAutoThreshold:       domain.DefaultLSHAutoThreshold,
		LSHSimilarityThreshold: domain.DefaultLSHSimilarityThreshold,
		LSHBands:               domain.DefaultLSHBands,
		LSHRows:                domain.DefaultLSHRows,
		LSHHashes:              domain.DefaultLSHHashes,

		// Performance
		MaxMemoryMB:    domain.DefaultMaxMemoryMB,
		BatchSize:      domain.DefaultBatchSize,
		MaxGoroutines:  domain.DefaultMaxGoroutines,
		TimeoutSeconds: domain.DefaultTimeoutSeconds,

		// CBO
		CBOLowThreshold:      domain.DefaultCBOLowThreshold,
		CBOMediumThreshold:   domain.DefaultCBOMediumThreshold,
		CBOLowThresholdPlus1: domain.DefaultCBOLowThreshold + 1,
	}
}

// GenerateDefaultConfigTOML renders the default config template with domain values
// and returns the resulting TOML string.
func GenerateDefaultConfigTOML() (string, error) {
	tmpl, err := template.New("default_config").Parse(defaultConfigTmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse default config template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, newDefaultConfigValues()); err != nil {
		return "", fmt.Errorf("failed to render default config template: %w", err)
	}

	return buf.String(), nil
}

// LoadDefaultConfigFromTOML parses the embedded default config and returns the full Config struct
func LoadDefaultConfigFromTOML() (*Config, error) {
	configTOML, err := GenerateDefaultConfigTOML()
	if err != nil {
		return nil, err
	}

	var tomlCfg PyscnTomlConfig
	if err := toml.Unmarshal([]byte(configTOML), &tomlCfg); err != nil {
		return nil, err
	}

	// Merge with defaults using the standard conversion
	defaults := DefaultPyscnConfig()
	loader := &TomlConfigLoader{}
	loader.mergePyscnTomlConfigs(defaults, &tomlCfg)

	cfg := PyscnConfigToConfig(defaults)

	// Copy architecture layers and rules from TOML config
	if len(tomlCfg.Architecture.Layers) > 0 {
		cfg.Architecture.Layers = make([]LayerDefinition, len(tomlCfg.Architecture.Layers))
		for i, layer := range tomlCfg.Architecture.Layers {
			cfg.Architecture.Layers[i] = LayerDefinition{
				Name:        layer.Name,
				Description: layer.Description,
				Packages:    layer.Packages,
			}
		}
	}
	if len(tomlCfg.Architecture.Rules) > 0 {
		cfg.Architecture.Rules = make([]LayerRule, len(tomlCfg.Architecture.Rules))
		for i, rule := range tomlCfg.Architecture.Rules {
			cfg.Architecture.Rules[i] = LayerRule{
				From:  rule.From,
				Allow: rule.Allow,
				Deny:  rule.Deny,
			}
		}
	}

	return cfg, nil
}

// LoadDefaultConfigTOMLString returns the rendered default config as a string
// This can be used to display to users
func LoadDefaultConfigTOMLString() (string, error) {
	return GenerateDefaultConfigTOML()
}
