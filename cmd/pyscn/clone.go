package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ludo-technologies/pyscn/app"
	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/constants"
	"github.com/ludo-technologies/pyscn/service"
)

// CloneCommand handles the clone detection CLI command
type CloneCommand struct {
	// Input parameters
	recursive       bool
	configFile      string
	includePatterns []string
	excludePatterns []string

	// Analysis configuration
	minLines            int
	minNodes            int
	similarityThreshold float64
	maxEditDistance     float64
	ignoreLiterals      bool
	ignoreIdentifiers   bool

	// Type-specific thresholds
	type1Threshold float64
	type2Threshold float64
	type3Threshold float64
	type4Threshold float64

	// Output format flags (only one should be true)
	html      bool
	json      bool
	csv       bool
	yaml      bool
	noOpen    bool
	
    // Output options
    showDetails  bool
    showContent  bool
    sortBy       string
    groupClones  bool

    // Grouping options
    groupMode      string  // "connected", "star", "complete_linkage", "k_core"
    groupThreshold float64 // グループ内最小類似度
    kCoreK         int     // k-coreのk値

	// Filtering
	minSimilarity float64
	maxSimilarity float64
	cloneTypes    []string

	// Advanced options
	costModelType string
	verbose       bool

	// Performance options
	timeout time.Duration

	// LSH options
	useLSH                 bool
	lshThreshold           float64
	lshBands               int
	lshRows                int
	lshHashes              int
	lshAutoThreshold       bool
}

// NewCloneCommand creates a new clone detection command
func NewCloneCommand() *CloneCommand {
    return &CloneCommand{
		recursive:           true,
		minLines:            5,
		minNodes:            5,
		similarityThreshold: 0.8,
		maxEditDistance:     50.0,
		ignoreLiterals:      false,
		ignoreIdentifiers:   false,
		type1Threshold:      constants.DefaultType1CloneThreshold,
		type2Threshold:      constants.DefaultType2CloneThreshold,
		type3Threshold:      constants.DefaultType3CloneThreshold,
		type4Threshold:      constants.DefaultType4CloneThreshold,
		html:                false,
		json:                false,
		csv:                 false,
		yaml:                false,
		noOpen:              false,
		showDetails:         false,
		showContent:         false,
        sortBy:              "similarity",
        groupClones:         true,
        groupMode:           "connected",
        groupThreshold:      constants.DefaultType3CloneThreshold,
        kCoreK:              2,
        minSimilarity:       0.0,
        maxSimilarity:       1.0,
        cloneTypes:          []string{"type1", "type2", "type3", "type4"},
        costModelType:       "python",
        verbose:             false,
		timeout:             5 * time.Minute,

		// LSH defaults
		useLSH:               false, // Disabled by default
		lshThreshold:         0.78,  // Default threshold for 32 bands, 4 rows
		lshBands:             32,    // Default number of bands
		lshRows:              4,     // Default rows per band
		lshHashes:            128,   // Default number of MinHash functions
		lshAutoThreshold:     true,  // Automatically determine threshold
	}
}

