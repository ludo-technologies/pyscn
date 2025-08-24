package config

import (
	"sync"
	"testing"
)

func TestWasExplicitlySet(t *testing.T) {
	tests := []struct {
		name     string
		flags    map[string]bool
		flagName string
		want     bool
	}{
		{
			name:     "nil flags map",
			flags:    nil,
			flagName: "test",
			want:     false,
		},
		{
			name:     "empty flags map",
			flags:    map[string]bool{},
			flagName: "test",
			want:     false,
		},
		{
			name:     "flag not set",
			flags:    map[string]bool{"other": true},
			flagName: "test",
			want:     false,
		},
		{
			name:     "flag set to true",
			flags:    map[string]bool{"test": true},
			flagName: "test",
			want:     true,
		},
		{
			name:     "flag set to false",
			flags:    map[string]bool{"test": false},
			flagName: "test",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WasExplicitlySet(tt.flags, tt.flagName); got != tt.want {
				t.Errorf("WasExplicitlySet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeString(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		override string
		flagName string
		flags    map[string]bool
		want     string
	}{
		{
			name:     "flag not set, use base",
			base:     "base",
			override: "override",
			flagName: "test",
			flags:    map[string]bool{},
			want:     "base",
		},
		{
			name:     "flag set, use override",
			base:     "base",
			override: "override",
			flagName: "test",
			flags:    map[string]bool{"test": true},
			want:     "override",
		},
		{
			name:     "nil flags, use base",
			base:     "base",
			override: "override",
			flagName: "test",
			flags:    nil,
			want:     "base",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeString(tt.base, tt.override, tt.flagName, tt.flags); got != tt.want {
				t.Errorf("MergeString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeInt(t *testing.T) {
	tests := []struct {
		name     string
		base     int
		override int
		flagName string
		flags    map[string]bool
		want     int
	}{
		{
			name:     "flag not set, use base",
			base:     10,
			override: 20,
			flagName: "test",
			flags:    map[string]bool{},
			want:     10,
		},
		{
			name:     "flag set, use override",
			base:     10,
			override: 20,
			flagName: "test",
			flags:    map[string]bool{"test": true},
			want:     20,
		},
		{
			name:     "flag set with zero override",
			base:     10,
			override: 0,
			flagName: "test",
			flags:    map[string]bool{"test": true},
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeInt(tt.base, tt.override, tt.flagName, tt.flags); got != tt.want {
				t.Errorf("MergeInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeBool(t *testing.T) {
	tests := []struct {
		name     string
		base     bool
		override bool
		flagName string
		flags    map[string]bool
		want     bool
	}{
		{
			name:     "flag not set, use base true",
			base:     true,
			override: false,
			flagName: "test",
			flags:    map[string]bool{},
			want:     true,
		},
		{
			name:     "flag not set, use base false",
			base:     false,
			override: true,
			flagName: "test",
			flags:    map[string]bool{},
			want:     false,
		},
		{
			name:     "flag set, use override false",
			base:     true,
			override: false,
			flagName: "test",
			flags:    map[string]bool{"test": true},
			want:     false,
		},
		{
			name:     "flag set, use override true",
			base:     false,
			override: true,
			flagName: "test",
			flags:    map[string]bool{"test": true},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeBool(tt.base, tt.override, tt.flagName, tt.flags); got != tt.want {
				t.Errorf("MergeBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeFloat64(t *testing.T) {
	tests := []struct {
		name     string
		base     float64
		override float64
		flagName string
		flags    map[string]bool
		want     float64
	}{
		{
			name:     "flag not set, use base",
			base:     1.5,
			override: 2.5,
			flagName: "test",
			flags:    map[string]bool{},
			want:     1.5,
		},
		{
			name:     "flag set, use override",
			base:     1.5,
			override: 2.5,
			flagName: "test",
			flags:    map[string]bool{"test": true},
			want:     2.5,
		},
		{
			name:     "flag set with zero override",
			base:     1.5,
			override: 0.0,
			flagName: "test",
			flags:    map[string]bool{"test": true},
			want:     0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeFloat64(tt.base, tt.override, tt.flagName, tt.flags); got != tt.want {
				t.Errorf("MergeFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		base     []string
		override []string
		flagName string
		flags    map[string]bool
		want     []string
	}{
		{
			name:     "flag not set, use base",
			base:     []string{"a", "b"},
			override: []string{"c", "d"},
			flagName: "test",
			flags:    map[string]bool{},
			want:     []string{"a", "b"},
		},
		{
			name:     "flag set, use override",
			base:     []string{"a", "b"},
			override: []string{"c", "d"},
			flagName: "test",
			flags:    map[string]bool{"test": true},
			want:     []string{"c", "d"},
		},
		{
			name:     "flag set with empty override",
			base:     []string{"a", "b"},
			override: []string{},
			flagName: "test",
			flags:    map[string]bool{"test": true},
			want:     []string{"a", "b"},
		},
		{
			name:     "flag not set, both empty",
			base:     []string{},
			override: []string{},
			flagName: "test",
			flags:    map[string]bool{},
			want:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeStringSlice(tt.base, tt.override, tt.flagName, tt.flags)
			if len(got) != len(tt.want) {
				t.Errorf("MergeStringSlice() len = %v, want len %v", len(got), len(tt.want))
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("MergeStringSlice()[%d] = %v, want %v", i, v, tt.want[i])
				}
			}
		})
	}
}

// TestConcurrentAccess tests that the merge functions are safe for concurrent read access
func TestConcurrentAccess(t *testing.T) {
	flags := map[string]bool{
		"flag1": true,
		"flag2": false,
		"flag3": true,
	}

	var wg sync.WaitGroup
	iterations := 100
	goroutines := 10

	// Test concurrent reads
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = WasExplicitlySet(flags, "flag1")
				_ = MergeString("base", "override", "flag1", flags)
				_ = MergeInt(10, 20, "flag2", flags)
				_ = MergeBool(true, false, "flag3", flags)
				_ = MergeFloat64(1.0, 2.0, "flag1", flags)
				_ = MergeStringSlice([]string{"a"}, []string{"b"}, "flag2", flags)
			}
		}()
	}

	wg.Wait()
	// If we get here without panic, the test passes
}