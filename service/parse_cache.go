package service

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/ludo-technologies/pyscn/internal/analyzer"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// FileParseResult holds the cached parse result for a single file.
type FileParseResult struct {
	Content     []byte
	ParseResult *parser.ParseResult
	CFGs        map[string]*analyzer.CFG // nil if not computed
	ParseErr    error
	CFGErr      error
}

// ParseCache stores pre-parsed results for sharing across analysis services.
// After Seal() is called the cache is read-only and safe for concurrent access
// without locks.
type ParseCache struct {
	results map[string]*FileParseResult
	sealed  bool
}

// NewParseCache creates a new empty ParseCache.
func NewParseCache() *ParseCache {
	return &ParseCache{
		results: make(map[string]*FileParseResult),
	}
}

// Put stores a parse result. Must be called before Seal().
func (c *ParseCache) Put(filePath string, result *FileParseResult) {
	if c.sealed {
		return
	}
	c.results[filePath] = result
}

// Seal marks the cache as read-only. After this call no more Put() is allowed
// and Get() can be safely called from multiple goroutines without locks.
func (c *ParseCache) Seal() {
	c.sealed = true
}

// Get retrieves a cached parse result. Returns (result, true) on hit.
func (c *ParseCache) Get(filePath string) (*FileParseResult, bool) {
	r, ok := c.results[filePath]
	return r, ok
}

// Len returns the number of entries in the cache.
func (c *ParseCache) Len() int {
	return len(c.results)
}

// ParseCacheAware is implemented by services that can accept a pre-populated
// parse cache to avoid redundant file parsing.
type ParseCacheAware interface {
	SetParseCache(cache *ParseCache)
}

// ParseCachePopulatorConfig controls how PopulateParseCache works.
type ParseCachePopulatorConfig struct {
	BuildCFGs   bool // whether to also build CFGs for each file
	Concurrency int  // 0 means runtime.GOMAXPROCS(0)
}

// PopulateParseCache parses all files in parallel and returns a sealed cache.
// Each goroutine creates its own parser.Parser because tree-sitter is not
// thread-safe.
func PopulateParseCache(ctx context.Context, files []string, cfg ParseCachePopulatorConfig) *ParseCache {
	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		concurrency = runtime.GOMAXPROCS(0)
	}

	cache := NewParseCache()

	type indexedResult struct {
		path   string
		result *FileParseResult
	}

	results := make([]indexedResult, len(files))

	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)

	for i, filePath := range files {
		wg.Add(1)
		go func(idx int, fp string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			r := &FileParseResult{}

			// Read file
			content, err := os.ReadFile(fp)
			if err != nil {
				r.ParseErr = fmt.Errorf("failed to read file %s: %w", fp, err)
				results[idx] = indexedResult{path: fp, result: r}
				return
			}
			r.Content = content

			// Parse — each goroutine gets its own parser (tree-sitter is not thread-safe)
			p := parser.New()
			parseResult, err := p.Parse(ctx, content)
			if err != nil {
				r.ParseErr = fmt.Errorf("parse error: %w", err)
				results[idx] = indexedResult{path: fp, result: r}
				return
			}
			r.ParseResult = parseResult

			// Optionally build CFGs
			if cfg.BuildCFGs && parseResult.AST != nil {
				builder := analyzer.NewCFGBuilder()
				cfgs, err := builder.BuildAll(parseResult.AST)
				if err != nil {
					r.CFGErr = fmt.Errorf("CFG construction failed: %w", err)
				}
				r.CFGs = cfgs
			}

			results[idx] = indexedResult{path: fp, result: r}
		}(i, filePath)
	}

	wg.Wait()

	// Populate cache from collected results (single-threaded, no lock needed)
	for _, ir := range results {
		if ir.result != nil {
			cache.Put(ir.path, ir.result)
		}
	}
	cache.Seal()

	return cache
}
