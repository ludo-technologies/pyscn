# Testing Guide

Comprehensive testing is crucial for pyqol's reliability and performance. This guide covers testing strategies, patterns, and best practices.

## Testing Philosophy

1. **Test Behavior, Not Implementation**: Focus on what the code does, not how
2. **Fast Feedback**: Tests should run quickly to encourage frequent execution
3. **Isolation**: Tests should not depend on each other
4. **Clarity**: Test names should describe what is being tested
5. **Coverage**: Aim for >80% code coverage, 100% for critical paths

## Test Organization

```
pyqol/
├── internal/
│   ├── parser/
│   │   ├── parser.go
│   │   ├── parser_test.go      # Unit tests
│   │   └── parser_bench_test.go # Benchmarks
│   └── analyzer/
│       ├── cfg.go
│       ├── cfg_test.go
│       └── cfg_integration_test.go
├── testdata/                    # Test fixtures
│   ├── python/
│   │   ├── simple/
│   │   ├── complex/
│   │   └── invalid/
│   └── golden/                  # Expected outputs
└── test/                        # Integration tests
    └── e2e/
        └── cli_test.go
```

## Test Types

### 1. Unit Tests

Test individual functions and methods in isolation.

```go
// parser_test.go
package parser

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestParseSimpleFunction(t *testing.T) {
    source := `
def hello():
    print("Hello, World!")
`
    parser := NewParser()
    ast, err := parser.Parse([]byte(source))
    
    require.NoError(t, err)
    assert.NotNil(t, ast)
    assert.Equal(t, "module", ast.Type)
    assert.Len(t, ast.Children, 1)
    
    funcDef := ast.Children[0]
    assert.Equal(t, "function_definition", funcDef.Type)
    assert.Equal(t, "hello", funcDef.Name)
}
```

### 2. Table-Driven Tests

Use for testing multiple scenarios with similar setup.

```go
func TestIdentifyDeadCode(t *testing.T) {
    tests := []struct {
        name     string
        source   string
        expected []Finding
        wantErr  bool
    }{
        {
            name: "unreachable after return",
            source: `
def foo():
    return 1
    print("unreachable")  # Dead code
`,
            expected: []Finding{
                {
                    Type:     DeadCode,
                    Line:     4,
                    Message:  "unreachable code after return",
                    Severity: Warning,
                },
            },
        },
        {
            name: "unreachable in conditional",
            source: `
def bar():
    if True:
        return 1
    else:
        return 2
    print("unreachable")  # Dead code
`,
            expected: []Finding{
                {
                    Type:     DeadCode,
                    Line:     7,
                    Message:  "unreachable code after return",
                    Severity: Warning,
                },
            },
        },
        {
            name: "no dead code",
            source: `
def baz():
    if condition:
        return 1
    print("reachable")
    return 2
`,
            expected: []Finding{},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            detector := NewDeadCodeDetector()
            findings, err := detector.Analyze(tt.source)
            
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            require.NoError(t, err)
            assert.Equal(t, tt.expected, findings)
        })
    }
}
```

### 3. Integration Tests

Test component interactions.

```go
// cfg_integration_test.go
//go:build integration

package analyzer_test

import (
    "testing"
    "path/filepath"
    
    "github.com/pyqol/pyqol/internal/parser"
    "github.com/pyqol/pyqol/internal/analyzer"
)

func TestCFGWithRealPythonFiles(t *testing.T) {
    files, err := filepath.Glob("../../testdata/python/complex/*.py")
    require.NoError(t, err)
    
    p := parser.NewParser()
    builder := analyzer.NewCFGBuilder()
    
    for _, file := range files {
        t.Run(filepath.Base(file), func(t *testing.T) {
            source, err := os.ReadFile(file)
            require.NoError(t, err)
            
            ast, err := p.Parse(source)
            require.NoError(t, err)
            
            cfg, err := builder.Build(ast)
            require.NoError(t, err)
            
            // Verify CFG properties
            assert.NotNil(t, cfg.Entry)
            assert.NotNil(t, cfg.Exit)
            assert.True(t, isConnected(cfg))
        })
    }
}
```

