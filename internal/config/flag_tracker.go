package config

import (
	"sync"
)

// FlagTracker provides thread-safe tracking of explicitly set flags
type FlagTracker struct {
	mu    sync.RWMutex
	flags map[string]bool
}

// NewFlagTracker creates a new thread-safe flag tracker
func NewFlagTracker() *FlagTracker {
	return &FlagTracker{
		flags: make(map[string]bool),
	}
}

// NewFlagTrackerWithFlags creates a new flag tracker with initial flags
func NewFlagTrackerWithFlags(flags map[string]bool) *FlagTracker {
	if flags == nil {
		flags = make(map[string]bool)
	}
	// Create a copy to avoid external modifications
	copiedFlags := make(map[string]bool, len(flags))
	for k, v := range flags {
		copiedFlags[k] = v
	}
	return &FlagTracker{
		flags: copiedFlags,
	}
}

// Set marks a flag as explicitly set
func (ft *FlagTracker) Set(flagName string) {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	ft.flags[flagName] = true
}

// WasSet checks if a flag was explicitly set
func (ft *FlagTracker) WasSet(flagName string) bool {
	ft.mu.RLock()
	defer ft.mu.RUnlock()
	return ft.flags[flagName]
}

// GetAll returns a copy of all flags (safe for concurrent access)
func (ft *FlagTracker) GetAll() map[string]bool {
	ft.mu.RLock()
	defer ft.mu.RUnlock()
	
	// Return a copy to prevent external modifications
	result := make(map[string]bool, len(ft.flags))
	for k, v := range ft.flags {
		result[k] = v
	}
	return result
}

// Clear removes all flag tracking
func (ft *FlagTracker) Clear() {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	ft.flags = make(map[string]bool)
}

// Count returns the number of explicitly set flags
func (ft *FlagTracker) Count() int {
	ft.mu.RLock()
	defer ft.mu.RUnlock()
	return len(ft.flags)
}

// MergeString merges a string value using thread-safe flag checking
func (ft *FlagTracker) MergeString(base, override, flagName string) string {
	if ft.WasSet(flagName) {
		return override
	}
	return base
}

// MergeInt merges an int value using thread-safe flag checking
func (ft *FlagTracker) MergeInt(base, override int, flagName string) int {
	if ft.WasSet(flagName) {
		return override
	}
	return base
}

// MergeBool merges a bool value using thread-safe flag checking
func (ft *FlagTracker) MergeBool(base, override bool, flagName string) bool {
	if ft.WasSet(flagName) {
		return override
	}
	return base
}

// MergeFloat64 merges a float64 value using thread-safe flag checking
func (ft *FlagTracker) MergeFloat64(base, override float64, flagName string) float64 {
	if ft.WasSet(flagName) {
		return override
	}
	return base
}

// MergeStringSlice merges a string slice using thread-safe flag checking
func (ft *FlagTracker) MergeStringSlice(base, override []string, flagName string) []string {
	if ft.WasSet(flagName) && len(override) > 0 {
		return override
	}
	return base
}