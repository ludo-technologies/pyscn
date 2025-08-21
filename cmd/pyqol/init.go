package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const defaultConfigTemplate = `# pyqol configuration file
# This file configures all analysis features of pyqol
# Place this file in your project root to customize analysis behavior

# =============================================================================
# COMPLEXITY ANALYSIS
# =============================================================================
complexity:
  enabled: true                    # Enable complexity analysis
  low_threshold: 9                 # Functions with complexity ≤ 9 are low risk
  medium_threshold: 19             # Functions with complexity 10-19 are medium risk
                                   # Functions with complexity ≥ 20 are high risk
  max_complexity: 0                # Maximum allowed complexity (0 = no limit)
  report_unchanged: true           # Report functions with complexity = 1

# =============================================================================
# DEAD CODE DETECTION
# =============================================================================
dead_code:
  enabled: true                    # Enable dead code detection
  min_severity: "warning"          # Minimum severity to report: critical, warning, info
  show_context: false              # Show surrounding code context
  context_lines: 3                 # Number of context lines to show
  sort_by: "severity"              # Sort by: severity, line, file, function
  
  # Detection options - configure what types of dead code to detect
  detect_after_return: true        # Code after return statements
  detect_after_break: true         # Code after break statements
  detect_after_continue: true      # Code after continue statements
  detect_after_raise: true         # Code after raise statements
  detect_unreachable_branches: true # Unreachable conditional branches
  
  # Patterns to ignore (regex patterns)
  ignore_patterns: []

# =============================================================================
# CLONE DETECTION
# =============================================================================
clones:
  analysis:
    min_lines: 5                   # Minimum lines for clone candidates
    min_nodes: 10                  # Minimum AST nodes for clone candidates
    max_edit_distance: 50.0        # Maximum edit distance allowed
    ignore_literals: false         # Ignore differences in literal values
    ignore_identifiers: false      # Ignore differences in identifier names
    cost_model_type: "python"      # Cost model: default, python, weighted
  
  thresholds:
    # Clone type similarity thresholds (0.0 - 1.0)
    type1_threshold: 0.95          # Type-1: Identical code (except whitespace/comments)
    type2_threshold: 0.85          # Type-2: Syntactically identical (different identifiers)
    type3_threshold: 0.75          # Type-3: Syntactically similar (small modifications)
    type4_threshold: 0.65          # Type-4: Functionally similar (different syntax)
    similarity_threshold: 0.8      # General minimum similarity threshold
  
  filtering:
    min_similarity: 0.0            # Minimum similarity to report
    max_similarity: 1.0            # Maximum similarity to report
    enabled_clone_types:           # Clone types to detect
      - "type1"
      - "type2"
      - "type3"
      - "type4"
    max_results: 0                 # Maximum results (0 = no limit)
  
  input:
    paths: []                      # Paths to analyze (empty = use command line)
    recursive: true                # Recursively analyze directories
    include_patterns:              # File patterns to include
      - "*.py"
    exclude_patterns:              # File patterns to exclude
      - "test_*.py"
      - "*_test.py"
  
  output:
    format: "text"                 # Output format: text, json, yaml, csv
    show_details: false            # Show detailed clone information
    show_content: false            # Include source code content in output
    sort_by: "similarity"          # Sort by: similarity, size, location, type
    group_clones: true             # Group related clones together
  
  performance:
    max_parallel_jobs: 0           # Max parallel jobs (0 = auto-detect)
    cache_enabled: true            # Enable result caching
    memory_limit_mb: 0             # Memory limit in MB (0 = no limit)

# =============================================================================
# OUTPUT CONFIGURATION
# =============================================================================
output:
  format: "text"                   # Default output format: text, json, yaml, csv
  show_details: false              # Show detailed breakdown by default
  sort_by: "name"                  # Default sort: name, complexity, risk
  min_complexity: 1                # Minimum complexity to report

# =============================================================================
# ANALYSIS CONFIGURATION
# =============================================================================
analysis:
  recursive: true                  # Recursively analyze directories
  follow_symlinks: false           # Follow symbolic links
  include_patterns:                # File patterns to include
    - "*.py"
  exclude_patterns:                # File patterns to exclude
    - "test_*.py"
    - "*_test.py"
    - "__pycache__/**"
    - "*.pyc"
    - ".pytest_cache/**"
    - ".tox/**"
    - "venv/**"
    - "env/**"
    - ".venv/**"
    - ".env/**"

# =============================================================================
# EXAMPLE CONFIGURATIONS
# =============================================================================

# Uncomment and modify these sections for common use cases:

# # Strict mode - fail on any issues
# complexity:
#   max_complexity: 10
# dead_code:
#   min_severity: "critical"
# 
# # Relaxed mode - only catch major issues  
# complexity:
#   low_threshold: 15
#   medium_threshold: 25
# dead_code:
#   min_severity: "warning"
# 
# # Clone detection focused on exact matches
# clones:
#   thresholds:
#     similarity_threshold: 0.95
#   filtering:
#     enabled_clone_types: ["type1", "type2"]
# 
# # Performance optimized for large codebases
# clones:
#   performance:
#     max_parallel_jobs: 4
#     memory_limit_mb: 1024
#   analysis:
#     min_lines: 10
#     min_nodes: 20
`