### 4. End-to-End Tests

Test the complete system from CLI to output.

```go
// test/e2e/cli_test.go
//go:build e2e

package e2e

import (
    "bytes"
    "os/exec"
    "testing"
)

func TestCLIAnalyzeCommand(t *testing.T) {
    tests := []struct {
        name     string
        args     []string
        fixture  string
        contains []string
        exitCode int
    }{
        {
            name:     "analyze simple file",
            args:     []string{"analyze", "testdata/simple.py"},
            fixture:  "simple.py",
            contains: []string{"Analysis complete", "0 issues found"},
            exitCode: 0,
        },
        {
            name:     "detect dead code",
            args:     []string{"analyze", "testdata/dead_code.py"},
            fixture:  "dead_code.py",
            contains: []string{"Dead code detected", "Line 10"},
            exitCode: 1,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd := exec.Command("./pyqol", tt.args...)
            var out bytes.Buffer
            cmd.Stdout = &out
            cmd.Stderr = &out
            
            err := cmd.Run()
            output := out.String()
            
            if e, ok := err.(*exec.ExitError); ok {
                assert.Equal(t, tt.exitCode, e.ExitCode())
            } else if tt.exitCode != 0 {
                t.Errorf("expected exit code %d, got 0", tt.exitCode)
            }
            
            for _, expected := range tt.contains {
                assert.Contains(t, output, expected)
            }
        })
    }
}
```

### 5. Benchmark Tests

Measure and track performance.

```go
// apted_bench_test.go
package analyzer

import (
    "testing"
)

func BenchmarkAPTEDSmallTree(b *testing.B) {
    tree1 := buildTree(10)  // 10 nodes
    tree2 := buildTree(10)
    
    analyzer := NewAPTEDAnalyzer()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = analyzer.Distance(tree1, tree2)
    }
}

func BenchmarkAPTEDMediumTree(b *testing.B) {
    tree1 := buildTree(100)  // 100 nodes
    tree2 := buildTree(100)
    
    analyzer := NewAPTEDAnalyzer()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = analyzer.Distance(tree1, tree2)
    }
}

func BenchmarkAPTEDLargeTree(b *testing.B) {
    tree1 := buildTree(1000)  // 1000 nodes
    tree2 := buildTree(1000)
    
    analyzer := NewAPTEDAnalyzer()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = analyzer.Distance(tree1, tree2)
    }
}

// Run with: go test -bench=APTED -benchmem
```

### 6. Fuzz Tests

Find edge cases automatically.

```go
// parser_fuzz_test.go
//go:build go1.18

package parser

import (
    "testing"
)

func FuzzParser(f *testing.F) {
    // Add seed corpus
    f.Add([]byte("def foo(): pass"))
    f.Add([]byte("class Bar: pass"))
    f.Add([]byte("if True: print(1)"))
    
    parser := NewParser()
    
    f.Fuzz(func(t *testing.T, data []byte) {
        // Parser should not panic on any input
        ast, err := parser.Parse(data)
        
        if err == nil {
            // If parsing succeeded, AST should be valid
            assert.NotNil(t, ast)
            assert.NotEmpty(t, ast.Type)
        }
        // Errors are expected for invalid input
    })
}

// Run with: go test -fuzz=FuzzParser -fuzztime=10s
```

## Test Fixtures

### Directory Structure

```
testdata/
├── python/
│   ├── simple/
│   │   ├── hello.py
│   │   ├── variables.py
│   │   └── functions.py
│   ├── complex/
│   │   ├── classes.py
│   │   ├── decorators.py
│   │   └── async.py
│   ├── edge_cases/
│   │   ├── empty.py
│   │   ├── syntax_error.py
│   │   └── unicode.py
│   └── benchmarks/
│       ├── small.py    # ~100 lines
│       ├── medium.py   # ~1000 lines
│       └── large.py    # ~10000 lines
└── golden/
    ├── hello.json      # Expected output
    └── hello.sarif     # Expected SARIF output
```

