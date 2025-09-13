package config

// WasExplicitlySet checks if a flag was explicitly set by the user
func WasExplicitlySet(flags map[string]bool, flagName string) bool {
	if flags == nil {
		return false
	}
	return flags[flagName]
}

// MergeString merges a string value, using override only if explicitly set
func MergeString(base, override, flagName string, flags map[string]bool) string {
	if WasExplicitlySet(flags, flagName) {
		return override
	}
	return base
}

// MergeInt merges an int value, using override only if explicitly set
func MergeInt(base, override int, flagName string, flags map[string]bool) int {
	if WasExplicitlySet(flags, flagName) {
		return override
	}
	return base
}

// MergeBool merges a bool value, using override only if explicitly set
func MergeBool(base, override bool, flagName string, flags map[string]bool) bool {
	if WasExplicitlySet(flags, flagName) {
		return override
	}
	return base
}

// MergeFloat64 merges a float64 value, using override only if explicitly set
func MergeFloat64(base, override float64, flagName string, flags map[string]bool) float64 {
	if WasExplicitlySet(flags, flagName) {
		return override
	}
	return base
}

// MergeStringSlice merges a string slice, using override only if explicitly set
func MergeStringSlice(base, override []string, flagName string, flags map[string]bool) []string {
	if WasExplicitlySet(flags, flagName) && len(override) > 0 {
		return override
	}
	return base
}