// InitCommand represents the init command
type InitCommand struct {
	force      bool
	configPath string
}

// NewInitCommand creates a new init command
func NewInitCommand() *InitCommand {
	return &InitCommand{
		force:      false,
		configPath: ".pyqol.yaml",
	}
}

// CreateCobraCommand creates the cobra command for configuration initialization
func (i *InitCommand) CreateCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize pyqol configuration file",
		Long: `Initialize a pyqol configuration file in the current directory.

Creates a .pyqol.yaml file with comprehensive configuration options and
helpful comments explaining each setting. This file allows you to customize
pyqol's behavior for your project.

The generated configuration includes settings for:
• Complexity analysis thresholds and options
• Dead code detection parameters  
• Clone detection configuration
• File inclusion/exclusion patterns
• Output formatting preferences

Examples:
  # Create .pyqol.yaml in current directory
  pyqol init

  # Create config file with custom name
  pyqol init --config myconfig.yaml

  # Overwrite existing configuration file
  pyqol init --force`,
		RunE: i.runInit,
	}

	// Add flags
	cmd.Flags().BoolVarP(&i.force, "force", "f", false, "Overwrite existing configuration file")
	cmd.Flags().StringVarP(&i.configPath, "config", "c", ".pyqol.yaml", "Configuration file path")

	return cmd
}

// runInit executes the init command
func (i *InitCommand) runInit(cmd *cobra.Command, args []string) error {
	// Resolve the absolute path
	configPath, err := filepath.Abs(i.configPath)
	if err != nil {
		return fmt.Errorf("failed to resolve config path: %w", err)
	}

	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil && !i.force {
		return fmt.Errorf("configuration file already exists: %s\nUse --force to overwrite", configPath)
	}

	// Create directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", configDir, err)
	}

	// Write the configuration file
	if err := os.WriteFile(configPath, []byte(defaultConfigTemplate), 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	// Print success message
	relPath, err := filepath.Rel(".", configPath)
	if err != nil {
		relPath = configPath // Fall back to absolute path if relative fails
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✅ Configuration file created: %s\n", relPath)
	fmt.Fprintf(cmd.OutOrStdout(), "\nTo customize pyqol for your project:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  1. Edit %s\n", relPath)
	fmt.Fprintf(cmd.OutOrStdout(), "  2. Uncomment and modify settings as needed\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  3. Run 'pyqol analyze .' to use your configuration\n")

	return nil
}

// NewInitCmd creates and returns the init cobra command
func NewInitCmd() *cobra.Command {
	initCommand := NewInitCommand()
	return initCommand.CreateCobraCommand()
}