### Example Fixture

```python
# testdata/python/simple/dead_code.py
"""Test fixture for dead code detection."""

def unreachable_after_return():
    return True
    print("This is dead code")  # line 5

def unreachable_after_raise():
    raise ValueError("Error")
    x = 1  # line 9

def unreachable_branch():
    if True:
        return 1
    else:
        # This branch is dead
        return 2  # line 15

def unused_variable():
    x = 1  # Unused variable
    y = 2
    return y
```

## Testing Patterns

### 1. Test Helpers

Create reusable test utilities.

```go
// test_helpers.go
package testutil

import (
    "os"
    "path/filepath"
    "testing"
)

func LoadFixture(t *testing.T, name string) []byte {
    t.Helper()
    
    path := filepath.Join("testdata", "python", name)
    data, err := os.ReadFile(path)
    if err != nil {
        t.Fatalf("failed to load fixture %s: %v", name, err)
    }
    
    return data
}

func LoadGolden(t *testing.T, name string) []byte {
    t.Helper()
    
    path := filepath.Join("testdata", "golden", name)
    data, err := os.ReadFile(path)
    if err != nil {
        t.Fatalf("failed to load golden file %s: %v", name, err)
    }
    
    return data
}

func AssertJSONEqual(t *testing.T, expected, actual []byte) {
    t.Helper()
    
    var exp, act interface{}
    
    if err := json.Unmarshal(expected, &exp); err != nil {
        t.Fatalf("failed to unmarshal expected JSON: %v", err)
    }
    
    if err := json.Unmarshal(actual, &act); err != nil {
        t.Fatalf("failed to unmarshal actual JSON: %v", err)
    }
    
    assert.Equal(t, exp, act)
}
```

### 2. Mock Objects

Use interfaces for easy mocking.

```go
// mocks.go
package mocks

type MockParser struct {
    ParseFunc func([]byte) (*AST, error)
}

func (m *MockParser) Parse(source []byte) (*AST, error) {
    if m.ParseFunc != nil {
        return m.ParseFunc(source)
    }
    return nil, nil
}

// Usage in tests
func TestAnalyzerWithMockParser(t *testing.T) {
    mockParser := &MockParser{
        ParseFunc: func(source []byte) (*AST, error) {
            return &AST{Type: "module"}, nil
        },
    }
    
    analyzer := NewAnalyzer(mockParser)
    // test analyzer behavior
}
```

### 3. Golden Files

Use golden files for complex output validation.

```go
func TestJSONOutput(t *testing.T) {
    source := LoadFixture(t, "simple/hello.py")
    
    analyzer := NewAnalyzer()
    findings, err := analyzer.Analyze(source)
    require.NoError(t, err)
    
    output, err := json.MarshalIndent(findings, "", "  ")
    require.NoError(t, err)
    
    golden := LoadGolden(t, "hello.json")
    
    if *update {
        // Update golden file with -update flag
        err := os.WriteFile("testdata/golden/hello.json", output, 0644)
        require.NoError(t, err)
    } else {
        AssertJSONEqual(t, golden, output)
    }
}

// Run with: go test -update to update golden files
```

### 4. Parallel Tests

Run independent tests in parallel.

```go
func TestParallelAnalysis(t *testing.T) {
    files := []string{"file1.py", "file2.py", "file3.py"}
    
    for _, file := range files {
        file := file // capture loop variable
        t.Run(file, func(t *testing.T) {
            t.Parallel() // Run this test in parallel
            
            source := LoadFixture(t, file)
            analyzer := NewAnalyzer()
            
            _, err := analyzer.Analyze(source)
            assert.NoError(t, err)
        })
    }
}
```

## Coverage

### Running Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# View coverage by package
go test -coverprofile=coverage.out -coverpkg=./... ./...
```

### Coverage Requirements

- Overall: >80%
- Critical paths: 100%
- Parser: >90%
- Analyzer: >85%
- CLI: >70%

### Excluding from Coverage

```go
// Code that should not be counted in coverage

//go:build !test

