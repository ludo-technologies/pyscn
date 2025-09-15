package main

import (
	"os"

	"github.com/ludo-technologies/pyscn/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pyscn",
	Short: "An Intelligent Python Code Quality Analyzer",
	Long: `pyscn is an intelligent Python code quality analyzer that uses 
Control Flow Graph (CFG) and APTED (tree edit distance) algorithms 
to provide deep code quality insights beyond traditional linters.

Features:
  • CFG-based dead code detection
  • Cyclomatic complexity analysis  
  • Clone detection with APTED algorithm
  • High-performance analysis (>10,000 lines/second)`,
	Version: version.Short(),
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")

	// Add main subcommands
	rootCmd.AddCommand(NewAnalyzeCmd())
	rootCmd.AddCommand(NewCheckCmd())
	rootCmd.AddCommand(NewVersionCmd())
	rootCmd.AddCommand(NewInitCmd())

	// Add deprecated commands (hidden from help)
	complexityCmd := NewComplexityCmd()
	complexityCmd.Hidden = true
	rootCmd.AddCommand(complexityCmd)

	deadCodeCmd := NewDeadCodeCmd()
	deadCodeCmd.Hidden = true
	rootCmd.AddCommand(deadCodeCmd)

	cboCmd := NewCBOCmd()
	cboCmd.Hidden = true
	rootCmd.AddCommand(cboCmd)

	depsCmd := NewDepsCmd()
	depsCmd.Hidden = true
	rootCmd.AddCommand(depsCmd)

	// Add clone command (uses different pattern)
	addCloneCommand(rootCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