// CreateCobraCommand creates the Cobra command for clone detection
func (c *CloneCommand) CreateCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone [files...]",
		Short: "Detect code clones using tree edit distance",
		Long: `Detect code clones in Python files using the APTED algorithm.

This command identifies structurally similar code fragments that may be candidates
for refactoring. It supports detection of different clone types:

- Type-1: Identical code (except whitespace and comments)
- Type-2: Syntactically identical but with different identifiers/literals
- Type-3: Syntactically similar with small modifications
- Type-4: Functionally similar but syntactically different

Examples:
  # Detect clones in current directory
  pyscn clone .

  # Detect clones with custom similarity threshold
  pyscn clone --similarity-threshold 0.9 src/

  # Show detailed clone information with content
  pyscn clone --details --show-content src/

  # Only detect Type-1 and Type-2 clones
  pyscn clone --clone-types type1,type2 src/

  # Output results as JSON
  pyscn clone --format json src/ > clones.json`,
		RunE: c.runCloneDetection,
	}

	// Input flags
	cmd.Flags().BoolVarP(&c.recursive, "recursive", "r", c.recursive,
		"Recursively analyze directories")
	cmd.Flags().StringVarP(&c.configFile, "config", "c", c.configFile,
		"Path to configuration file")
	cmd.Flags().StringSliceVar(&c.includePatterns, "include", []string{"*.py"},
		"File patterns to include")
	cmd.Flags().StringSliceVar(&c.excludePatterns, "exclude", []string{"test_*.py", "*_test.py"},
		"File patterns to exclude")

	// Analysis configuration flags
	cmd.Flags().IntVar(&c.minLines, "min-lines", c.minLines,
		"Minimum number of lines for clone candidates")
	cmd.Flags().IntVar(&c.minNodes, "min-nodes", c.minNodes,
		"Minimum number of AST nodes for clone candidates")
	cmd.Flags().Float64VarP(&c.similarityThreshold, "similarity-threshold", "s", c.similarityThreshold,
		"Minimum similarity threshold for clone detection (0.0-1.0)")
	cmd.Flags().Float64Var(&c.maxEditDistance, "max-distance", c.maxEditDistance,
		"Maximum edit distance allowed for clones")
	cmd.Flags().BoolVar(&c.ignoreLiterals, "ignore-literals", c.ignoreLiterals,
		"Ignore differences in literal values")
	cmd.Flags().BoolVar(&c.ignoreIdentifiers, "ignore-identifiers", c.ignoreIdentifiers,
		"Ignore differences in identifier names")

	// Type-specific threshold flags
	cmd.Flags().Float64Var(&c.type1Threshold, "type1-threshold", c.type1Threshold,
		"Similarity threshold for Type-1 clones (identical)")
	cmd.Flags().Float64Var(&c.type2Threshold, "type2-threshold", c.type2Threshold,
		"Similarity threshold for Type-2 clones (renamed)")
	cmd.Flags().Float64Var(&c.type3Threshold, "type3-threshold", c.type3Threshold,
		"Similarity threshold for Type-3 clones (near-miss)")
	cmd.Flags().Float64Var(&c.type4Threshold, "type4-threshold", c.type4Threshold,
		"Similarity threshold for Type-4 clones (semantic)")

	// Output format flags
	cmd.Flags().BoolVar(&c.html, "html", false, "Generate HTML report file")
	cmd.Flags().BoolVar(&c.json, "json", false, "Generate JSON report file")
	cmd.Flags().BoolVar(&c.csv, "csv", false, "Generate CSV report file")
	cmd.Flags().BoolVar(&c.yaml, "yaml", false, "Generate YAML report file")
	cmd.Flags().BoolVar(&c.noOpen, "no-open", false, "Don't auto-open HTML in browser")
	
	// Output options
	cmd.Flags().BoolVarP(&c.showDetails, "details", "d", c.showDetails,
		"Show detailed clone information")
	cmd.Flags().BoolVar(&c.showContent, "show-content", c.showContent,
		"Include source code content in output")
	cmd.Flags().StringVar(&c.sortBy, "sort", c.sortBy,
		"Sort results by: similarity, size, location, type")
    cmd.Flags().BoolVar(&c.groupClones, "group", c.groupClones,
        "Group related clones together")

    // Grouping strategy flags
    cmd.Flags().StringVar(&c.groupMode, "group-mode", c.groupMode,
        "Grouping strategy: connected, star, complete_linkage, k_core")
    cmd.Flags().Float64Var(&c.groupThreshold, "group-threshold", c.groupThreshold,
        "Minimum similarity for group membership")
    cmd.Flags().IntVar(&c.kCoreK, "k-core-k", c.kCoreK,
        "Minimum neighbors for k-core mode")

	// Filtering flags
	cmd.Flags().Float64Var(&c.minSimilarity, "min-similarity", c.minSimilarity,
		"Minimum similarity to report (0.0-1.0)")
	cmd.Flags().Float64Var(&c.maxSimilarity, "max-similarity", c.maxSimilarity,
		"Maximum similarity to report (0.0-1.0)")
	cmd.Flags().StringSliceVar(&c.cloneTypes, "clone-types", c.cloneTypes,
		"Clone types to detect: type1, type2, type3, type4")

	// Advanced flags
	cmd.Flags().StringVar(&c.costModelType, "cost-model", c.costModelType,
		"Cost model to use: default, python, weighted")
	cmd.Flags().BoolVarP(&c.verbose, "verbose", "v", c.verbose,
		"Enable verbose output")

	// Performance flags
	cmd.Flags().DurationVar(&c.timeout, "clone-timeout", c.timeout,
		"Maximum time for clone analysis (e.g., 5m, 30s)")

	// LSH flags
	cmd.Flags().BoolVar(&c.useLSH, "use-lsh", c.useLSH,
		"Enable LSH (Locality Sensitive Hashing) acceleration")
	cmd.Flags().Float64Var(&c.lshThreshold, "lsh-threshold", c.lshThreshold,
		"LSH similarity threshold for candidate pairs (0.0-1.0)")
	cmd.Flags().IntVar(&c.lshBands, "lsh-bands", c.lshBands,
		"Number of LSH bands (more bands = higher precision)")
	cmd.Flags().IntVar(&c.lshRows, "lsh-rows", c.lshRows,
		"Rows per LSH band (more rows = higher recall)")
	cmd.Flags().IntVar(&c.lshHashes, "lsh-hashes", c.lshHashes,
		"Number of MinHash functions to use")
	cmd.Flags().BoolVar(&c.lshAutoThreshold, "lsh-auto-threshold", c.lshAutoThreshold,
		"Automatically determine LSH threshold from bands/rows")

	return cmd
}

