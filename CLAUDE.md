# Claude Code Development Instructions

This file contains specific instructions for Claude Code when working on the pyqol project.

## Additional Documentation

For detailed information, reference these documents using @:
- `@docs/DEVELOPMENT.md` - Development workflow and setup
- `@docs/ARCHITECTURE.md` - System design and components
- `@docs/CODING_STANDARDS.md` - Go coding conventions
- `@docs/TESTING.md` - Testing strategies and patterns
- `@docs/ROADMAP.md` - Development roadmap and milestones

## Project Overview

pyqol is a next-generation Python static analysis tool that uses Control Flow Graph (CFG) and APTED (tree edit distance) algorithms to provide deep code quality insights beyond traditional linters.

## Development Language

**ALL development must be conducted in English**, including:
- Code comments
- Commit messages
- Documentation
- Variable and function names
- Test descriptions
- Issue and PR descriptions

## Project Goals

### Primary Objectives
1. Build a high-performance Python analyzer in Go
2. Implement CFG-based dead code detection
3. Implement APTED-based clone detection
4. Complete MVP in 1 month (by September 6, 2025)

### Technical Requirements
- Performance: >10,000 lines/second for analysis
- Accuracy: <5% false positive rate
- Memory: <10x file size usage
- Go version: 1.24 for development, 1.22+ for CI compatibility

## Development Workflow

### Task Management
1. Check current tasks: `./scripts/tasks.sh list`
2. Start a task: `./scripts/tasks.sh start <issue-number>`
3. Create feature branch: `git checkout -b feature/issue-<number>-<description>`
4. Implement with TDD approach
5. Commit with conventional commits format
6. Create PR and link to issue

### Code Style Requirements
- Follow Go idioms and best practices
- Use meaningful variable names (no single letters except in loops)
- Write tests FIRST (TDD)
- Keep functions small (<50 lines)
- Document "why" not "what"
- No TODO comments in code (create issues instead)

### Testing Requirements
- Write table-driven tests
- Achieve >80% code coverage
- Test edge cases and error conditions
- Run tests before committing: `go test -race ./...`
- Add benchmarks for performance-critical code

## Current Sprint (Week 1: Aug 6-12, 2025)

### Priority Tasks
1. **Issue #1**: Tree-sitter Go binding integration
2. **Issue #2**: Python AST construction  
3. **Issue #3**: CFG algorithm implementation

### Focus Areas
- Set up tree-sitter with Python grammar
- Build robust AST representation
- Implement CFG construction algorithm
- Ensure cross-platform compatibility

## Architecture Decisions

### Core Technologies
- **Parser**: tree-sitter (github.com/smacker/go-tree-sitter)
- **CLI**: Cobra framework
- **Testing**: testify for assertions
- **Config**: YAML with viper

### Package Structure
```
internal/
  parser/     # Tree-sitter integration (private)
  analyzer/   # CFG and APTED algorithms (private)
  config/     # Configuration management (private)
pkg/
  api/        # Public API (if needed)
```

### Performance Targets
- Parsing: >100,000 lines/second
- CFG construction: >10,000 lines/second  
- APTED comparison: <1 second for 1000 nodes
- Memory usage: <100MB for 100k LOC

## Implementation Guidelines

### Parser Module
- Use tree-sitter-python grammar
- Support Python 3.8+ syntax
- Handle syntax errors gracefully
- Build clean AST representation

### CFG Module
- Handle all Python control flow constructs
- Support nested structures
- Track variable definitions and uses
- Identify unreachable code accurately

### APTED Module
- Implement efficient tree edit distance
- Support configurable cost models
- Optimize for large ASTs
- Provide similarity threshold configuration

## Quality Checklist

Before committing code, ensure:
- [ ] All tests pass: `go test ./...`
- [ ] No race conditions: `go test -race ./...`
- [ ] Code formatted: `go fmt ./...`
- [ ] No vet issues: `go vet ./...`
- [ ] Coverage >80%: `go test -cover ./...`
- [ ] Documentation updated
- [ ] Meaningful commit message

## Common Commands

```bash
# Development
go test ./...                    # Run all tests
go test -race ./...              # Test with race detection
go test -cover ./...             # Check coverage
go test -bench=. ./...           # Run benchmarks
go build ./cmd/pyqol             # Build binary

# Task management
./scripts/tasks.sh list          # View all tasks
./scripts/tasks.sh week 1        # View current week tasks
./scripts/tasks.sh start N       # Start working on issue N
./scripts/tasks.sh done N        # Complete issue N

# Git workflow
git checkout -b feature/issue-N  # Create feature branch
git commit -m "feat: ..."        # Conventional commit
gh pr create                     # Create pull request
```

## Error Handling

- Always wrap errors with context
- Use custom error types for domain errors
- Handle all error cases explicitly
- Provide helpful error messages
- Never panic in library code

## Testing Approach

1. **Unit tests**: Test individual functions
2. **Integration tests**: Test component interactions
3. **E2E tests**: Test CLI commands
4. **Benchmarks**: Monitor performance
5. **Fuzz tests**: Find edge cases

## Documentation

- Document all exported functions
- Include examples in documentation
- Keep README.md updated
- Update CHANGELOG.md for releases
- Write clear commit messages

## Performance Considerations

- Profile before optimizing
- Use benchmarks to track performance
- Prefer simple solutions
- Avoid premature optimization
- Consider memory allocation

## Security

- Validate all inputs
- Prevent path traversal
- Limit resource usage
- Handle malformed Python code safely
- Don't execute Python code

## Communication

- Be concise and direct
- Provide specific technical details
- Suggest solutions, not just problems
- Include relevant code examples
- Reference issue numbers

## Constraints

- No external dependencies without justification
- Maintain backward compatibility
- Keep binary size reasonable (<50MB)
- Support Linux, macOS, Windows
- Work with Python 3.8+

## Success Metrics

- Analysis speed: >10,000 lines/second
- False positive rate: <5%
- Memory usage: <10x file size
- User satisfaction: High
- Code quality: Excellent

## Remember

1. **Quality over quantity** - Better to do fewer things well
2. **Test everything** - Untested code is broken code
3. **Performance matters** - This tool must be fast
4. **User experience** - Make it easy to use
5. **Open source** - Build in public, welcome contributions

## Questions to Ask

When implementing features, consider:
- Is this the simplest solution?
- Have I written tests first?
- Will this scale to large codebases?
- Is the error handling comprehensive?
- Does this follow Go conventions?

## Final Notes

The goal is to create a tool that Python developers will love to use. Focus on accuracy, performance, and user experience. Make pyqol the go-to tool for Python code quality analysis.

---

*Last updated: August 6, 2025*
*Project deadline: September 6, 2025*