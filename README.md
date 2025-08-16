# pyqol - Python Quality of Life

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Status](https://img.shields.io/badge/Status-Pre--Alpha-orange.svg)](https://github.com/pyqol/pyqol)

**pyqol is a static analysis tool focused on software design and architectural quality, going beyond traditional linting.**

While generative AI excels at writing code, it can struggle with maintaining a clean architecture. `pyqol` acts as your AI-powered teammate, specifically designed to detect structural issues that linters often miss, such as:

- **Low Cohesion & High Coupling:** Identifying classes and modules that are hard to maintain, test, and reuse (using metrics like LCOM4 and CBO).
- **Architectural Principle Violations:** Detecting circular dependencies and deviations from principles like SOLID.
- **Structural Duplication:** Finding structurally similar code blocks that could be refactored (using tree-edit distance algorithms).

Instead of just checking for style, `pyqol` helps you build a more robust, maintainable, and scalable codebase.

## 🚀 Features

### Core Technologies
- **CFG-based Dead Code Detection**: Precisely identifies unreachable code using control flow analysis
- **APTED Clone Detection**: Finds structurally similar code blocks using tree edit distance
- **Complexity Metrics**: Calculates cyclomatic complexity and LCOM4 cohesion metrics

### Why pyqol?

| Tool | Focus | pyqol's Advantage |
|------|-------|-------------------|
| Ruff | Speed & Rule Coverage | Deep structural analysis with CFG |
| Pylint | Comprehensive Rules | Intelligent clone detection with APTED |
| mypy | Type Checking | Design quality metrics (cohesion, coupling) |
| Bandit | Security | Architecture and design improvements |

## 🗓️ Development Roadmap

### Sprint 1: MVP (1 Month) - September 2025
- [x] Requirements Definition
- [ ] Tree-sitter Integration
- [ ] CFG Implementation
- [ ] APTED Algorithm
- [ ] CLI Interface
- [ ] **Open Source Release** 🎉

### Sprint 2: Extended Features
- [ ] Dependency Analysis
- [ ] Performance Patterns
- [ ] VS Code Extension

### Sprint 3: Pro Features
- [ ] LLM Integration
- [ ] Advanced Suggestions
- [ ] Team Features

## 🛠️ Technology Stack

- **Language**: Go 1.21+
- **Parser**: Tree-sitter
- **Algorithms**: CFG, APTED
- **CLI**: Cobra

## 📦 Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/pyqol/pyqol.git
cd pyqol

# Build with Make
make build

# Or install directly with Go
go install github.com/pyqol/pyqol/cmd/pyqol@latest
```

## 🔨 Development

### Prerequisites

- Go 1.22+ (recommended: 1.24)
- Make (optional but recommended)

### Quick Start

```bash
# Build the project
make build

# Run tests
make test

# Run with hot reload (requires air)
make dev

# See all available commands
make help
```

### Common Make Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the binary |
| `make test` | Run tests with race detection |
| `make bench` | Run benchmarks |
| `make coverage` | Generate coverage report |
| `make fmt` | Format code |
| `make lint` | Run linters |
| `make clean` | Clean build artifacts |
| `make install` | Install the binary |
| `make version` | Show version information |
| `make build-all` | Build for all platforms |
| `make release VERSION=v0.1.0` | Create a new release |

### Building from Source

```bash
# Simple build
make build

# Build with specific version
make build VERSION=v0.1.0

# Build for all platforms
make build-all

# Manual build with version info
go build -ldflags "-X github.com/pyqol/pyqol/internal/version.Version=v0.1.0" \
         -o pyqol ./cmd/pyqol
```

## 🤝 Contributing

We're building pyqol in public! While we're still in early development, we welcome:

- 🐛 Bug reports
- 💡 Feature suggestions
- 📖 Documentation improvements
- 🔧 Code contributions

Please see our [Contributing Guide](docs/CONTRIBUTING.md) (coming soon).

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🌟 Star History

Give us a star if you're excited about the future of Python code quality!

---

**Note**: pyqol is currently in active development. The first public release is targeted for September 6, 2025.

Built with ❤️ by the pyqol team