func debugOnly() {
    // Debug code not used in tests
}
```

## Performance Testing

### Benchmarking Best Practices

```go
func BenchmarkCFGConstruction(b *testing.B) {
    // Setup - not timed
    source := LoadLargeFixture()
    parser := NewParser()
    ast, _ := parser.Parse(source)
    
    b.ResetTimer() // Start timing here
    
    for i := 0; i < b.N; i++ {
        builder := NewCFGBuilder()
        _ = builder.Build(ast)
    }
    
    b.ReportMetric(float64(len(source))/float64(b.Elapsed().Seconds()), "bytes/s")
}
```

### Memory Profiling

```bash
# Run benchmarks with memory profiling
go test -bench=. -benchmem -memprofile=mem.prof

# Analyze memory profile
go tool pprof mem.prof
```

### CPU Profiling

```bash
# Run benchmarks with CPU profiling
go test -bench=. -cpuprofile=cpu.prof

# Analyze CPU profile
go tool pprof cpu.prof

# Generate flame graph
go tool pprof -http=:8080 cpu.prof
```

## Continuous Integration

### CI Test Stages

```yaml
# .github/workflows/ci.yml
test:
  runs-on: ubuntu-latest
  steps:
    - name: Unit Tests
      run: go test -v -race ./...
    
    - name: Integration Tests
      run: go test -v -tags=integration ./...
    
    - name: Coverage
      run: |
        go test -coverprofile=coverage.out ./...
        go tool cover -func=coverage.out
    
    - name: Benchmarks
      run: go test -bench=. -benchtime=10x ./...
```

## Test Commands

### Quick Reference

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package
go test ./internal/parser

# Run specific test
go test -run TestParseSimpleFunction ./internal/parser

# Run with race detection
go test -race ./...

# Run with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./...

# Run fuzz tests
go test -fuzz=FuzzParser -fuzztime=30s ./internal/parser

# Run integration tests
go test -tags=integration ./...

# Run e2e tests
go test -tags=e2e ./test/e2e

# Update golden files
go test -update ./...

# Run tests with timeout
go test -timeout 30s ./...

# Run tests in parallel
go test -parallel 4 ./...
```

## Debugging Tests

### Verbose Output

```go
func TestWithLogging(t *testing.T) {
    t.Logf("Starting test with input: %v", input)
    
    result, err := SomeFunction(input)
    
    t.Logf("Result: %v, Error: %v", result, err)
    
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
        t.Logf("Debug info: %+v", debugInfo)
    }
}
```

### Using Delve Debugger

```bash
# Debug a specific test
dlv test ./internal/parser -- -test.run TestParseSimpleFunction

# Set breakpoint and run
(dlv) break parser.go:42
(dlv) continue
```

## Test Maintenance

### Regular Tasks

1. **Weekly**: Review and update test coverage
2. **Monthly**: Run and analyze benchmarks
3. **Quarterly**: Update test fixtures and golden files
4. **Before Release**: Full test suite with all tags

### Test Review Checklist

- [ ] Test name clearly describes what is being tested
- [ ] Test covers both success and failure cases
- [ ] Test data is realistic
- [ ] Test is independent of other tests
- [ ] Test runs quickly (<1s for unit tests)
- [ ] Test has appropriate assertions
- [ ] Test cleanup is handled properly
- [ ] Test is properly categorized (unit/integration/e2e)

## Troubleshooting

### Common Issues

1. **Flaky Tests**: Use `t.Parallel()` carefully, ensure proper isolation
2. **Slow Tests**: Move to integration tests, use smaller fixtures
3. **Coverage Gaps**: Add table-driven tests for edge cases
4. **Race Conditions**: Always run with `-race` flag in CI

### Test Utilities

```bash
# Find slowest tests
go test -json ./... | go-test-report

# Generate test report
go test -json ./... > test.json
go-test-viewer < test.json

# Watch tests
watch -n 2 'go test ./...'
```

## Conclusion

Comprehensive testing is essential for pyqol's success. Follow these guidelines to ensure reliable, maintainable, and performant tests.