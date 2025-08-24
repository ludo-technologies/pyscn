package main

// ConfigMergeHelper provides configuration merging functionality that respects
// which CLI flags were explicitly set by the user
type ConfigMergeHelper struct {
	ExplicitFlags map[string]bool
}

// NewConfigMergeHelper creates a new configuration merge helper
func NewConfigMergeHelper(explicitFlags map[string]bool) *ConfigMergeHelper {
	if explicitFlags == nil {
		explicitFlags = make(map[string]bool)
	}
	return &ConfigMergeHelper{
		ExplicitFlags: explicitFlags,
	}
}

// WasExplicitlySet checks if a flag was explicitly set by the user
func (h *ConfigMergeHelper) WasExplicitlySet(flagName string) bool {
	if h.ExplicitFlags == nil {
		return false
	}
	return h.ExplicitFlags[flagName]
}

// MergeString merges a string value, using override only if explicitly set
func (h *ConfigMergeHelper) MergeString(base, override, flagName string) string {
	if h.WasExplicitlySet(flagName) || (override != "" && !h.hasExplicitFlags()) {
		return override
	}
	return base
}

// MergeInt merges an int value, using override only if explicitly set
func (h *ConfigMergeHelper) MergeInt(base, override int, flagName string) int {
	if h.WasExplicitlySet(flagName) {
		return override
	}
	return base
}

// MergeBool merges a bool value, using override only if explicitly set
func (h *ConfigMergeHelper) MergeBool(base, override bool, flagName string) bool {
	if h.WasExplicitlySet(flagName) {
		return override
	}
	return base
}

// MergeFloat64 merges a float64 value, using override only if explicitly set
func (h *ConfigMergeHelper) MergeFloat64(base, override float64, flagName string) float64 {
	if h.WasExplicitlySet(flagName) {
		return override
	}
	return base
}

// MergeStringSlice merges a string slice, using override only if explicitly set
func (h *ConfigMergeHelper) MergeStringSlice(base, override []string, flagName string) []string {
	if h.WasExplicitlySet(flagName) && len(override) > 0 {
		return override
	}
	if len(base) == 0 && len(override) > 0 && !h.hasExplicitFlags() {
		return override
	}
	return base
}

// hasExplicitFlags returns true if any flags were explicitly set
func (h *ConfigMergeHelper) hasExplicitFlags() bool {
	return len(h.ExplicitFlags) > 0
}