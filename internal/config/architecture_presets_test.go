package config

import (
	"reflect"
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
