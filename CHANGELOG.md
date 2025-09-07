# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0b1] - 2025-09-07

### Initial Beta Release

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
# Install beta version
pip install --pre pyqol

# Or specify exact version
pip install pyqol==0.1.0b1
```

#### Usage
```bash
# Quick quality check
pyqol check .

# Comprehensive analysis
pyqol analyze --html src/

# Individual analyses
pyqol complexity src/
pyqol deadcode src/
pyqol clone src/
pyqol cbo src/
```

[0.1.0b1]: https://github.com/pyqol/pyqol/releases/tag/v0.1.0b1