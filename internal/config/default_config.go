package config

import (
	_ "embed"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// DefaultConfigTOML contains the embedded default configuration file
//
//go:embed default_config.toml
var DefaultConfigTOML string

// LoadDefaultConfigFromTOML parses the embedded default config and returns the full Config struct
func LoadDefaultConfigFromTOML() (*Config, error) {
	var tomlCfg PyscnTomlConfig
	if err := toml.Unmarshal([]byte(DefaultConfigTOML), &tomlCfg); err != nil {
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

// LoadDefaultConfigTOMLString returns the embedded default config as a string
// This can be used to parse the TOML directly or display to users
func LoadDefaultConfigTOMLString() string {
	return strings.TrimSpace(DefaultConfigTOML)
}
