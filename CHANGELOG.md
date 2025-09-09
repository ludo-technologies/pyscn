# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0-beta.13] - 2025-09-08

### Latest Beta Release

*Note: Previous beta versions (0.1.0-beta.1 through 0.1.0-beta.12) contained distribution issues and have been removed from both PyPI and GitHub releases.*

#### Features
- **Complexity Analysis**: CFG-based cyclomatic complexity calculation with risk thresholds
  - McCabe cyclomatic complexity using Control Flow Graph
  - Risk level classification (low/medium/high)
  - Sorting by complexity, risk, or name
  - Configurable thresholds

- **Dead Code Detection**: Unreachable code identification
  - Code after return/break/continue/raise statements
  - Unreachable branches detection
  - Severity levels (critical/warning/info)
  - Context display with surrounding lines

- **Clone Detection**: APTED algorithm for structural code similarity
  - Type 1-4 clone detection
  - Configurable similarity thresholds (0.0-1.0)
  - Clone grouping and detailed reporting
  - Multiple cost models for tree edit distance

- **CBO Metrics**: Coupling Between Objects analysis
  - Class-level coupling measurement
  - Risk level assessment
  - Include/exclude built-ins and imports options
  - Sorting by coupling or name

- **Unified Analysis**: Combined reporting across all metrics
  - HTML reports with interactive visualization
  - JSON/YAML/CSV export formats
  - Timestamped output files
  - Health score calculation

- **Configuration System**: Flexible configuration management
  - Hierarchical config discovery (project → XDG → home)
  - YAML/JSON config file support
  - Command-line flag overrides
  - Init command for starter config

#### Performance
- Fast analysis: 10,000+ lines per second
- Parallel processing for multiple files
- Efficient memory usage with streaming parsers
- Optimized tree-sitter integration

#### Known Limitations
- Python 3.10+ features not fully supported:
  - Match statements (PEP 634)
  - Walrus operator (:=) in complex contexts
  - Some async/await patterns
- System-level structural analysis planned for future release
- Module dependency graphs not yet implemented

#### Technical Details
- Built with Go 1.24+ for optimal performance
- Clean Architecture with domain/use-case separation
- Tree-sitter for robust Python parsing
- Comprehensive test suite (12 packages)
- Cross-platform support (macOS, Linux, Windows)

#### Installation
```bash
# Install latest beta version
pip install --pre pyscn

# Or specify exact version
pip install pyscn==0.1.0b13
```

#### Usage
```bash
# Quick quality check
pyscn check .

# Comprehensive analysis
pyscn analyze --html src/

# Individual analyses
pyscn complexity src/
pyscn deadcode src/
pyscn clone src/
pyscn cbo src/
```