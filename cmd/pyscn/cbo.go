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
	cboMinCBO    int
	cboMaxCBO    int
	cboShowZeros bool
	cboSortBy    string

	cboLowThreshold    int
	cboMediumThreshold int

	cboIncludeBuiltins bool
	cboIncludeImports  bool

	// Output format flags (following pattern of other commands)
	cboJSON bool
	cboCSV  bool
	cboHTML bool
	cboYAML bool

	cboNoOpen      bool
	cboShowDetails bool

	cboRecursive       bool
	cboIncludePatterns []string
	cboExcludePatterns []string
	cboConfigPath      string
)

// cboCmd represents the cbo command
var cboCmd = &cobra.Command{
	Use:   "cbo [paths...]",
	Short: "Analyze CBO (Coupling Between Objects) metrics",
	Long: `Analyze CBO (Coupling Between Objects) metrics for Python classes.

CBO measures the number of classes to which a class is coupled. High coupling
indicates that a class depends on many other classes, making it harder to
maintain, test, and reuse.

Examples:
  pyscn cbo src/                    # Analyze all Python files in src/
  pyscn cbo --min-cbo 5 src/        # Show only classes with CBO >= 5
  pyscn cbo --sort coupling src/    # Sort by CBO count (default)
  pyscn cbo --json src/             # Output as JSON
  pyscn cbo --html src/             # Generate HTML report
  pyscn cbo --show-zeros src/       # Include classes with CBO = 0
  pyscn cbo --include-builtins src/ # Include built-in type dependencies

Sort options:
  coupling  - Sort by CBO count (default)
  name      - Sort alphabetically by class name
  risk      - Sort by risk level (high to low)
  location  - Sort by file path and line number

Risk levels are determined by thresholds:
  Low:    CBO <= 5 (default low threshold)
  Medium: 6 <= CBO <= 10 (default medium threshold)
  High:   CBO > 10`,
	Args: cobra.MinimumNArgs(1),
	RunE: runCBOCommand,
}

func init() {
	rootCmd.AddCommand(cboCmd)

	// Filtering options
	cboCmd.Flags().IntVar(&cboMinCBO, "min-cbo", 0, "Minimum CBO to report")
	cboCmd.Flags().IntVar(&cboMaxCBO, "max-cbo", 0, "Maximum CBO to report (0 = no limit)")
	cboCmd.Flags().BoolVar(&cboShowZeros, "show-zeros", false, "Include classes with CBO = 0")
	cboCmd.Flags().StringVar(&cboSortBy, "sort", "coupling", "Sort criteria (coupling|name|risk|location)")

	// Threshold configuration
	cboCmd.Flags().IntVar(&cboLowThreshold, "low-threshold", 5, "Low risk threshold")
	cboCmd.Flags().IntVar(&cboMediumThreshold, "medium-threshold", 10, "Medium risk threshold")

	// Analysis scope options
	cboCmd.Flags().BoolVar(&cboIncludeBuiltins, "include-builtins", false, "Include built-in type dependencies")
	cboCmd.Flags().BoolVar(&cboIncludeImports, "include-imports", true, "Include imported class dependencies")

	// Output options (following pattern of other commands)
	cboCmd.Flags().BoolVar(&cboJSON, "json", false, "Generate JSON report file")
	cboCmd.Flags().BoolVar(&cboCSV, "csv", false, "Generate CSV report file")
	cboCmd.Flags().BoolVar(&cboHTML, "html", false, "Generate HTML report file")
	cboCmd.Flags().BoolVar(&cboYAML, "yaml", false, "Generate YAML report file")
	cboCmd.Flags().BoolVar(&cboNoOpen, "no-open", false, "Don't auto-open HTML in browser")
	cboCmd.Flags().BoolVar(&cboShowDetails, "details", false, "Show detailed dependency information")

	// File selection options
	cboCmd.Flags().BoolVar(&cboRecursive, "recursive", true, "Recursively analyze subdirectories")
	cboCmd.Flags().StringSliceVar(&cboIncludePatterns, "include", []string{"*.py"}, "Include file patterns")
	cboCmd.Flags().StringSliceVar(&cboExcludePatterns, "exclude", []string{}, "Exclude file patterns")

	// Configuration  
	cboCmd.Flags().StringVarP(&cboConfigPath, "config", "c", "", "Configuration file path")
}

func runCBOCommand(cmd *cobra.Command, args []string) error {
	// Determine output format from flags
	outputFormat := domain.OutputFormatText // Default
	outputPath := ""
	
	if cboJSON {
		outputFormat = domain.OutputFormatJSON
	} else if cboCSV {
		outputFormat = domain.OutputFormatCSV
	} else if cboHTML {
		outputFormat = domain.OutputFormatHTML
		// Generate default filename for HTML if not specified
		outputPath = "cbo_report.html"
	} else if cboYAML {
		outputFormat = domain.OutputFormatYAML
	}

	// Build CBO request from flags and arguments
	request := domain.CBORequest{
		Paths:           args,
		OutputFormat:    outputFormat,
		OutputWriter:    os.Stdout,
		OutputPath:      outputPath,
		NoOpen:          cboNoOpen,
		ShowDetails:     cboShowDetails,
		MinCBO:          cboMinCBO,
		MaxCBO:          cboMaxCBO,
		SortBy:          domain.SortCriteria(cboSortBy),
		ShowZeros:       cboShowZeros,
		LowThreshold:    cboLowThreshold,
		MediumThreshold: cboMediumThreshold,
		ConfigPath:      cboConfigPath,
		Recursive:       cboRecursive,
		IncludePatterns: cboIncludePatterns,
		ExcludePatterns: cboExcludePatterns,
		IncludeBuiltins: cboIncludeBuiltins,
		IncludeImports:  cboIncludeImports,
	}

	// Validate sort criteria
	if err := validateCBOSortCriteria(request.SortBy); err != nil {
		return fmt.Errorf("invalid sort criteria: %w", err)
	}

	// Build dependencies
	cboService := service.NewCBOService()
	fileReader := service.NewFileReader()
	formatter := service.NewCBOFormatter()

	// Create use case
	cboUseCase, err := app.NewCBOUseCaseBuilder().
		WithService(cboService).
		WithFileReader(fileReader).
		WithFormatter(formatter).
		Build()
	if err != nil {
		return fmt.Errorf("failed to create CBO use case: %w", err)
	}

	// Execute analysis
	ctx := cmd.Context()
	if err := cboUseCase.Execute(ctx, request); err != nil {
		return fmt.Errorf("CBO analysis failed: %w", err)
	}

	return nil
}

func validateCBOSortCriteria(sortBy domain.SortCriteria) error {
	switch sortBy {
	case domain.SortByCoupling, domain.SortByName, domain.SortByRisk, domain.SortByLocation:
		return nil
	default:
		return fmt.Errorf("unsupported sort criteria '%s'. Valid options: coupling, name, risk, location", sortBy)
	}
}