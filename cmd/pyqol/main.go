package main

import (
	"os"

	"github.com/pyqol/pyqol/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pyqol",
	Short: "Python Quality of Life - Advanced Python static analysis",
	Long: `pyqol is a next-generation Python static analysis tool that uses 
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

	// Add subcommands
	rootCmd.AddCommand(NewAnalyzeCmd())
	rootCmd.AddCommand(NewCheckCmd())
	rootCmd.AddCommand(NewVersionCmd())
	rootCmd.AddCommand(NewInitCmd())
	rootCmd.AddCommand(complexityCmd)
	rootCmd.AddCommand(NewDeadCodeCmd())
	addCloneCommand(rootCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
