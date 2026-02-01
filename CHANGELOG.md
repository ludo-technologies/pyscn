# Changelog

## [1.9.2] - 2026-02-01

### Fixed
- Prevent Type-3 clone false positives for different class/function names (#313)
- Reduce clone detection false positives for framework patterns (#312)
- Exclude test modules from layer detection (#311)

## [1.9.1] - 2026-01-28

### Fixed
- Reduce false positives in Type-3/Type-4 clone detection (#305)

## [1.9.0] - 2026-01-27

### Added
- Add `pyscn-mcp` skill for MCP tools integration (#291, #303)

### Changed
- Enable Type-2 clone detection by default with Jaccard coefficient (#302)

## [1.8.2] - 2026-01-22

### Fixed
- Remove MaxComplexity filtering from service layer (#298)
- Check command now respects .pyscn.toml settings (#296)
- Consolidate default settings to single source of truth (#294)

## [1.8.1] - 2026-01-20

### Fixed
- Fix marketplace plugin source path for GitHub registration (#288, #289)

## [1.8.0] - 2026-01-20

### Added
- Claude Code plugin marketplace support (#281)

### Fixed
- Improve duplication scoring with K-Core groups (#286)

### Changed
- Bump the go-dependencies group with 4 updates (#275)

## [1.7.1] - 2026-01-15

### Fixed
- Fix reachability analysis exponential time complexity with memoization (#283)

## [1.7.0] - 2026-01-14

### Added
- Mock data detection with `--select mockdata` flag (#278)

## [1.6.0] - 2026-01-06

### Added
- Pyscn Bot (GitHub App) for automated PR reviews and weekly code audits (#276)

### Changed
- Consolidate default CBO thresholds to single source of truth (#274)
- Consolidate default CloneTypes to single source of truth (#273)

### Improved
- Service layer test coverage (36.3% → 53.7%) (#269, #270)

## [1.5.5] - 2025-12-15

### Fixed
- Improve clone detection similarity calculation to reduce false positives (#266)

### Changed
- Adjust duplication threshold from 0-20% to 0-10% (#267)

## [1.5.4] - 2025-12-15

### Improved
- Change duplication rate calculation from line-based to function-based for consistent measurements (#264)

## [1.5.3] - 2025-12-14

### Fixed
- Disable Type2 clone detection by default and fix clone grouping (#262)

## [1.5.2] - 2025-12-13

### Fixed
- Skip docstrings from clone detection to reduce false positives (#258)

## [1.5.1] - 2025-12-12

### Fixed
- Enable architecture strict mode by default (#255)

### Improved
- Consolidate default configuration values in `domain/defaults.go` (#254)

### Removed
- macOS Intel (x86_64) build support (#259)

## [1.5.0] - 2025-12-08

### Added
- Data Flow Analysis (DFA) for Type-4 clone detection with `--enable-dfa` flag (#250)
- Multi-dimensional clone type classification (#248)

### Fixed
- Ensure deterministic architecture analysis results (#252)
- Count skipped files as analyzed (#243)
- Replace hardcoded thresholds with constants in CentroidGrouping (#247)

## [1.4.2] - 2025-11-30

### Fixed
- Route break/continue/raise through finally blocks correctly in CFG (#210)

### Improved
- Add comprehensive tests for apted_tree.go (#238)
- Add jscan link to README (#234)
- Bump Go dependencies (#232)
- Remove unused labels from dependabot.yml (#240)

## [1.4.1] - 2025-11-27

### Fixed
- Fix binary auto-detection for `pyscn_mcp` package in wheel build script (#224)

### Improved
- Remove Viper-based config loader and unify on TOML-only (#231)

## [1.4.0] - 2025-11-25

### Added
- Circular dependency detection in `check` command with `--select deps`, `--allow-circular-deps`, and `--max-cycles` flags (#213)
- Universal TOML configuration support for all pyscn sections in `.pyscn.toml` and `pyproject.toml` (#229)

### Fixed
- Complexity thresholds from `.pyscn.toml` were being ignored (#226)

## [1.3.0] - 2025-11-06

### Fixed
- Fix finally clause parsing in try-except-finally statements (#209)
- Fix score normalization for dependencies and architecture categories (#221)

### Added
- Circular dependency details in HTML reports (#176)
- Scoring reference documentation (#211)
- MCP use case examples in README (#222)

### Enhanced
- Stricter analyze scoring with continuous penalties and weighted dead code (#212)

### Improved
- Add test coverage for version package (#218)
- Add test coverage for system-analysis helper utilities (#216)
- Remove unused clone detection code (#214)

## [1.2.2] - 2025-10-19

### Enhanced
- Compact MCP responses for MCP tools via the new `output_mode` parameter, including summary/detailed formats and correct totals when `max_results` is used (#203)

### Improved
- Drop unused `python/pyproject-mcp.toml` now that the MCP wheel uses the primary config (#204)
- Remove redundant guidance from the Claude MCP setup snippet (#205)

## [1.2.1] - 2025-10-18

### Added
- Standalone pyscn-mcp PyPI package for MCP-only usage (#194)

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
