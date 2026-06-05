package config

// Architecture style preset names. Users select one via `style` in the
// [architecture] section of pyscn.toml / pyproject.toml to auto-apply a
// matching set of layers and dependency rules.
const (
	// ArchitectureStyleLayered is the default, backward-compatible preset.
	// An empty style is treated as layered.
	ArchitectureStyleLayered = "layered"

	// ArchitectureStyleHexagonal enforces Hexagonal / Onion architecture: the
	// domain has no outward dependencies (Dependency Inversion).
	ArchitectureStyleHexagonal = "hexagonal"

	// ArchitectureStyleClean enforces Clean Architecture: inner layers never
	// depend on outer ones.
	ArchitectureStyleClean = "clean"

	// ArchitectureStyleMVC enforces MVC/MVT: the view may not depend directly on
	// the model (such a dependency is discouraged and emits a warning).
	ArchitectureStyleMVC = "mvc"
)

// ArchitectureStylePreset returns the layer definitions and dependency rules
// for the named architecture style. An empty style is treated as "layered".
// Returns (nil, nil) for an unrecognized style so callers can fall back to
// auto-detection.
func ArchitectureStylePreset(style string) ([]LayerDefinition, []LayerRule) {
	switch style {
	case "", ArchitectureStyleLayered:
		return layeredPresetLayers(), layeredPresetRules()
	case ArchitectureStyleHexagonal:
		return hexagonalPresetLayers(), hexagonalPresetRules()
	case ArchitectureStyleClean:
		return cleanPresetLayers(), cleanPresetRules()
	case ArchitectureStyleMVC:
		return mvcPresetLayers(), mvcPresetRules()
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

// hexagonalPresetLayers defines the layers for Hexagonal / Onion architecture.
// Ports (interfaces) live on the domain side; adapters (implementations) live on
// the infrastructure side.
func hexagonalPresetLayers() []LayerDefinition {
	return []LayerDefinition{
		{
			Name:     "domain",
			Packages: []string{"domain", "domains", "model", "models", "entity", "entities", "core", "business", "aggregate", "aggregates", "valueobject", "valueobjects"},
		},
		{
			Name:     "ports",
			Packages: []string{"port", "ports", "interface", "interfaces", "contract", "contracts", "usecase", "usecases", "use_case", "use_cases", "service", "services"},
		},
		{
			Name:     "adapters",
			Packages: []string{"adapter", "adapters", "handler", "handlers", "controller", "controllers", "router", "routers", "api", "apis", "repository", "repositories", "repo", "repos", "db", "database", "persistence", "storage", "client", "clients", "external", "web", "rest", "graphql"},
		},
	}
}

// hexagonalPresetRules enforces the Dependency Inversion Principle: the domain
// depends on nothing outward, ports depend only on the domain, and adapters may
// depend on ports and the domain.
func hexagonalPresetRules() []LayerRule {
	return []LayerRule{
		{
			From:  "domain",
			Allow: []string{"domain"},
			Deny:  []string{"ports", "adapters"},
		},
		{
			From:  "ports",
			Allow: []string{"ports", "domain"},
			Deny:  []string{"adapters"},
		},
		{
			From:  "adapters",
			Allow: []string{"adapters", "ports", "domain"},
		},
	}
}

// cleanPresetLayers defines the four concentric layers of Clean Architecture,
// from innermost (entities) to outermost (frameworks).
func cleanPresetLayers() []LayerDefinition {
	return []LayerDefinition{
		{
			Name:     "entities",
			Packages: []string{"entity", "entities", "model", "models", "domain", "domains", "core", "aggregate", "aggregates", "valueobject", "valueobjects"},
		},
		{
			Name:     "use_cases",
			Packages: []string{"usecase", "usecases", "use_case", "use_cases", "interactor", "interactors", "service", "services", "workflow", "workflows", "command", "commands", "query", "queries"},
		},
		{
			Name:     "interface_adapters",
			Packages: []string{"interface_adapter", "interface_adapters", "adapter", "adapters", "controller", "controllers", "presenter", "presenters", "gateway", "gateways", "repository", "repositories", "repo", "repos", "router", "routers"},
		},
		{
			Name:     "frameworks",
			Packages: []string{"framework", "frameworks", "infrastructure", "db", "database", "web", "api", "apis", "external", "client", "clients", "persistence", "storage", "ui"},
		},
	}
}

// cleanPresetRules enforces the dependency rule of Clean Architecture: source
// code dependencies point only inward. Each layer may depend on itself and any
// inner layer, never an outer one.
func cleanPresetRules() []LayerRule {
	return []LayerRule{
		{
			From:  "entities",
			Allow: []string{"entities"},
			Deny:  []string{"use_cases", "interface_adapters", "frameworks"},
		},
		{
			From:  "use_cases",
			Allow: []string{"use_cases", "entities"},
			Deny:  []string{"interface_adapters", "frameworks"},
		},
		{
			From:  "interface_adapters",
			Allow: []string{"interface_adapters", "use_cases", "entities"},
			Deny:  []string{"frameworks"},
		},
		{
			From:  "frameworks",
			Allow: []string{"frameworks", "interface_adapters", "use_cases", "entities"},
		},
	}
}

// mvcPresetLayers defines the three layers of the MVC / MVT pattern. Template
// engines (MVT) are treated as views.
func mvcPresetLayers() []LayerDefinition {
	return []LayerDefinition{
		{
			Name:     "model",
			Packages: []string{"model", "models", "entity", "entities", "schema", "schemas", "domain", "domains"},
		},
		{
			Name:     "view",
			Packages: []string{"view", "views", "template", "templates", "serializer", "serializers", "form", "forms"},
		},
		{
			Name:     "controller",
			Packages: []string{"controller", "controllers", "handler", "handlers", "router", "routers", "route", "routes", "viewset", "viewsets"},
		},
	}
}

// mvcPresetRules enforces MVC/MVT boundaries. A view depending directly on the
// model is permitted but discouraged (emits a warning); routing data through a
// controller is preferred.
func mvcPresetRules() []LayerRule {
	return []LayerRule{
		{
			From:  "model",
			Allow: []string{"model"},
			Deny:  []string{"view", "controller"},
		},
		{
			From:  "view",
			Allow: []string{"view", "controller"},
			Warn:  []string{"model"},
		},
		{
			From:  "controller",
			Allow: []string{"controller", "model", "view"},
		},
	}
}
