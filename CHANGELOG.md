# Changelog

## [1.2.0] - 2025-10-18

### Added
- MCP (Model Context Protocol) server integration for AI coding assistants (#184)

### Improved
- Package pyscn-mcp binary in wheels for uv workflows (#189)
- Add MCP integration documentation section (#190)
- Refactor file resolution logic into shared helper (#183)
- Add Python 3.13 support (#174)
- Improve build system and MCP documentation (#191)
- Update issue templates for user feedback (#192)

## [1.1.1] - 2025-10-12

### Fixed
- **Critical:** Fix `pyscn analyze` failing with "no Python files found" error (#178)

### Improved
- Improve README with dev.to article link and streamline CI/CD section (#172)

## [1.1.0] - 2025-10-11

### Added
- `file:line:col` output format to `check` command for better editor integration (#168)
- `--select` flag to `check` command for running specific analyses (#159)

### Enhanced
- Embed default config as TOML and load architecture rules from config (#163)

### Fixed
- Prevent browser from opening HTML reports when connected via SSH (#154)
- Fix Makefile ASCII escape codes on cross-platform (#153)

### Improved
- Add Dependabot configuration for Go dependency updates (#165)
- Improve issue template readability (#169)

## [1.0.3] - 2025-10-10

### Fixed
- Improve progress estimation for clone detection with LSH (#161)
- Support for `**` globstar patterns in include/exclude patterns (#152)

### Note
- Pattern matching now uses `doublestar` library for proper `**` support
- If you use custom patterns, update `*.py` to `**/*.py` for recursive matching

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