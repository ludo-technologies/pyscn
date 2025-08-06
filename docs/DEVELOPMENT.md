# Development Guide

This guide covers everything you need to know to contribute to pyqol development.

## Table of Contents

- [Getting Started](#getting-started)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Task Management](#task-management)
- [Building and Testing](#building-and-testing)
- [Code Style](#code-style)
- [Git Workflow](#git-workflow)

## Getting Started

### Prerequisites

- Go 1.22+ (recommended: 1.24)
- Git
- GitHub CLI (`gh`)
- Make (optional but recommended)

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/pyqol/pyqol.git
cd pyqol

# Install dependencies
go mod download

# Run tests to verify setup
go test ./...

# Build the binary
go build ./cmd/pyqol
```

## Project Structure

```
pyqol/
├── cmd/
│   └── pyqol/         # CLI entry point
│       └── main.go    # Main function
├── internal/          # Private packages
│   ├── parser/        # Tree-sitter integration
│   │   ├── python.go  # Python-specific parsing
│   │   └── ast.go     # AST definitions
│   ├── analyzer/      # Analysis algorithms
│   │   ├── cfg.go     # Control Flow Graph
│   │   ├── dead.go    # Dead code detection
│   │   └── apted.go   # Clone detection
│   └── config/        # Configuration
├── pkg/               # Public packages
│   └── api/           # Public API
├── testdata/          # Test fixtures
├── docs/              # Documentation
└── scripts/           # Utility scripts
```

## Development Workflow

### 1. Pick a Task

```bash
# View available tasks for current week
./scripts/tasks.sh week 1

# View task details
./scripts/tasks.sh view 1

# Assign task to yourself
./scripts/tasks.sh start 1
```

### 2. Create a Feature Branch

Follow our branching strategy (see [BRANCHING.md](BRANCHING.md)):

```bash
# Create branch from main
git checkout main
git pull origin main
git checkout -b feature/issue-1-tree-sitter-integration

# Branch naming patterns:
# feature/issue-{number}-{description}  # New features
# fix/issue-{number}-{description}      # Bug fixes
# docs/{description}                    # Documentation
# refactor/{description}                # Code improvements
# chore/{description}                   # Maintenance
```

### 3. Implement the Feature

Follow the implementation checklist:
- [ ] Write tests first (TDD approach)
- [ ] Implement the feature
- [ ] Ensure all tests pass
- [ ] Add documentation
- [ ] Run linters

### 4. Submit Pull Request

```bash
# Push your branch
git push origin feature/issue-1-tree-sitter

# Create PR via GitHub CLI
gh pr create --title "feat: Add tree-sitter integration (#1)" \
  --body "Closes #1" \
  --milestone "Week 1 - Foundation"
```

## Task Management

### GitHub Issues

All development tasks are tracked as GitHub Issues with the following labels:

- **Priority**: `P0` (Critical), `P1` (High), `P2` (Medium), `P3` (Low)
- **Type**: `task`, `bug`, `enhancement`, `documentation`
- **Component**: `parser`, `analyzer`, `cli`

### Milestones

Development is organized into weekly sprints:

- **Week 1 - Foundation**: Core parsing and CFG
- **Week 2 - Dead Code**: Dead code detection
- **Week 3 - Clone Detection**: APTED implementation
- **Week 4 - Release**: CLI and release preparation

### Task Commands

```bash
# List all tasks
gh issue list

# View tasks by milestone
gh issue list --milestone "Week 1 - Foundation"

# View tasks by label
gh issue list --label "P0"

# Update task status
gh issue comment 1 --body "Progress update: Completed AST implementation"

# Close completed task
gh issue close 1 --comment "Implemented in PR #10"
```

## Building and Testing

### Build Commands

```bash
# Build for current platform
go build -o pyqol ./cmd/pyqol

# Build for all platforms
GOOS=linux GOARCH=amd64 go build -o pyqol-linux-amd64 ./cmd/pyqol
GOOS=darwin GOARCH=amd64 go build -o pyqol-darwin-amd64 ./cmd/pyqol
GOOS=windows GOARCH=amd64 go build -o pyqol-windows-amd64.exe ./cmd/pyqol

# Build with version info
go build -ldflags "-X main.version=v0.1.0" ./cmd/pyqol
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run specific package tests
go test ./internal/parser

# Run tests with verbose output
go test -v ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Benchmarking

```bash
# Run benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkCFG ./internal/analyzer

# Benchmark with memory profiling
go test -bench=. -benchmem ./...
```

## Code Style

### Go Conventions

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Use meaningful variable names
- Keep functions small and focused
- Document exported functions

### Code Quality Tools

```bash
# Format code
go fmt ./...

# Vet code for issues
go vet ./...

# Run golangci-lint (if installed)
golangci-lint run

# Check for security issues
gosec ./...
```

### Testing Standards

- Write table-driven tests
- Use meaningful test names
- Test edge cases
- Aim for >80% code coverage
- Mock external dependencies

Example test structure:

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr bool
    }{
        {
            name:  "valid input",
            input: InputType{...},
            want:  OutputType{...},
        },
        {
            name:    "invalid input",
            input:   InputType{...},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("FunctionName() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("FunctionName() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Git Workflow

### Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification (detailed in [BRANCHING.md](BRANCHING.md)):

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `refactor`: Code refactoring
- `test`: Testing changes
- `docs`: Documentation
- `perf`: Performance improvements
- `ci`: CI/CD changes
- `chore`: Maintenance tasks

Examples:
```bash
git commit -m "feat(parser): add tree-sitter Python integration"
git commit -m "fix(cfg): handle break statements in loops"
git commit -m "test(analyzer): add benchmarks for APTED algorithm"
```

### Pull Request Process

1. **Create focused PRs**: One feature/fix per PR
2. **Write descriptive PR titles**: Include issue number
3. **Fill out PR template**: Describe changes and testing
4. **Request reviews**: Tag relevant maintainers
5. **Address feedback**: Respond to all comments
6. **Keep PRs updated**: Rebase on main if needed

### Code Review Guidelines

When reviewing PRs:
- Check test coverage
- Verify documentation updates
- Run code locally
- Provide constructive feedback
- Approve when satisfied

## Continuous Integration

GitHub Actions runs the following checks on every PR:

- Go 1.22 and 1.23 compatibility
- Unit tests
- Race condition detection
- Code coverage reporting
- Linting (go vet)
- Build verification

## Performance Considerations

### Benchmarking Targets

- **Parsing**: >100,000 lines/second
- **CFG Construction**: >10,000 lines/second
- **APTED Comparison**: <1 second for 1000-node trees
- **Memory Usage**: <10x file size

### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof

# Generate flame graph
go test -cpuprofile=cpu.prof -bench=.
go tool pprof -http=:8080 cpu.prof
```

## Debugging

### Debug Build

```bash
# Build with debug symbols
go build -gcflags="all=-N -l" ./cmd/pyqol

# Run with debug logging
PYQOL_DEBUG=1 ./pyqol analyze test.py

# Use delve debugger
dlv debug ./cmd/pyqol -- analyze test.py
```

### Logging

Use structured logging for debugging:

```go
import "log/slog"

slog.Debug("parsing file", 
    "file", filename,
    "size", fileSize,
    "duration", duration)
```

## Release Process

Releases are automated via GitHub Actions when a tag is pushed:

```bash
# Create and push a tag
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0

# This triggers:
# 1. Build for all platforms
# 2. Run full test suite
# 3. Create GitHub release
# 4. Upload binaries
```

## Getting Help

- **Issues**: [GitHub Issues](https://github.com/pyqol/pyqol/issues)
- **Discussions**: [GitHub Discussions](https://github.com/pyqol/pyqol/discussions)
- **Documentation**: This guide and `/docs` directory

## Quick Reference

```bash
# Development cycle
./scripts/tasks.sh list           # View tasks
git checkout -b feature/issue-N   # Start feature
go test ./...                      # Test changes
git commit -m "feat: ..."          # Commit
gh pr create                       # Create PR
./scripts/tasks.sh done N          # Close issue
```

Happy coding! 🚀