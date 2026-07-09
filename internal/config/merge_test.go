package config

import (
	"testing"
)

func TestMerge(t *testing.T) {
	t.Run("int: zero override keeps base", func(t *testing.T) {
		if got := Merge(10, 0); got != 10 {
			t.Errorf("expected 10, got %d", got)
		}
	})
	t.Run("int: non-zero override wins", func(t *testing.T) {
		if got := Merge(10, 25); got != 25 {
			t.Errorf("expected 25, got %d", got)
		}
	})
	t.Run("int: override equal to a typical default still wins", func(t *testing.T) {
		// Regression for issue #553: explicit value matching a default
		// must not be discarded.
		if got := Merge(15, 9); got != 9 {
			t.Errorf("expected 9, got %d", got)
		}
	})
	t.Run("string: empty override keeps base", func(t *testing.T) {
		if got := Merge("base", ""); got != "base" {
			t.Errorf("expected base, got %q", got)
		}
	})
	t.Run("string: non-empty override wins", func(t *testing.T) {
		if got := Merge("base", "override"); got != "override" {
			t.Errorf("expected override, got %q", got)
		}
	})
	t.Run("float64: zero override keeps base", func(t *testing.T) {
		if got := Merge(0.65, 0.0); got != 0.65 {
			t.Errorf("expected 0.65, got %f", got)
		}
	})
	t.Run("float64: non-zero override wins", func(t *testing.T) {
		if got := Merge(0.65, 0.8); got != 0.8 {
			t.Errorf("expected 0.8, got %f", got)
		}
	})
}

func TestMergePtr(t *testing.T) {
	boolPtr := func(v bool) *bool { return &v }

	t.Run("nil override keeps base", func(t *testing.T) {
		base := boolPtr(true)
		if got := MergePtr(base, nil); got != base {
			t.Error("expected base pointer to be kept")
		}
	})
	t.Run("non-nil override wins even when false", func(t *testing.T) {
		base := boolPtr(true)
		override := boolPtr(false)
		if got := MergePtr(base, override); got != override {
			t.Error("expected override pointer to win")
		}
	})
	t.Run("both nil returns nil", func(t *testing.T) {
		if got := MergePtr[bool](nil, nil); got != nil {
			t.Error("expected nil")
		}
	})
}

func TestMergeSlice(t *testing.T) {
	t.Run("empty override keeps base", func(t *testing.T) {
		base := []string{"a", "b"}
		got := MergeSlice(base, nil)
		if len(got) != 2 || got[0] != "a" {
			t.Errorf("expected base slice, got %v", got)
		}
	})
	t.Run("non-empty override wins", func(t *testing.T) {
		base := []string{"a"}
		override := []string{"x", "y"}
		got := MergeSlice(base, override)
		if len(got) != 2 || got[0] != "x" {
			t.Errorf("expected override slice, got %v", got)
		}
	})
}
