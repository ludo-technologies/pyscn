package config

// Architecture style preset names. Users select one via `style` in the
// [architecture] section of pyscn.toml / pyproject.toml to auto-apply a
// matching set of layers and dependency rules.
const (
	// ArchitectureStyleLayered is the default, backward-compatible preset.
	// An empty style is treated as layered.
	ArchitectureStyleLayered = "layered"
)

// ArchitectureStylePreset returns the layer definitions and dependency rules
// for the named architecture style. An empty style is treated as "layered".
// Returns (nil, nil) for an unrecognized style so callers can fall back to
// auto-detection.
func ArchitectureStylePreset(style string) ([]LayerDefinition, []LayerRule) {
	switch style {
	case "", ArchitectureStyleLayered:
		return layeredPresetLayers(), layeredPresetRules()
	default:
		return nil, nil
	}
}

// layeredPresetLayers mirrors the [architecture.layers] section of
// internal/config/default_config.toml.tmpl. Kept identical to preserve the
// behavior projects had before style presets existed.
func layeredPresetLayers() []LayerDefinition {
	return []LayerDefinition{
		{
			Name:     "presentation",
			Packages: []string{"router", "routers", "route", "routes", "endpoint", "endpoints", "handler", "handlers", "controller", "controllers", "view", "views", "api", "apis", "ui", "web", "rest", "graphql"},
		},
		{
			Name:     "application",
			Packages: []string{"service", "services", "usecase", "usecases", "use_case", "use_cases", "workflow", "workflows", "command", "commands", "query", "queries", "manager", "managers"},
		},
		{
			Name:     "domain",
			Packages: []string{"model", "models", "entity", "entities", "schema", "schemas", "domain", "domains", "core", "business", "aggregate", "aggregates", "valueobject", "valueobjects"},
		},
		{
			Name:     "infrastructure",
			Packages: []string{"repository", "repositories", "repo", "repos", "db", "database", "adapter", "adapters", "persistence", "storage", "cache", "client", "clients", "external"},
		},
	}
}

// layeredPresetRules mirrors the [architecture.rules] section of
// internal/config/default_config.toml.tmpl.
//
// Note: this preset intentionally allows domain -> infrastructure. That relaxes
// the strict Dependency Inversion Principle, but it is the long-standing default
// behavior and is preserved here for backward compatibility. Use the hexagonal
// or clean presets for stricter DIP enforcement.
func layeredPresetRules() []LayerRule {
	return []LayerRule{
		{
			From:  "presentation",
			Allow: []string{"presentation", "application", "domain", "infrastructure"},
		},
		{
			From:  "application",
			Allow: []string{"application", "domain", "infrastructure"},
		},
		{
			From:  "domain",
			Allow: []string{"domain", "infrastructure"},
			Deny:  []string{"presentation", "application"},
		},
		{
			From:  "infrastructure",
			Allow: []string{"infrastructure", "domain", "application"},
		},
	}
}
