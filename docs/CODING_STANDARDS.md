# Coding Standards

This document defines the coding standards for the pyscn project. All contributors must follow these guidelines to ensure consistency and maintainability.

## Go Code Style

### General Principles

1. **Clarity over cleverness**: Write code that is easy to understand
2. **Consistency**: Follow existing patterns in the codebase
3. **Simplicity**: Avoid unnecessary complexity
4. **Documentation**: Document why, not what

### Naming Conventions

#### Packages

- Use lowercase, single-word names
- No underscores or mixedCaps
- Be concise but descriptive

```go
// Good
package parser
package analyzer
package config

// Bad
package parserUtils
package dead_code
package cfg_builder
```

#### Files

- Use lowercase with underscores for multi-word files
- Test files must end with `_test.go`
- Benchmark files should include `_bench` suffix

```go
// Good
parser.go
dead_code.go
cfg_builder.go
parser_test.go
apted_bench_test.go

// Bad
Parser.go
deadCode.go
CFGBuilder.go
```

#### Variables and Functions

- Use camelCase for local variables
- Use PascalCase for exported functions and types
- Use meaningful, descriptive names
- Avoid single-letter variables except for loops and very short scopes

```go
// Good
func ParsePythonFile(filename string) (*AST, error)
func buildCFG(ast *AST) *CFG
nodeCount := len(nodes)

// Bad
func parse_python_file(fn string) (*AST, error)
func BuildCfg(a *AST) *CFG
n := len(nodes) // unclear what 'n' represents
```

#### Constants

- Use PascalCase for exported constants
- Use camelCase for unexported constants
- Group related constants with `iota`

```go
// Good
const (
    MaxDepth = 100
    DefaultTimeout = 30 * time.Second
)

const (
    stateIdle = iota
    stateRunning
    stateComplete
)

// Bad
const MAX_DEPTH = 100
const default_timeout = 30
```

#### Interfaces

- Use "-er" suffix for single-method interfaces
- Keep interfaces small and focused
- Define interfaces where they are used, not where implemented

```go
// Good
type Parser interface {
    Parse([]byte) (*AST, error)
}

type Walker interface {
    Walk(Node) error
}

// Bad
type ParserInterface interface {
    Parse([]byte) (*AST, error)
    ParseFile(string) (*AST, error)
    ParseString(string) (*AST, error)
    // too many methods
}
```

### Code Organization

#### Package Structure

```go
// Each file should have this structure:
// 1. Package declaration
// 2. Imports (grouped)
// 3. Constants
// 4. Types
// 5. Constructor/New functions
// 6. Methods (receiver functions)
// 7. Functions
// 8. Helper functions (unexported)

package analyzer

import (
    "context"
    "fmt"
    
    "github.com/ludo-technologies/pyscn/internal/parser"
    "github.com/ludo-technologies/pyscn/pkg/api"
    
    "github.com/third-party/lib"
)

const defaultThreshold = 0.8

type Analyzer struct {
    threshold float64
}

func NewAnalyzer(threshold float64) *Analyzer {
    return &Analyzer{threshold: threshold}
}

func (a *Analyzer) Analyze(ast *parser.AST) ([]Finding, error) {
    // implementation
}

func processNode(node *parser.Node) error {
    // helper function
}
```

#### Import Grouping

Group imports in the following order, separated by blank lines:

1. Standard library
2. Internal packages
3. Third-party packages

```go
import (
    "context"
    "fmt"
    "io"
    
    "github.com/ludo-technologies/pyscn/internal/parser"
    "github.com/ludo-technologies/pyscn/internal/analyzer"
    
    "github.com/spf13/cobra"
    "github.com/stretchr/testify/assert"
)
```

### Error Handling

#### Error Messages

- Start with lowercase letter
- No punctuation at the end
- Include context

```go
// Good
return fmt.Errorf("failed to parse file %s: %w", filename, err)
return errors.New("unexpected end of input")

// Bad
return fmt.Errorf("Failed to parse file.")
return errors.New("ERROR!!!")
```

#### Error Wrapping

Always wrap errors with context:

```go
// Good
if err != nil {
    return fmt.Errorf("parsing Python file: %w", err)
}

// Bad
if err != nil {
    return err // loses context
}
```

#### Error Types

Define custom error types for domain-specific errors:

```go
type ParseError struct {
    File   string
    Line   int
    Column int
    Msg    string
}

func (e *ParseError) Error() string {
    return fmt.Sprintf("%s:%d:%d: %s", e.File, e.Line, e.Column, e.Msg)
}
```

### Comments and Documentation

#### Package Comments

Every package must have a package comment:

```go
// Package parser implements Python code parsing using tree-sitter.
// It provides functionality to parse Python source code and build
// an abstract syntax tree (AST) for further analysis.
package parser
```

#### Function Comments

Export functions must have comments starting with the function name:

```go
// ParseFile parses a Python source file and returns an AST.
// It returns an error if the file cannot be read or parsed.
func ParseFile(filename string) (*AST, error) {
    // implementation
}
```

#### Inline Comments

- Explain why, not what
- Keep comments up-to-date
- Remove commented-out code

