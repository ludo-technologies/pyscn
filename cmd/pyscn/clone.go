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
	"github.com/ludo-technologies/pyscn/internal/config"
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
	html   bool
	json   bool
	csv    bool
	yaml   bool
	noOpen bool

	// Output options
	showDetails bool
	showContent bool
	sortBy      string
	groupClones bool

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
	useLSH       bool
	lshThreshold float64
	lshBands     int
	lshRows      int
	lshHashes    int

	// Simplified preset options
	fast    bool   // Large project preset
	precise bool   // Small project preset
	preset  string // Named preset
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
		useLSH:       false,
		lshThreshold: 0.78,
		lshBands:     32,
		lshRows:      4,
		lshHashes:    128,
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
  pyscn clone --json src/ > clones.json`,
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

	// Simplified preset flags
	cmd.Flags().BoolVar(&c.fast, "fast", false, "Fast mode for large projects (enables LSH)")
	cmd.Flags().BoolVar(&c.precise, "precise", false, "Precise mode for small projects (star grouping)")
	cmd.Flags().StringVar(&c.preset, "preset", "", "Use preset configuration: fast, precise, balanced")

	// Advanced grouping flags (hidden from main help)
	cmd.Flags().StringVar(&c.groupMode, "group-mode", c.groupMode,
		"Grouping strategy: connected, star, complete_linkage, k_core")
	cmd.Flags().Float64Var(&c.groupThreshold, "group-threshold", c.groupThreshold,
		"Minimum similarity for group membership")
	cmd.Flags().IntVar(&c.kCoreK, "k-core-k", c.kCoreK,
		"Minimum neighbors for k-core mode")

	// Mark advanced flags as hidden
	_ = cmd.Flags().MarkHidden("group-threshold")
	_ = cmd.Flags().MarkHidden("k-core-k")

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

	// LSH acceleration flags (hidden advanced options)
	cmd.Flags().BoolVar(&c.useLSH, "use-lsh", c.useLSH, "Enable LSH acceleration")
	cmd.Flags().Float64Var(&c.lshThreshold, "lsh-threshold", c.lshThreshold, "LSH MinHash similarity threshold (0.0-1.0)")
	cmd.Flags().IntVar(&c.lshBands, "lsh-bands", c.lshBands, "Number of LSH bands")
	cmd.Flags().IntVar(&c.lshRows, "lsh-rows", c.lshRows, "Rows per LSH band")
	cmd.Flags().IntVar(&c.lshHashes, "lsh-hashes", c.lshHashes, "MinHash function count")

	// Hide all advanced algorithm flags from help
	// These should be configured in .pyscn.toml or pyproject.toml
	_ = cmd.Flags().MarkHidden("max-distance")
	_ = cmd.Flags().MarkHidden("type1-threshold")
	_ = cmd.Flags().MarkHidden("type2-threshold")
	_ = cmd.Flags().MarkHidden("type3-threshold")
	_ = cmd.Flags().MarkHidden("type4-threshold")
	_ = cmd.Flags().MarkHidden("cost-model")
	_ = cmd.Flags().MarkHidden("group-threshold")
	_ = cmd.Flags().MarkHidden("group-mode")
	_ = cmd.Flags().MarkHidden("k-core-k")
	_ = cmd.Flags().MarkHidden("use-lsh")
	_ = cmd.Flags().MarkHidden("lsh-threshold")
	_ = cmd.Flags().MarkHidden("lsh-bands")
	_ = cmd.Flags().MarkHidden("lsh-rows")
	_ = cmd.Flags().MarkHidden("lsh-hashes")
	_ = cmd.Flags().MarkHidden("ignore-literals")
	_ = cmd.Flags().MarkHidden("ignore-identifiers")
	_ = cmd.Flags().MarkHidden("min-lines")
	_ = cmd.Flags().MarkHidden("min-nodes")
	_ = cmd.Flags().MarkHidden("min-similarity")
	_ = cmd.Flags().MarkHidden("max-similarity")

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
	// Load configuration from pyproject.toml (if available)
	workDir := "."
	if len(paths) > 0 {
		workDir = paths[0]
	}

	config, err := c.loadConfigWithFallback(workDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Apply presets and CLI overrides
	c.applyPresets(config)
	c.applyCliOverrides(config, cmd)
	// Determine output format from flags
	outputFormat, extension, err := c.determineOutputFormat()
	if err != nil {
		return nil, err
	}

	// Parse sort criteria using config value (CLI overrides have already been applied)
	sortBy, err := c.parseSortCriteria(config.Output.SortBy)
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

	// Use paths from CLI args first, then config
	inputPaths := paths
	if len(inputPaths) == 0 {
		inputPaths = config.Input.Paths
	}

	// Parse clone types from config
	configCloneTypes, err := c.parseCloneTypesFromConfig(config.Filtering.EnabledCloneTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse clone types from config: %w", err)
	}

	// Use CLI clone types if specified, otherwise use config
	finalCloneTypes := cloneTypes
	if len(c.cloneTypes) == 0 {
		finalCloneTypes = configCloneTypes
	}

	request := &domain.CloneRequest{
		Paths:               inputPaths,
		Recursive:           config.Input.Recursive,
		IncludePatterns:     config.Input.IncludePatterns,
		ExcludePatterns:     config.Input.ExcludePatterns,
		MinLines:            config.Analysis.MinLines,
		MinNodes:            config.Analysis.MinNodes,
		SimilarityThreshold: config.Thresholds.SimilarityThreshold,
		MaxEditDistance:     config.Analysis.MaxEditDistance,
		IgnoreLiterals:      config.Analysis.IgnoreLiterals,
		IgnoreIdentifiers:   config.Analysis.IgnoreIdentifiers,
		Type1Threshold:      config.Thresholds.Type1Threshold,
		Type2Threshold:      config.Thresholds.Type2Threshold,
		Type3Threshold:      config.Thresholds.Type3Threshold,
		Type4Threshold:      config.Thresholds.Type4Threshold,
		OutputFormat:        outputFormat,
		OutputWriter:        outputWriter,
		OutputPath:          outputPath,
		NoOpen:              c.noOpen,
		ShowDetails:         config.Output.ShowDetails,
		ShowContent:         config.Output.ShowContent,
		SortBy:              sortBy,
		GroupClones:         config.Output.GroupClones,
		GroupMode:           config.Grouping.Mode,
		GroupThreshold:      config.Grouping.Threshold,
		KCoreK:              config.Grouping.KCoreK,
		MinSimilarity:       config.Filtering.MinSimilarity,
		MaxSimilarity:       config.Filtering.MaxSimilarity,
		CloneTypes:          finalCloneTypes,
		ConfigPath:          c.configFile,
		Timeout:             time.Duration(config.Performance.TimeoutSeconds) * time.Second,

		// LSH settings from config
		UseLSH:                 config.LSH.Enabled == "true" || (config.LSH.Enabled == "auto" && c.shouldAutoEnableLSH()),
		LSHSimilarityThreshold: config.LSH.SimilarityThreshold,
		LSHBands:               config.LSH.Bands,
		LSHRows:                config.LSH.Rows,
		LSHHashes:              config.LSH.Hashes,
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

// parseCloneTypesFromConfig parses clone types from config string slice
func (c *CloneCommand) parseCloneTypesFromConfig(typeStrs []string) ([]domain.CloneType, error) {
	var cloneTypes []domain.CloneType

	for _, typeStr := range typeStrs {
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

// shouldAutoEnableLSH determines if LSH should be auto-enabled
func (c *CloneCommand) shouldAutoEnableLSH() bool {
	// Simple heuristic: enable LSH for potentially large codebases
	// This would need actual analysis of project size in a real implementation
	return false // Conservative default for now
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

// loadConfigWithFallback loads configuration using TOML-only strategy
// Priority: pyproject.toml > .pyscn.toml > defaults
func (c *CloneCommand) loadConfigWithFallback(workDir string) (*config.CloneConfig, error) {
	loader := config.NewTomlConfigLoader()
	return loader.LoadConfig(workDir)
}

// applyPresets applies preset configurations
func (c *CloneCommand) applyPresets(cfg *config.CloneConfig) {
	// Handle preset flag first
	if c.preset != "" {
		switch strings.ToLower(c.preset) {
		case "fast":
			c.fast = true
		case "precise":
			c.precise = true
		case "balanced":
			// Use default configuration
		}
	}

	// Apply fast preset
	if c.fast {
		cfg.LSH.Enabled = "true"
		cfg.Grouping.Mode = "connected" // Faster for large scale
	}

	// Apply precise preset
	if c.precise {
		cfg.LSH.Enabled = "false"
		cfg.Grouping.Mode = "star" // More precise grouping
	}
}

// applyCliOverrides applies CLI flag overrides to config
func (c *CloneCommand) applyCliOverrides(cfg *config.CloneConfig, cmd *cobra.Command) {
	// Only override if flags were explicitly set

	if cmd.Flags().Changed("group-mode") {
		cfg.Grouping.Mode = c.groupMode
	}
	if cmd.Flags().Changed("group-threshold") {
		cfg.Grouping.Threshold = c.groupThreshold
	}
	if cmd.Flags().Changed("k-core-k") {
		cfg.Grouping.KCoreK = c.kCoreK
	}

	if cmd.Flags().Changed("use-lsh") {
		if c.useLSH {
			cfg.LSH.Enabled = "true"
		} else {
			cfg.LSH.Enabled = "false"
		}
	}
	if cmd.Flags().Changed("lsh-threshold") {
		cfg.LSH.SimilarityThreshold = c.lshThreshold
	}
	if cmd.Flags().Changed("lsh-bands") {
		cfg.LSH.Bands = c.lshBands
	}
	if cmd.Flags().Changed("lsh-rows") {
		cfg.LSH.Rows = c.lshRows
	}
	if cmd.Flags().Changed("lsh-hashes") {
		cfg.LSH.Hashes = c.lshHashes
	}

	if cmd.Flags().Changed("similarity-threshold") {
		cfg.Thresholds.SimilarityThreshold = c.similarityThreshold
	}
	if cmd.Flags().Changed("min-lines") {
		cfg.Analysis.MinLines = c.minLines
	}
	if cmd.Flags().Changed("min-nodes") {
		cfg.Analysis.MinNodes = c.minNodes
	}
	if cmd.Flags().Changed("sort") {
		cfg.Output.SortBy = c.sortBy
	}
	if cmd.Flags().Changed("details") {
		cfg.Output.ShowDetails = c.showDetails
	}
	if cmd.Flags().Changed("show-content") {
		cfg.Output.ShowContent = c.showContent
	}
	if cmd.Flags().Changed("group") {
		cfg.Output.GroupClones = c.groupClones
	}
}

// Helper function to add the clone command to the root command
func addCloneCommand(rootCmd *cobra.Command) {
	cloneCmd := NewCloneCommand()
	rootCmd.AddCommand(cloneCmd.CreateCobraCommand())
}
