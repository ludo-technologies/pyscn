package config

import (
	"reflect"
	"slices"
	"testing"
)

// TestArchitectureStylePreset_LayeredMatchesEmpty verifies that an empty style
// resolves to the same preset as the explicit "layered" style.
func TestArchitectureStylePreset_LayeredMatchesEmpty(t *testing.T) {
	emptyLayers, emptyRules := ArchitectureStylePreset("")
	layeredLayers, layeredRules := ArchitectureStylePreset(ArchitectureStyleLayered)

	if !reflect.DeepEqual(emptyLayers, layeredLayers) {
		t.Errorf("empty-style layers differ from layered layers")
	}
	if !reflect.DeepEqual(emptyRules, layeredRules) {
		t.Errorf("empty-style rules differ from layered rules")
	}
}

// TestArchitectureStylePreset_LayeredBackwardCompat verifies that the layered
// preset matches the embedded default_config.toml architecture section exactly,
// so projects that relied on the previous default keep identical behavior.
func TestArchitectureStylePreset_LayeredBackwardCompat(t *testing.T) {
	defaultCfg, err := LoadDefaultConfigFromTOML()
	if err != nil {
		t.Fatalf("LoadDefaultConfigFromTOML failed: %v", err)
	}

	layers, rules := ArchitectureStylePreset(ArchitectureStyleLayered)

	// Compare layer names and packages.
	if len(layers) != len(defaultCfg.Architecture.Layers) {
		t.Fatalf("layer count mismatch: preset=%d default=%d", len(layers), len(defaultCfg.Architecture.Layers))
	}
	for i, l := range layers {
		d := defaultCfg.Architecture.Layers[i]
		if l.Name != d.Name {
			t.Errorf("layer[%d] name mismatch: preset=%q default=%q", i, l.Name, d.Name)
		}
		if !reflect.DeepEqual(l.Packages, d.Packages) {
			t.Errorf("layer[%d] (%s) packages mismatch:\n preset=%v\ndefault=%v", i, l.Name, l.Packages, d.Packages)
		}
	}

	// Compare rules (From/Allow/Deny).
	if len(rules) != len(defaultCfg.Architecture.Rules) {
		t.Fatalf("rule count mismatch: preset=%d default=%d", len(rules), len(defaultCfg.Architecture.Rules))
	}
	for i, r := range rules {
		d := defaultCfg.Architecture.Rules[i]
		if r.From != d.From {
			t.Errorf("rule[%d] from mismatch: preset=%q default=%q", i, r.From, d.From)
		}
		if !reflect.DeepEqual(r.Allow, d.Allow) {
			t.Errorf("rule[%d] (%s) allow mismatch:\n preset=%v\ndefault=%v", i, r.From, r.Allow, d.Allow)
		}
		if !reflect.DeepEqual(r.Deny, d.Deny) {
			t.Errorf("rule[%d] (%s) deny mismatch:\n preset=%v\ndefault=%v", i, r.From, r.Deny, d.Deny)
		}
	}
}

// rulesByFrom indexes a preset's rules by their From layer.
func rulesByFrom(rules []LayerRule) map[string]LayerRule {
	m := make(map[string]LayerRule, len(rules))
	for _, r := range rules {
		m[r.From] = r
	}
	return m
}

// TestArchitectureStylePreset_HexagonalDomainIsolated verifies that the
// hexagonal preset enforces the Dependency Inversion Principle: the domain may
// depend only on itself and explicitly denies ports and adapters.
func TestArchitectureStylePreset_HexagonalDomainIsolated(t *testing.T) {
	_, rules := ArchitectureStylePreset(ArchitectureStyleHexagonal)
	byFrom := rulesByFrom(rules)

	domain, ok := byFrom["domain"]
	if !ok {
		t.Fatal("hexagonal preset must define a 'domain' rule")
	}
	if !reflect.DeepEqual(domain.Allow, []string{"domain"}) {
		t.Errorf("domain should allow only itself, got %v", domain.Allow)
	}
	for _, forbidden := range []string{"ports", "adapters"} {
		if !slices.Contains(domain.Deny, forbidden) {
			t.Errorf("domain should deny %q, deny=%v", forbidden, domain.Deny)
		}
	}
}

// TestArchitectureStylePreset_CleanEntitiesIsolated verifies that the clean
// preset keeps the innermost layer (entities) free of outward dependencies.
func TestArchitectureStylePreset_CleanEntitiesIsolated(t *testing.T) {
	_, rules := ArchitectureStylePreset(ArchitectureStyleClean)
	byFrom := rulesByFrom(rules)

	entities, ok := byFrom["entities"]
	if !ok {
		t.Fatal("clean preset must define an 'entities' rule")
	}
	if !reflect.DeepEqual(entities.Allow, []string{"entities"}) {
		t.Errorf("entities should allow only itself, got %v", entities.Allow)
	}
	for _, forbidden := range []string{"use_cases", "interface_adapters", "frameworks"} {
		if !slices.Contains(entities.Deny, forbidden) {
			t.Errorf("entities should deny %q, deny=%v", forbidden, entities.Deny)
		}
	}

	// frameworks (outermost) may depend on every inner layer.
	frameworks := byFrom["frameworks"]
	for _, inner := range []string{"interface_adapters", "use_cases", "entities"} {
		if !slices.Contains(frameworks.Allow, inner) {
			t.Errorf("frameworks should allow %q, allow=%v", inner, frameworks.Allow)
		}
	}
}

// TestArchitectureStylePreset_AllStylesNonEmpty verifies every named style
// returns a non-empty preset.
func TestArchitectureStylePreset_AllStylesNonEmpty(t *testing.T) {
	for _, style := range []string{ArchitectureStyleLayered, ArchitectureStyleHexagonal, ArchitectureStyleClean} {
		layers, rules := ArchitectureStylePreset(style)
		if len(layers) == 0 {
			t.Errorf("style %q returned no layers", style)
		}
		if len(rules) == 0 {
			t.Errorf("style %q returned no rules", style)
		}
	}
}

// TestArchitectureStylePreset_UnknownReturnsNil verifies that an unrecognized
// style returns no preset so callers can fall back to auto-detection.
func TestArchitectureStylePreset_UnknownReturnsNil(t *testing.T) {
	layers, rules := ArchitectureStylePreset("nonexistent-style")
	if layers != nil {
		t.Errorf("expected nil layers for unknown style, got %v", layers)
	}
	if rules != nil {
		t.Errorf("expected nil rules for unknown style, got %v", rules)
	}
}
