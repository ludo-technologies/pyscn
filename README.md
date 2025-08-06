# pyqol - Python Quality of Life

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Status](https://img.shields.io/badge/Status-Pre--Alpha-orange.svg)](https://github.com/pyqol/pyqol)

**pyqol** is a next-generation static analysis tool for Python that goes beyond traditional linters. By leveraging Control Flow Graph (CFG) analysis and tree edit distance algorithms (APTED), pyqol provides deep structural insights into your codebase.

## 🚀 Features (Coming Soon)

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

```bash
# Coming soon!
go install github.com/pyqol/pyqol/cmd/pyqol@latest
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