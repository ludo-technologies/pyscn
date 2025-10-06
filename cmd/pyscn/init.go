package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const defaultConfigTemplate = `# pyscn configuration file (.pyscn.toml)
# This file configures all analysis features of pyscn
# Place this file in your project root to customize analysis behavior

# =============================================================================
# COMPLEXITY ANALYSIS
# =============================================================================
[complexity]
enabled = true                    # Enable complexity analysis
low_threshold = 9                 # Functions with complexity ≤ 9 are low risk
medium_threshold = 19             # Functions with complexity 10-19 are medium risk
                                  # Functions with complexity ≥ 20 are high risk
max_complexity = 0                # Maximum allowed complexity (0 = no limit)
report_unchanged = true           # Report functions with complexity = 1

# =============================================================================
# DEAD CODE DETECTION
# =============================================================================
[dead_code]
enabled = true                    # Enable dead code detection
min_severity = "warning"          # Minimum severity to report: critical, warning, info
show_context = false              # Show surrounding code context
context_lines = 3                 # Number of context lines to show
sort_by = "severity"              # Sort by: severity, line, file, function

# Detection options - configure what types of dead code to detect
detect_after_return = true        # Code after return statements
detect_after_break = true         # Code after break statements
detect_after_continue = true      # Code after continue statements
detect_after_raise = true         # Code after raise statements
detect_unreachable_branches = true # Unreachable conditional branches

# Patterns to ignore (regex patterns)
ignore_patterns = []

# =============================================================================
# CLONE DETECTION (Nested structure for better organization)
# =============================================================================

# Clone Analysis Configuration
[clones.analysis]
min_lines = 5                     # Minimum lines for clone candidates
min_nodes = 10                    # Minimum AST nodes for clone candidates
max_edit_distance = 50.0          # Maximum edit distance allowed
ignore_literals = false           # Ignore differences in literal values
ignore_identifiers = false        # Ignore differences in identifier names
cost_model_type = "python"        # Cost model: default, python, weighted

# Threshold settings for clone type classification (0.0 - 1.0)
[clones.thresholds]
type1_threshold = 0.95            # Type-1: Identical code (except whitespace/comments)
type2_threshold = 0.85            # Type-2: Syntactically identical (different identifiers)
type3_threshold = 0.80            # Type-3: Syntactically similar (small modifications)
type4_threshold = 0.75            # Type-4: Functionally similar (different syntax)
similarity_threshold = 0.8        # General minimum similarity threshold

# Filtering settings
[clones.filtering]
min_similarity = 0.0              # Minimum similarity to report
max_similarity = 1.0              # Maximum similarity to report
enabled_clone_types = ["type1", "type2", "type3", "type4"] # Clone types to detect
max_results = 0                   # Maximum results (0 = no limit)

# Grouping settings
[clones.grouping]
# Grouping strategy:
#   - connected: Group by transitive similarity (simple, fast, default)
#   - star: Star-based grouping around centroids
#   - complete_linkage: Hierarchical clustering (high quality)
#   - k_core: K-core decomposition (balanced quality/performance)
mode = "connected"
threshold = 0.85                  # Minimum similarity for group membership
k_core_k = 2                      # K value for k-core mode (minimum connections per node)

# LSH acceleration settings
[clones.lsh]
enabled = "auto"                  # LSH acceleration: true, false, auto (based on project size)
auto_threshold = 500              # Enable LSH for 500+ fragments
similarity_threshold = 0.50       # LSH similarity threshold
bands = 32                        # Number of LSH bands
rows = 4                          # Rows per LSH band
hashes = 128                      # MinHash function count

# Performance settings
[clones.performance]
max_memory_mb = 100               # Memory limit in MB
batch_size = 100                  # Batch size for processing
enable_batching = true            # Enable batching
max_goroutines = 4                # Maximum concurrent goroutines
timeout_seconds = 300             # Timeout in seconds (5 minutes)

# Input settings
[clones.input]
paths = ["."]                     # Paths to analyze (default: current directory)
recursive = true                  # Recursively analyze directories
include_patterns = ["*.py"]       # File patterns to include
exclude_patterns = ["test_*.py", "*_test.py"] # File patterns to exclude

# Output settings
[clones.output]
format = "text"                   # Output format: text, json, yaml, csv, html
show_details = false              # Show detailed clone information
show_content = false              # Include source code content in output
sort_by = "similarity"            # Sort by: similarity, size, location, type
group_clones = true               # Group related clones together

# =============================================================================
# EXAMPLE CONFIGURATIONS
# =============================================================================

# Uncomment and modify these settings for common use cases:

# # Strict mode - high precision (add to [clones.thresholds] and [clones.filtering])
# # [clones.thresholds]
# # similarity_threshold = 0.95
# # [clones.filtering]
# # enabled_clone_types = ["type1", "type2"]
#
# # Relaxed mode - catch more potential clones (add to [clones.thresholds] and [clones.analysis])
# # [clones.thresholds]
# # similarity_threshold = 0.7
# # [clones.analysis]
# # min_lines = 3
#
# # Performance optimized for large codebases (add to [clones.lsh] and [clones.performance])
# # [clones.lsh]
# # enabled = "true"
# # [clones.performance]
# # max_goroutines = 8
# # batch_size = 200
`

// InitCommand represents the init command
type InitCommand struct {
	force      bool
	configPath string
	// format removed - TOML only now
}

// NewInitCommand creates a new init command
func NewInitCommand() *InitCommand {
	return &InitCommand{
		force:      false,
		configPath: ".pyscn.toml", // TOML only
	}
}

// CreateCobraCommand creates the cobra command for configuration initialization
func (i *InitCommand) CreateCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize pyscn configuration file",
		Long: `Initialize a pyscn configuration file in the current directory.

Creates a .pyscn.toml file with comprehensive configuration options and
helpful comments explaining each setting. This file allows you to customize
pyscn's behavior for your project.

The generated configuration includes settings for:
• Complexity analysis thresholds and options
• Dead code detection parameters  
• Clone detection configuration
• File inclusion/exclusion patterns
• Output formatting preferences

Examples:
  # Create .pyscn.toml in current directory (recommended)
  pyscn init

  # Create config file with custom name  
  pyscn init --config myconfig.toml

  # Overwrite existing configuration file
  pyscn init --force`,
		RunE: i.runInit,
	}

	// Add flags
	cmd.Flags().BoolVarP(&i.force, "force", "f", false, "Overwrite existing configuration file")
	cmd.Flags().StringVarP(&i.configPath, "config", "c", ".pyscn.toml", "Configuration file path")

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
	fmt.Fprintf(cmd.OutOrStdout(), "\nTo customize pyscn for your project:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  1. Edit %s\n", relPath)
	fmt.Fprintf(cmd.OutOrStdout(), "  2. Uncomment and modify settings as needed\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  3. Run 'pyscn analyze .' to use your configuration\n")

	return nil
}

// NewInitCmd creates and returns the init cobra command
func NewInitCmd() *cobra.Command {
	initCommand := NewInitCommand()
	return initCommand.CreateCobraCommand()
}
