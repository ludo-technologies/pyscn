package config

import (
	_ "embed"
	"strings"

	"github.com/spf13/viper"
)

// DefaultConfigTOML contains the embedded default configuration file
//
//go:embed default_config.toml
var DefaultConfigTOML string

// LoadDefaultConfig parses the embedded default config and returns the full Config struct
func LoadDefaultConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigType("toml")

	if err := v.ReadConfig(strings.NewReader(DefaultConfigTOML)); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
