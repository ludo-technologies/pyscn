# Contributing to pyscn

First off, thank you for considering contributing to pyscn! It's people like you that make pyscn such a great tool.

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues as you might find out that you don't need to create one. When you are creating a bug report, please include as many details as possible using our issue template.

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. Create an issue using the feature request template and provide the following information:

- Use a clear and descriptive title
- Provide a step-by-step description of the suggested enhancement
- Provide specific examples to demonstrate the steps
- Describe the current behavior and explain which behavior you expected to see instead

### Your First Code Contribution

Unsure where to begin contributing? You can start by looking through these `good-first-issue` and `help-wanted` issues:

- [Good first issues](https://github.com/pyscn/pyscn/labels/good%20first%20issue) - issues which should only require a few lines of code
- [Help wanted issues](https://github.com/pyscn/pyscn/labels/help%20wanted) - issues which should be a bit more involved

## Development Process

1. Fork the repo and create your branch from `main`
2. If you've added code that should be tested, add tests
3. If you've changed APIs, update the documentation
4. Ensure the test suite passes
5. Make sure your code follows the existing style
6. Issue that pull request!

## Setting Up Your Development Environment

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/pyscn.git
cd pyscn

# Add upstream remote
git remote add upstream https://github.com/pyscn/pyscn.git

# Install dependencies
go mod download

# Run tests
go test ./...

# Build the project
go build ./cmd/pyscn
```

## Project Structure

```
pyscn/
├── cmd/pyscn/       # CLI entry point
├── internal/        # Private packages
│   ├── parser/      # Tree-sitter integration
│   ├── analyzer/    # CFG and APTED implementations
│   └── config/      # Configuration management
├── pkg/             # Public packages
├── testdata/        # Test fixtures
└── docs/            # Documentation
```

## Testing

- Write unit tests for all new functionality
- Ensure all tests pass: `go test ./...`
- Run with race detection: `go test -race ./...`
- Check coverage: `go test -cover ./...`

## Code Style

- Follow standard Go conventions
- Use `gofmt` to format your code
- Use `golint` and `go vet` to check for issues
- Keep functions small and focused
- Write clear, self-documenting code
- Add comments for complex logic

## Commit Messages

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests liberally after the first line

Example:
```
Add CFG-based dead code detection

- Implement control flow graph construction
- Add reachability analysis algorithm
- Include comprehensive test cases

Fixes #123
```

## Pull Request Process

1. Update the README.md with details of changes if applicable
2. Update documentation for any changed functionality
3. The PR will be merged once you have the sign-off of at least one maintainer

## Release Process

We use semantic versioning (SemVer) for releases:
- MAJOR version for incompatible API changes
- MINOR version for backwards-compatible functionality additions
- PATCH version for backwards-compatible bug fixes

## Questions?

Feel free to open an issue with your question or reach out to the maintainers directly.

## Recognition

Contributors will be recognized in our README and release notes. Thank you for your contributions!