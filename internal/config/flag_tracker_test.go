package config

import (
	"sync"
	"testing"
)

func TestFlagTracker_Basic(t *testing.T) {
	ft := NewFlagTracker()

	// Test initial state
	if ft.WasSet("test") {
		t.Error("Expected flag 'test' to not be set initially")
	}

	// Test setting a flag
	ft.Set("test")
	if !ft.WasSet("test") {
		t.Error("Expected flag 'test' to be set after Set()")
	}

	// Test count
	if ft.Count() != 1 {
		t.Errorf("Expected count to be 1, got %d", ft.Count())
	}

	// Test clear
	ft.Clear()
	if ft.WasSet("test") {
		t.Error("Expected flag 'test' to not be set after Clear()")
	}
	if ft.Count() != 0 {
		t.Errorf("Expected count to be 0 after Clear(), got %d", ft.Count())
	}
}

func TestFlagTracker_WithInitialFlags(t *testing.T) {
	initial := map[string]bool{
		"flag1": true,
		"flag2": true,
		"flag3": false,
	}

	ft := NewFlagTrackerWithFlags(initial)

	if !ft.WasSet("flag1") {
		t.Error("Expected flag1 to be set")
	}
	if !ft.WasSet("flag2") {
		t.Error("Expected flag2 to be set")
	}
	if ft.WasSet("flag3") {
		t.Error("Expected flag3 to not be set")
	}
	if ft.WasSet("flag4") {
		t.Error("Expected flag4 to not be set")
	}
}

func TestFlagTracker_GetAll(t *testing.T) {
	ft := NewFlagTracker()
	ft.Set("flag1")
	ft.Set("flag2")

	all := ft.GetAll()
	if len(all) != 2 {
		t.Errorf("Expected 2 flags, got %d", len(all))
	}
	if !all["flag1"] || !all["flag2"] {
		t.Error("Expected both flag1 and flag2 to be true")
	}

	// Modify the returned map and ensure it doesn't affect the tracker
	all["flag3"] = true
	if ft.WasSet("flag3") {
		t.Error("Modifying returned map should not affect tracker")
	}
}

func TestFlagTracker_ConcurrentReadWrite(t *testing.T) {
	ft := NewFlagTracker()
	var wg sync.WaitGroup
	iterations := 1000
	goroutines := 10

	// Concurrent writes
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				flagName := "flag"
				if j%2 == 0 {
					flagName = "even"
				} else {
					flagName = "odd"
				}
				ft.Set(flagName)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = ft.WasSet("flag")
				_ = ft.WasSet("even")
				_ = ft.WasSet("odd")
				_ = ft.Count()
				_ = ft.GetAll()
			}
		}()
	}

	wg.Wait()
	// If we get here without panic or race condition, test passes
}

func TestFlagTracker_MergeMethods(t *testing.T) {
	ft := NewFlagTracker()
	ft.Set("explicit")

	// Test MergeString
	result := ft.MergeString("base", "override", "explicit")
	if result != "override" {
		t.Errorf("MergeString with explicit flag: expected 'override', got '%s'", result)
	}

	result = ft.MergeString("base", "override", "notset")
	if result != "base" {
		t.Errorf("MergeString without explicit flag: expected 'base', got '%s'", result)
	}

	// Test MergeInt
	intResult := ft.MergeInt(10, 20, "explicit")
	if intResult != 20 {
		t.Errorf("MergeInt with explicit flag: expected 20, got %d", intResult)
	}

	intResult = ft.MergeInt(10, 20, "notset")
	if intResult != 10 {
		t.Errorf("MergeInt without explicit flag: expected 10, got %d", intResult)
	}

	// Test MergeBool
	boolResult := ft.MergeBool(true, false, "explicit")
	if boolResult != false {
		t.Error("MergeBool with explicit flag: expected false, got true")
	}

	boolResult = ft.MergeBool(true, false, "notset")
	if boolResult != true {
		t.Error("MergeBool without explicit flag: expected true, got false")
	}

	// Test MergeFloat64
	floatResult := ft.MergeFloat64(1.5, 2.5, "explicit")
	if floatResult != 2.5 {
		t.Errorf("MergeFloat64 with explicit flag: expected 2.5, got %f", floatResult)
	}

	floatResult = ft.MergeFloat64(1.5, 2.5, "notset")
	if floatResult != 1.5 {
		t.Errorf("MergeFloat64 without explicit flag: expected 1.5, got %f", floatResult)
	}

	// Test MergeStringSlice
	sliceResult := ft.MergeStringSlice([]string{"a"}, []string{"b"}, "explicit")
	if len(sliceResult) != 1 || sliceResult[0] != "b" {
		t.Errorf("MergeStringSlice with explicit flag: expected ['b'], got %v", sliceResult)
	}

	sliceResult = ft.MergeStringSlice([]string{"a"}, []string{"b"}, "notset")
	if len(sliceResult) != 1 || sliceResult[0] != "a" {
		t.Errorf("MergeStringSlice without explicit flag: expected ['a'], got %v", sliceResult)
	}
}

func TestFlagTracker_NilInitialization(t *testing.T) {
	ft := NewFlagTrackerWithFlags(nil)
	
	// Should not panic and should work normally
	ft.Set("test")
	if !ft.WasSet("test") {
		t.Error("Expected flag 'test' to be set")
	}
}

func BenchmarkFlagTracker_WasSet(b *testing.B) {
	ft := NewFlagTracker()
	ft.Set("flag1")
	ft.Set("flag2")
	ft.Set("flag3")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = ft.WasSet("flag2")
		}
	})
}

func BenchmarkFlagTracker_Set(b *testing.B) {
	ft := NewFlagTracker()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			ft.Set("flag")
			i++
		}
	})
}