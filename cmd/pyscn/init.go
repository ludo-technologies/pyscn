package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ludo-technologies/pyscn/internal/config"
	"github.com/spf13/cobra"
)

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

	// Use embedded default config
	configData := config.DefaultConfigTOML

	// Write the configuration file
	if err := os.WriteFile(configPath, []byte(configData), 0644); err != nil {
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