// runCloneDetection executes the clone detection command
func (c *CloneCommand) runCloneDetection(cmd *cobra.Command, args []string) error {
	// Set default paths if none provided
	if len(args) == 0 {
		args = []string{"."}
	}

	// Create clone request from command flags
	request, err := c.createCloneRequest(cmd, args)
	if err != nil {
		return fmt.Errorf("failed to create clone request: %w", err)
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	// Create clone use case with dependencies
	useCase, err := c.createCloneUseCase(cmd)
	if err != nil {
		return fmt.Errorf("failed to create clone use case: %w", err)
	}

	// Execute clone detection
	ctx := context.Background()
	err = useCase.Execute(ctx, *request)
	if err != nil {
		return fmt.Errorf("clone detection failed: %w", err)
	}

	return nil
}

// determineOutputFormat determines the output format based on flags
func (c *CloneCommand) determineOutputFormat() (domain.OutputFormat, string, error) {
    resolver := service.NewOutputFormatResolver()
    return resolver.Determine(c.html, c.json, c.csv, c.yaml)
}

// createCloneRequest creates a clone request from command line flags
func (c *CloneCommand) createCloneRequest(cmd *cobra.Command, paths []string) (*domain.CloneRequest, error) {
	// Determine output format from flags
	outputFormat, extension, err := c.determineOutputFormat()
	if err != nil {
		return nil, err
	}

	// Parse sort criteria
	sortBy, err := c.parseSortCriteria(c.sortBy)
	if err != nil {
		return nil, err
	}

	// Parse clone types
	cloneTypes, err := c.parseCloneTypes()
	if err != nil {
		return nil, err
	}

	// Determine output destination
	var outputWriter io.Writer
	var outputPath string
	
	if outputFormat == domain.OutputFormatText {
		// Text format goes to stdout
		outputWriter = os.Stdout
	} else {
		// Other formats generate a file
		// Use first path as target for config discovery
		targetPath := getTargetPathFromArgs(paths)
		var err error
		outputPath, err = generateOutputFilePath("clone", extension, targetPath)
		if err != nil {
			return nil, fmt.Errorf("failed to generate output path: %w", err)
		}
	}

	request := &domain.CloneRequest{
		Paths:               paths,
		Recursive:           c.recursive,
		IncludePatterns:     c.includePatterns,
		ExcludePatterns:     c.excludePatterns,
		MinLines:            c.minLines,
		MinNodes:            c.minNodes,
		SimilarityThreshold: c.similarityThreshold,
		MaxEditDistance:     c.maxEditDistance,
		IgnoreLiterals:      c.ignoreLiterals,
		IgnoreIdentifiers:   c.ignoreIdentifiers,
		Type1Threshold:      c.type1Threshold,
		Type2Threshold:      c.type2Threshold,
		Type3Threshold:      c.type3Threshold,
		Type4Threshold:      c.type4Threshold,
		OutputFormat:        outputFormat,
		OutputWriter:        outputWriter,
		OutputPath:          outputPath,
		NoOpen:              c.noOpen,
		ShowDetails:         c.showDetails,
		ShowContent:         c.showContent,
		SortBy:              sortBy,
        GroupClones:         c.groupClones,
        GroupMode:           c.groupMode,
        GroupThreshold:      c.groupThreshold,
        KCoreK:              c.kCoreK,
        MinSimilarity:       c.minSimilarity,
        MaxSimilarity:       c.maxSimilarity,
        CloneTypes:          cloneTypes,
        ConfigPath:          c.configFile,
        Timeout:             c.timeout,

        // LSH configuration
        UseLSH:                 c.useLSH,
        LSHSimilarityThreshold: c.lshThreshold,
        LSHBands:               c.lshBands,
        LSHRows:                c.lshRows,
        LSHMinHashCount:        c.lshHashes,
        LSHAutoThreshold:       c.lshAutoThreshold,
	}

	return request, nil
}

// parseCloneTypes parses clone types from string slice
func (c *CloneCommand) parseCloneTypes() ([]domain.CloneType, error) {
	var cloneTypes []domain.CloneType

	for _, typeStr := range c.cloneTypes {
		switch strings.ToLower(typeStr) {
		case "type1":
			cloneTypes = append(cloneTypes, domain.Type1Clone)
		case "type2":
			cloneTypes = append(cloneTypes, domain.Type2Clone)
		case "type3":
			cloneTypes = append(cloneTypes, domain.Type3Clone)
		case "type4":
			cloneTypes = append(cloneTypes, domain.Type4Clone)
		default:
			return nil, fmt.Errorf("invalid clone type '%s', must be one of: type1, type2, type3, type4", typeStr)
		}
	}

	if len(cloneTypes) == 0 {
		// Default to all types
		cloneTypes = []domain.CloneType{domain.Type1Clone, domain.Type2Clone, domain.Type3Clone, domain.Type4Clone}
	}

	return cloneTypes, nil
}

// createCloneUseCase creates a clone use case with all dependencies
func (c *CloneCommand) createCloneUseCase(cmd *cobra.Command) (*app.CloneUseCase, error) {
	// Track which flags were explicitly set by the user
	explicitFlags := GetExplicitFlags(cmd)
	
	// Create services
	fileReader := service.NewFileReader()
	formatter := service.NewCloneOutputFormatter()
	configLoader := service.NewCloneConfigurationLoaderWithFlags(explicitFlags)
	
	cloneService := service.NewCloneService()

    // Build use case with dependencies
    return app.NewCloneUseCaseBuilder().
        WithService(cloneService).
        WithFileReader(fileReader).
        WithFormatter(formatter).
        WithConfigLoader(configLoader).
        WithOutputWriter(service.NewFileOutputWriter(cmd.ErrOrStderr())).
        Build()
}


// parseSortCriteria parses and validates the sort criteria
func (c *CloneCommand) parseSortCriteria(sort string) (domain.SortCriteria, error) {
	switch strings.ToLower(sort) {
	case "similarity":
		return domain.SortBySimilarity, nil
	case "size":
		return domain.SortBySize, nil
	case "location":
		return domain.SortByLocation, nil
	case "type":
		return domain.SortByComplexity, nil // Reuse existing constant
	default:
		return "", fmt.Errorf("unsupported sort criteria: %s (supported: similarity, size, location, type)", sort)
	}
}

// Helper function to add the clone command to the root command
func addCloneCommand(rootCmd *cobra.Command) {
	cloneCmd := NewCloneCommand()
	rootCmd.AddCommand(cloneCmd.CreateCobraCommand())
}