```go
// Good
// Use a larger buffer for performance on large files
buffer := make([]byte, 8192)

// Bad
// Increment i
i++ // obvious from code
```

### Testing

#### Test Names

Use descriptive test names that explain what is being tested:

```go
// Good
func TestParser_ParsesValidPythonCode(t *testing.T)
func TestCFG_HandlesNestedLoops(t *testing.T)
func TestAPTED_CalculatesCorrectDistance(t *testing.T)

// Bad
func TestParse(t *testing.T)
func Test1(t *testing.T)
func TestCFG(t *testing.T)
```

#### Table-Driven Tests

Use table-driven tests for testing multiple scenarios:

```go
func TestCalculateComplexity(t *testing.T) {
    tests := []struct {
        name     string
        input    *CFG
        expected int
    }{
        {
            name:     "empty function",
            input:    &CFG{},
            expected: 1,
        },
        {
            name:     "single if statement",
            input:    buildCFGWithIf(),
            expected: 2,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CalculateComplexity(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

#### Test Helpers

Extract common test setup into helper functions:

```go
func setupTestParser(t *testing.T) *Parser {
    t.Helper()
    parser := NewParser()
    // common setup
    return parser
}

func assertNoError(t *testing.T, err error) {
    t.Helper()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

### Performance

#### Preallocate Slices

When the size is known, preallocate slices:

```go
// Good
nodes := make([]*Node, 0, len(items))
for _, item := range items {
    nodes = append(nodes, processItem(item))
}

// Bad
var nodes []*Node
for _, item := range items {
    nodes = append(nodes, processItem(item))
}
```

#### Use Strings Builder

For string concatenation in loops:

```go
// Good
var sb strings.Builder
for _, s := range strings {
    sb.WriteString(s)
}
result := sb.String()

// Bad
result := ""
for _, s := range strings {
    result += s // creates new string each time
}
```

#### Defer Expensive Operations

```go
// Good
func processFile(filename string) error {
    if !needsProcessing(filename) {
        return nil // early return before expensive operations
    }
    
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    // expensive processing
}
```

### Concurrency

#### Goroutine Lifecycle

Always ensure goroutines can terminate:

```go
// Good
func worker(ctx context.Context, jobs <-chan Job) {
    for {
        select {
        case <-ctx.Done():
            return
        case job, ok := <-jobs:
            if !ok {
                return
            }
            process(job)
        }
    }
}

// Bad
func worker(jobs <-chan Job) {
    for job := range jobs { // no way to stop
        process(job)
    }
}
```

#### Synchronization

Use appropriate synchronization primitives:

```go
// For shared state
type SafeCounter struct {
    mu    sync.RWMutex
    count int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

func (c *SafeCounter) Value() int {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.count
}
```

### Resource Management

#### Always Close Resources

```go
// Good
file, err := os.Open(filename)
if err != nil {
    return err
}
defer file.Close()

// For resources that return errors on close
func processFile(filename string) (err error) {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer func() {
        if cerr := file.Close(); cerr != nil && err == nil {
            err = cerr
        }
    }()
    
    // process file
    return nil
}
```

#### Context Usage

Use context for cancellation and timeouts:

```go
func analyze(ctx context.Context, files []string) error {
    for _, file := range files {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            if err := analyzeFile(ctx, file); err != nil {
                return err
            }
        }
    }
    return nil
}
```

## Code Review Checklist

Before submitting a PR, ensure:

- [ ] Code follows naming conventions
- [ ] All exported functions have comments
- [ ] Errors are wrapped with context
- [ ] Tests cover new functionality
- [ ] No commented-out code
- [ ] No TODO comments (create issues instead)
- [ ] Resources are properly closed
- [ ] Concurrent code is safe
- [ ] Performance considerations addressed
- [ ] Code is formatted with `gofmt`
- [ ] Passes `go vet`
- [ ] No lint warnings

## Tools

### Required Tools

```bash
# Format code
go fmt ./...

# Check for issues
go vet ./...

# Run tests
go test ./...

# Test with race detection
go test -race ./...
```

### Recommended Tools

```bash
# Install tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run linter
golangci-lint run

# Fix imports
goimports -w .

# Security scan
gosec ./...
```

## Git Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Test changes
- `build`: Build system changes
- `ci`: CI/CD changes
- `chore`: Maintenance tasks

### Examples

```bash
feat(parser): add support for Python 3.11 syntax

Add support for:
- Exception groups (PEP 654)
- Task groups
- New typing features

Closes #123

---

fix(cfg): correctly handle break statements in nested loops

The CFG builder was not properly connecting break statements
to the correct loop exit when multiple loops were nested.

Fixes #456

---

perf(apted): optimize tree comparison for large ASTs

- Use memoization for subtree comparisons
- Implement early termination for identical subtrees
- Reduce memory allocations

Benchmark results:
- Before: 2.5s for 1000-node tree
- After: 0.8s for 1000-node tree
```

## Exceptions

While these standards should be followed in general, there may be cases where deviation is justified. In such cases:

1. Document the reason for deviation
2. Get team consensus
3. Be consistent within the module

## Updates

These coding standards are living documents and will be updated as the project evolves. Propose changes through pull requests with clear justification.