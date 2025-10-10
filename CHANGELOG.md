# Changelog

## [1.0.2] - 2025-10-10

### Enhanced
- Add nesting depth metric to complexity analysis (#150)

### Fixed
- **Critical:** Fix `pyscn check .` and `pyscn analyze .` failing with "no Python files found" (#147)
- Ignore circular dependencies for imports inside TYPE_CHECKING (#151)

### Improved
- Add comprehensive test coverage for error categorizer (#142)
- Improve README structure and installation section (#148)

## [1.0.1] - 2025-10-07

### Fixed
- Fixed `pyscn init` missing `[output]` and `[analysis]` sections in generated config (#131)
- Fixed HTML reports to sort by complexity instead of name by default (#130)

## [1.0.0] - 2025-10-05

### 🎉 First Stable Release

High-performance Python code quality analyzer built with Go.
Designed for the AI-assisted development era.

### Key Features

- **Blazing Fast** - 100,000+ lines/sec with Go + tree-sitter
- **Advanced Clone Detection** - APTED algorithm with LSH acceleration
- **Dead Code Analysis** - CFG-based unreachable code detection
- **Architecture Metrics** - CBO coupling and cyclomatic complexity
- **Multiple Formats** - HTML, JSON, YAML, CSV reports
- **CI/CD Ready** - `pyscn check` for quality gates

### Installation

```bash
pipx install pyscn        # Recommended
uvx pyscn analyze .       # Run without install
```

### Quick Start

```bash
pyscn analyze .           # Full analysis with HTML report
pyscn check .            # Quick CI check
pyscn init              # Generate config file
```

---

## Previous Releases

Beta versions (0.1.0-beta.1 through 0.8.0-beta.1) were development releases.
For details, see git commit history.