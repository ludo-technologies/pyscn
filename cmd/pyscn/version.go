package main

import (
	"fmt"

	"github.com/ludo-technologies/pyscn/internal/version"
	"github.com/spf13/cobra"
)

// VersionCommand represents the version command
type VersionCommand struct {
	short bool
}

// NewVersionCommand creates a new version command
func NewVersionCommand() *VersionCommand {
	return &VersionCommand{
		short: false,
	}
}

// CreateCobraCommand creates the cobra command for version display
func (v *VersionCommand) CreateCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long: `Display detailed version information for pyscn.

Shows version number, build commit, build date, Go version, and platform information.
Use --short to display only the version number.

Examples:
  # Show full version information
  pyscn version

  # Show only version number (useful for scripts)
  pyscn version --short`,
		RunE: v.runVersion,
	}

	// Add flags
	cmd.Flags().BoolVarP(&v.short, "short", "s", false, "Show only version number")

	return cmd
}

// runVersion executes the version command
func (v *VersionCommand) runVersion(cmd *cobra.Command, args []string) error {
	if v.short {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", version.Short())
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", version.Info())
	}
	return nil
}

// NewVersionCmd creates and returns the version cobra command
func NewVersionCmd() *cobra.Command {
	versionCommand := NewVersionCommand()
	return versionCommand.CreateCobraCommand()
}
