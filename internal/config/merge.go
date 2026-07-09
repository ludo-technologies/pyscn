package config

// Merge returns override unless it is the zero value for its type, in which
// case base is returned. The zero value (0, "", 0.0, ...) means "not set";
// actual defaults live in the config layer, never in override values.
func Merge[T comparable](base, override T) T {
	var zero T
	if override == zero {
		return base
	}
	return override
}

// MergePtr returns override unless it is nil, in which case base is returned.
// nil means "not set"; a non-nil pointer is an explicit value, including the
// zero value (e.g. explicit false for *bool).
func MergePtr[T any](base, override *T) *T {
	if override == nil {
		return base
	}
	return override
}

// MergeSlice returns override unless it is empty, in which case base is
// returned. An empty slice means "not set".
func MergeSlice[T any](base, override []T) []T {
	if len(override) == 0 {
		return base
	}
	return override
}
