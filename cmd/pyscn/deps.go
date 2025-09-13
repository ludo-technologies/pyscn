package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ludo-technologies/pyscn/app"
	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/service"
)

var (
	// Analysis options
	depsIncludeStdLib     bool
	depsIncludeThirdParty bool
	depsFollowRelative    bool
	depsDetectCycles      bool

	// Output format flags
	depsJSON bool
	depsCSV  bool
	depsHTML bool
	depsYAML bool
	depsDOT  bool // DOT format for graph visualization

	depsNoOpen      bool

	// File selection options
	depsRecursive       bool
	depsIncludePatterns []string
	depsExcludePatterns []string
	depsConfigPath      string

)

// depsCmd represents the deps command
var depsCmd = &cobra.Command{
	Use:   "deps [paths...]",
	Short: "Analyze module dependencies and coupling",
	Long: `Analyze module dependencies, detect circular dependencies, and calculate coupling metrics.

This command performs comprehensive dependency analysis including:
• Module dependency graph construction
• Circular dependency detection using Tarjan's algorithm
• Robert Martin's coupling metrics (Ca, Ce, I, A, D)
• Dependency chain analysis
• Architecture quality assessment

Examples:
  pyscn deps src/                  # Analyze all modules in src/
  pyscn deps --html src/           # Generate interactive HTML report

Output formats:
  --html       - Interactive HTML report with visualizations (recommended)
  --json       - JSON output for programmatic processing
  --csv        - CSV output for spreadsheet analysis
  --yaml       - YAML output
  --dot        - DOT graph for external visualization tools`,
	Args: cobra.MinimumNArgs(1),
	RunE: runDepsCommand,
}

func init() {
	rootCmd.AddCommand(depsCmd)

	// Analysis options
	depsCmd.Flags().BoolVar(&depsIncludeStdLib, "include-stdlib", false, "Include standard library dependencies")
	depsCmd.Flags().BoolVar(&depsIncludeThirdParty, "include-third-party", true, "Include third-party dependencies")
	depsCmd.Flags().BoolVar(&depsFollowRelative, "follow-relative", true, "Follow relative imports")
	depsCmd.Flags().BoolVar(&depsDetectCycles, "detect-cycles", true, "Detect circular dependencies")

	// Output options
	depsCmd.Flags().BoolVar(&depsJSON, "json", false, "Generate JSON report file")
	depsCmd.Flags().BoolVar(&depsCSV, "csv", false, "Generate CSV report file")
	depsCmd.Flags().BoolVar(&depsHTML, "html", false, "Generate HTML report file")
	depsCmd.Flags().BoolVar(&depsYAML, "yaml", false, "Generate YAML report file")
	depsCmd.Flags().BoolVar(&depsDOT, "dot", false, "Generate DOT graph file")
	depsCmd.Flags().BoolVar(&depsNoOpen, "no-open", false, "Don't auto-open HTML in browser")

	// File selection options
	depsCmd.Flags().BoolVar(&depsRecursive, "recursive", true, "Recursively analyze subdirectories")
	depsCmd.Flags().StringSliceVar(&depsIncludePatterns, "include", []string{"*.py"}, "Include file patterns")
	depsCmd.Flags().StringSliceVar(&depsExcludePatterns, "exclude", []string{}, "Exclude file patterns")

	// Configuration
	depsCmd.Flags().StringVarP(&depsConfigPath, "config", "c", "", "Configuration file path")
}

func runDepsCommand(cmd *cobra.Command, args []string) error {
	// Determine output format from flags
	outputFormat := domain.OutputFormatText // Default
	outputPath := ""
	outputWriter := os.Stdout
	extension := ""
	
	formatCount := 0
	if depsJSON {
		formatCount++
		outputFormat = domain.OutputFormatJSON
		extension = "json"
	}
	if depsCSV {
		formatCount++
		outputFormat = domain.OutputFormatCSV
		extension = "csv"
	}
	if depsHTML {
		formatCount++
		outputFormat = domain.OutputFormatHTML
		extension = "html"
	}
	if depsYAML {
		formatCount++
		outputFormat = domain.OutputFormatYAML
		extension = "yaml"
	}
	if depsDOT {
		formatCount++
		outputFormat = domain.OutputFormatDOT
		extension = "dot"
	}
	
	// Check for conflicting format flags
	if formatCount > 1 {
		return fmt.Errorf("only one output format flag can be specified")
	}

	// Generate output path for non-text formats
	if outputFormat != domain.OutputFormatText && extension != "" {
		targetPath := getTargetPathFromArgs(args)
		var err error
		outputPath, err = generateOutputFilePath("deps", extension, targetPath)
		if err != nil {
			return fmt.Errorf("failed to generate output path: %w", err)
		}
		outputWriter = nil // Don't write to stdout for file output
	}

	// Build dependency analysis request
	request := domain.SystemAnalysisRequest{
		Paths:           args,
		OutputFormat:    outputFormat,
		OutputWriter:    outputWriter,
		OutputPath:      outputPath,
		NoOpen:          depsNoOpen,
		
		// Enable only dependency analysis
		AnalyzeDependencies: true,
		AnalyzeArchitecture: false,
		AnalyzeQuality:      false,
		
		// Analysis options
		IncludeStdLib:     depsIncludeStdLib,
		IncludeThirdParty: depsIncludeThirdParty,
		FollowRelative:    depsFollowRelative,
		DetectCycles:      depsDetectCycles,
		
		// File selection
		ConfigPath:      depsConfigPath,
		Recursive:       depsRecursive,
		IncludePatterns: depsIncludePatterns,
		ExcludePatterns: depsExcludePatterns,
	}

	// Build dependencies
	systemService := service.NewSystemAnalysisService()
	fileReader := service.NewFileReader()
	formatter := service.NewSystemAnalysisFormatter()
	configLoader := service.NewSystemAnalysisConfigurationLoader()

	// Create use case
	systemUseCase, err := app.NewSystemAnalysisUseCaseBuilder().
		WithService(systemService).
		WithFileReader(fileReader).
		WithFormatter(formatter).
		WithConfigLoader(configLoader).
		Build()
	if err != nil {
		return fmt.Errorf("failed to create system analysis use case: %w", err)
	}

	// Execute analysis
	ctx := cmd.Context()
	if err := systemUseCase.Execute(ctx, request); err != nil {
		return fmt.Errorf("system analysis failed: %w", err)
	}

	return nil
}

