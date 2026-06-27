package service

import (
	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"
)

// ArchitectureRulesFromPyscnConfig extracts explicit architecture rules from pyscn config.
func ArchitectureRulesFromPyscnConfig(cfg *config.PyscnConfig) *domain.ArchitectureRules {
	if cfg == nil {
		return nil
	}

	if cfg.ArchitectureStrictMode == nil && cfg.ArchitectureStyle == "" &&
		len(cfg.ArchitectureAllowedPatterns) == 0 && len(cfg.ArchitectureForbiddenPatterns) == 0 &&
		len(cfg.ArchitectureLayers) == 0 && len(cfg.ArchitectureRules) == 0 &&
		len(cfg.ArchitectureNeutralPrefixes) == 0 {
		return nil
	}

	rules := &domain.ArchitectureRules{}
	if cfg.ArchitectureStyle != "" {
		rules.Style = cfg.ArchitectureStyle
	}
	if cfg.ArchitectureStrictMode != nil {
		rules.StrictMode = *cfg.ArchitectureStrictMode
	}
	if len(cfg.ArchitectureAllowedPatterns) > 0 {
		rules.AllowedPatterns = cfg.ArchitectureAllowedPatterns
	}
	if len(cfg.ArchitectureForbiddenPatterns) > 0 {
		rules.ForbiddenPatterns = cfg.ArchitectureForbiddenPatterns
	}
	if len(cfg.ArchitectureLayers) > 0 {
		rules.Layers = convertLayerDefinitions(cfg.ArchitectureLayers)
	}
	if len(cfg.ArchitectureRules) > 0 {
		rules.Rules = convertLayerRules(cfg.ArchitectureRules)
	}
	if len(cfg.ArchitectureNeutralPrefixes) > 0 {
		rules.NeutralPrefixes = cfg.ArchitectureNeutralPrefixes
	}
	return rules
}

// HasExplicitArchitectureConfig reports whether architecture rules were configured by the user.
func HasExplicitArchitectureConfig(rules *domain.ArchitectureRules) bool {
	if rules == nil {
		return false
	}
	return rules.Style != "" ||
		len(rules.Layers) > 0 ||
		len(rules.Rules) > 0 ||
		len(rules.NeutralPrefixes) > 0
}
