package service

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

func TestNewParseCache(t *testing.T) {
	cache := NewParseCache()
	if cache == nil {
		t.Fatal("NewParseCache returned nil")
	}
	if cache.Len() != 0 {
		t.Fatalf("expected empty cache, got %d entries", cache.Len())
	}
}

func TestParseCachePutAndGet(t *testing.T) {
	cache := NewParseCache()

	result := &FileParseResult{
		Content: []byte("print('hello')"),
	}
	cache.Put("test.py", result)

	got, ok := cache.Get("test.py")
	if !ok {
		t.Fatal("expected cache hit for test.py")
	}
	if string(got.Content) != "print('hello')" {
		t.Fatalf("unexpected content: %s", got.Content)
	}
}

func TestParseCacheGetMiss(t *testing.T) {
	cache := NewParseCache()

	_, ok := cache.Get("nonexistent.py")
	if ok {
		t.Fatal("expected cache miss for nonexistent.py")
	}
}

func TestParseCacheSealPreventsWrite(t *testing.T) {
	cache := NewParseCache()
	cache.Put("a.py", &FileParseResult{Content: []byte("a")})
	cache.Seal()

	// Put after Seal should be a no-op
	cache.Put("b.py", &FileParseResult{Content: []byte("b")})

	_, ok := cache.Get("b.py")
	if ok {
		t.Fatal("expected Put after Seal to be ignored")
	}
	if cache.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", cache.Len())
	}
}

func TestParseCacheSealedConcurrentReads(t *testing.T) {
	cache := NewParseCache()
	for i := 0; i < 100; i++ {
		cache.Put(filepath.Join("dir", "file"+string(rune('0'+i%10))+".py"),
			&FileParseResult{Content: []byte("content")})
	}
	cache.Seal()

	var wg sync.WaitGroup
	for i := 0; i < runtime.GOMAXPROCS(0)*2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				cache.Get(filepath.Join("dir", "file"+string(rune('0'+j%10))+".py"))
			}
		}()
	}
	wg.Wait()
}

func TestParseCacheLen(t *testing.T) {
	cache := NewParseCache()
	cache.Put("a.py", &FileParseResult{})
	cache.Put("b.py", &FileParseResult{})
	cache.Put("c.py", &FileParseResult{})

	if cache.Len() != 3 {
		t.Fatalf("expected 3 entries, got %d", cache.Len())
	}
}

func TestPopulateParseCache(t *testing.T) {
	// Use a real testdata file
	testFile := filepath.Join("..", "testdata", "python", "simple", "functions.py")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("testdata file not found")
	}

	ctx := context.Background()
	cache := PopulateParseCache(ctx, []string{testFile}, ParseCachePopulatorConfig{
		BuildCFGs:   true,
		Concurrency: 2,
	})

	if cache.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", cache.Len())
	}

	result, ok := cache.Get(testFile)
	if !ok {
		t.Fatal("expected cache hit for test file")
	}
	if result.ParseErr != nil {
		t.Fatalf("unexpected parse error: %v", result.ParseErr)
	}
	if result.ParseResult == nil {
		t.Fatal("expected non-nil ParseResult")
	}
	if result.ParseResult.AST == nil {
		t.Fatal("expected non-nil AST")
	}
	if result.Content == nil {
		t.Fatal("expected non-nil Content")
	}
	if result.CFGs == nil {
		t.Fatal("expected non-nil CFGs when BuildCFGs=true")
	}
}

func TestPopulateParseCacheWithoutCFGs(t *testing.T) {
	testFile := filepath.Join("..", "testdata", "python", "simple", "functions.py")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("testdata file not found")
	}

	ctx := context.Background()
	cache := PopulateParseCache(ctx, []string{testFile}, ParseCachePopulatorConfig{
		BuildCFGs:   false,
		Concurrency: 1,
	})

	result, ok := cache.Get(testFile)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if result.CFGs != nil {
		t.Fatal("expected nil CFGs when BuildCFGs=false")
	}
}

func TestPopulateParseCacheNonexistentFile(t *testing.T) {
	ctx := context.Background()
	cache := PopulateParseCache(ctx, []string{"/nonexistent/file.py"}, ParseCachePopulatorConfig{
		BuildCFGs: false,
	})

	if cache.Len() != 1 {
		t.Fatalf("expected 1 entry (with error), got %d", cache.Len())
	}

	result, ok := cache.Get("/nonexistent/file.py")
	if !ok {
		t.Fatal("expected cache entry for nonexistent file")
	}
	if result.ParseErr == nil {
		t.Fatal("expected parse error for nonexistent file")
	}
}

func TestPopulateParseCacheMultipleFiles(t *testing.T) {
	files := []string{
		filepath.Join("..", "testdata", "python", "simple", "functions.py"),
		filepath.Join("..", "testdata", "python", "simple", "classes.py"),
		filepath.Join("..", "testdata", "python", "simple", "control_flow.py"),
	}

	for _, f := range files {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			t.Skipf("testdata file not found: %s", f)
		}
	}

	ctx := context.Background()
	cache := PopulateParseCache(ctx, files, ParseCachePopulatorConfig{
		BuildCFGs:   true,
		Concurrency: 2,
	})

	if cache.Len() != 3 {
		t.Fatalf("expected 3 entries, got %d", cache.Len())
	}

	for _, f := range files {
		result, ok := cache.Get(f)
		if !ok {
			t.Fatalf("expected cache hit for %s", f)
		}
		if result.ParseErr != nil {
			t.Fatalf("unexpected parse error for %s: %v", f, result.ParseErr)
		}
		if result.ParseResult == nil {
			t.Fatalf("expected non-nil ParseResult for %s", f)
		}
	}
}
