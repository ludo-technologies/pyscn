package app

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
	svc "github.com/ludo-technologies/pyscn/service"
)

// CloneUseCase orchestrates clone detection operations
type CloneUseCase struct {
	service      domain.CloneService
	fileReader   domain.FileReader
	formatter    domain.CloneOutputFormatter
	configLoader domain.CloneConfigurationLoader
	output       domain.ReportWriter
}

// NewCloneUseCase creates a new clone use case with the given dependencies
func NewCloneUseCase(
	service domain.CloneService,
	fileReader domain.FileReader,
	formatter domain.CloneOutputFormatter,
	configLoader domain.CloneConfigurationLoader,
) *CloneUseCase {
	return &CloneUseCase{
		service:      service,
		fileReader:   fileReader,
		formatter:    formatter,
		configLoader: configLoader,
		// Default implementation; CLI may override via builder
		output: svc.NewFileOutputWriter(nil),
	}
}

// Execute executes the clone detection use case
func (uc *CloneUseCase) Execute(ctx context.Context, req domain.CloneRequest) error {
	startTime := time.Now()

	// Step 1: Validate the request
	if err := req.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Step 2: Load configuration if specified or try default config
	if req.ConfigPath != "" {
		configReq, err := uc.configLoader.LoadCloneConfig(req.ConfigPath)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Merge configuration with request (request takes precedence)
		req = uc.mergeConfiguration(*configReq, req)
	} else if uc.configLoader != nil {
		// Try to load default configuration
		defaultConfigReq := uc.configLoader.GetDefaultCloneConfig()
		if defaultConfigReq != nil {
			// Merge default configuration with request (request takes precedence)
			req = uc.mergeConfiguration(*defaultConfigReq, req)
		}
	}

	// Step 3: Collect files to analyze
	files, err := uc.fileReader.CollectPythonFiles(req.Paths, req.Recursive, req.IncludePatterns, req.ExcludePatterns)
	if err != nil {
		return fmt.Errorf("failed to collect files: %w", err)
	}

	if len(files) == 0 {
		return uc.outputEmptyResults(req)
	}

	// Create a copy of the request to avoid modifying the original
	reqCopy := req
	reqCopy.Paths = files

	// Step 4: Perform clone detection
	response, err := uc.service.DetectClones(ctx, &reqCopy)
	if err != nil {
		return fmt.Errorf("clone detection failed: %w", err)
	}

	// Step 5: Update response with timing information
	response.Duration = time.Since(startTime).Milliseconds()

	// Step 6: Format and output results
	if !req.HasValidOutputWriter() && req.OutputPath == "" {
		return fmt.Errorf("no valid output writer or output path specified")
	}

	// Delegate output handling to ReportWriter
	var out io.Writer
	if req.OutputPath == "" {
		out = req.OutputWriter
	}
	if err := uc.output.Write(out, req.OutputPath, req.OutputFormat, req.NoOpen, func(w io.Writer) error {
		return uc.formatter.FormatCloneResponse(response, req.OutputFormat, w)
	}); err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	return nil
}

// ExecuteAndReturn performs clone detection and returns the response without formatting
func (uc *CloneUseCase) ExecuteAndReturn(ctx context.Context, req domain.CloneRequest) (*domain.CloneResponse, error) {
	startTime := time.Now()

	// Step 1: Basic validation
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if len(req.Paths) == 0 {
		return nil, fmt.Errorf("no paths specified for clone detection")
	}

	// Step 2: Load configuration if specified or try default config
	if req.ConfigPath != "" {
		configReq, err := uc.configLoader.LoadCloneConfig(req.ConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load configuration: %w", err)
		}

		// Merge configuration with request (request takes precedence)
		req = uc.mergeConfiguration(*configReq, req)
	} else if uc.configLoader != nil {
		// Try to load default configuration
		defaultConfigReq := uc.configLoader.GetDefaultCloneConfig()
		if defaultConfigReq != nil {
			// Merge default configuration with request (request takes precedence)
			req = uc.mergeConfiguration(*defaultConfigReq, req)
		}
	}

	// Step 3: Collect files to analyze
	if uc.fileReader == nil {
		return nil, fmt.Errorf("file reader not initialized")
	}

	// Check if all paths are already files (not directories)
	// This happens when called from AnalyzeUseCase which pre-collects files
	allFiles := true
	for _, path := range req.Paths {
		if !uc.fileReader.IsValidPythonFile(path) {
			allFiles = false
			break
		}
		exists, err := uc.fileReader.FileExists(path)
		if err != nil || !exists {
			allFiles = false
			break
		}
	}

	var files []string
	if allFiles {
		// All paths are already files, no need to collect again
		files = req.Paths
	} else {
		// Collect Python files from directories
		var err error
		files, err = uc.fileReader.CollectPythonFiles(
			req.Paths,
			req.Recursive,
			req.IncludePatterns,
			req.ExcludePatterns,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to collect files: %w", err)
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no Python files found in the specified paths")
	}

	// Step 3: Update request with discovered files
	req.Paths = files

	// Step 4: Perform clone detection
	response, err := uc.service.DetectClones(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("clone detection failed: %w", err)
	}

	// Step 5: Update response with timing information
	response.Duration = time.Since(startTime).Milliseconds()

	return response, nil
}

// ExecuteWithFiles executes clone detection on specific files
func (uc *CloneUseCase) ExecuteWithFiles(ctx context.Context, filePaths []string, req domain.CloneRequest) error {
	startTime := time.Now()

	// Validate the request
	if err := req.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Validate files exist and are Python files
	validFiles := []string{}
	for _, filePath := range filePaths {
		if uc.fileReader.IsValidPythonFile(filePath) {
			validFiles = append(validFiles, filePath)
		}
		// Non-Python files are silently skipped
	}

	if len(validFiles) == 0 {
		return uc.outputEmptyResults(req)
	}

	// Perform clone detection on specific files
	response, err := uc.service.DetectClonesInFiles(ctx, validFiles, &req)
	if err != nil {
		return fmt.Errorf("clone detection failed: %w", err)
	}

	// Update response with timing information
	response.Duration = time.Since(startTime).Milliseconds()

	// Format and output results
	if !req.HasValidOutputWriter() && req.OutputPath == "" {
		return fmt.Errorf("no valid output writer or output path specified")
	}

	if err := uc.formatter.FormatCloneResponse(response, req.OutputFormat, req.OutputWriter); err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	return nil
}

// ComputeFragmentSimilarity computes similarity between two code fragments
func (uc *CloneUseCase) ComputeFragmentSimilarity(ctx context.Context, fragment1, fragment2 string) (float64, error) {
	similarity, err := uc.service.ComputeSimilarity(ctx, fragment1, fragment2)
	if err != nil {
		return 0.0, fmt.Errorf("failed to compute similarity: %w", err)
	}

	return similarity, nil
}

// SaveConfiguration saves the current clone detection configuration
func (uc *CloneUseCase) SaveConfiguration(req domain.CloneRequest, configPath string) error {
	if err := uc.configLoader.SaveCloneConfig(&req, configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

// mergeConfiguration merges configuration from file with request parameters
// Request parameters take precedence over configuration file values
func (uc *CloneUseCase) mergeConfiguration(configReq, requestReq domain.CloneRequest) domain.CloneRequest {
	// Start with configuration from file
	merged := configReq

	// Override with non-default request values
	if len(requestReq.Paths) > 0 {
		merged.Paths = requestReq.Paths
	}

	// Override boolean flags if explicitly set
	merged.Recursive = requestReq.Recursive
	merged.ShowDetails = requestReq.ShowDetails
	merged.ShowContent = requestReq.ShowContent
	merged.GroupClones = requestReq.GroupClones

	// Override numeric values if they differ from defaults
	defaultReq := domain.DefaultCloneRequest()

	if requestReq.MinLines != defaultReq.MinLines {
		merged.MinLines = requestReq.MinLines
	}
	if requestReq.MinNodes != defaultReq.MinNodes {
		merged.MinNodes = requestReq.MinNodes
	}
	if requestReq.SimilarityThreshold != defaultReq.SimilarityThreshold {
		merged.SimilarityThreshold = requestReq.SimilarityThreshold
	}
	if requestReq.MaxEditDistance != defaultReq.MaxEditDistance {
		merged.MaxEditDistance = requestReq.MaxEditDistance
	}

	// Override threshold values if different from defaults
	if requestReq.Type1Threshold != defaultReq.Type1Threshold {
		merged.Type1Threshold = requestReq.Type1Threshold
	}
	if requestReq.Type2Threshold != defaultReq.Type2Threshold {
		merged.Type2Threshold = requestReq.Type2Threshold
	}
	if requestReq.Type3Threshold != defaultReq.Type3Threshold {
		merged.Type3Threshold = requestReq.Type3Threshold
	}
	if requestReq.Type4Threshold != defaultReq.Type4Threshold {
		merged.Type4Threshold = requestReq.Type4Threshold
	}

	// Always use request values for output settings
	merged.OutputFormat = requestReq.OutputFormat
	merged.OutputWriter = requestReq.OutputWriter
	merged.OutputPath = requestReq.OutputPath
	merged.NoOpen = requestReq.NoOpen
	merged.SortBy = requestReq.SortBy

	// Override patterns if provided
	if len(requestReq.IncludePatterns) > 0 {
		merged.IncludePatterns = requestReq.IncludePatterns
	}
	if len(requestReq.ExcludePatterns) > 0 {
		merged.ExcludePatterns = requestReq.ExcludePatterns
	}

	// Override clone types if provided
	if len(requestReq.CloneTypes) > 0 && len(requestReq.CloneTypes) != len(defaultReq.CloneTypes) {
		merged.CloneTypes = requestReq.CloneTypes
	}

	return merged
}

// outputEmptyResults outputs empty results when no files are found
func (uc *CloneUseCase) outputEmptyResults(req domain.CloneRequest) error {
	emptyResponse := &domain.CloneResponse{
		Clones:      []*domain.Clone{},
		ClonePairs:  []*domain.ClonePair{},
		CloneGroups: []*domain.CloneGroup{},
		Statistics: &domain.CloneStatistics{
			TotalClones:       0,
			TotalClonePairs:   0,
			TotalCloneGroups:  0,
			ClonesByType:      make(map[string]int),
			AverageSimilarity: 0.0,
			LinesAnalyzed:     0,
			FilesAnalyzed:     0,
		},
		Request:  &req,
		Duration: 0,
		Success:  true,
	}

	if req.HasValidOutputWriter() {
		return uc.formatter.FormatCloneResponse(emptyResponse, req.OutputFormat, req.OutputWriter)
	}

	return nil
}

// CloneUseCaseBuilder helps build CloneUseCase with dependencies
type CloneUseCaseBuilder struct {
	service      domain.CloneService
	fileReader   domain.FileReader
	formatter    domain.CloneOutputFormatter
	configLoader domain.CloneConfigurationLoader
	output       domain.ReportWriter
}

// NewCloneUseCaseBuilder creates a new builder for CloneUseCase
func NewCloneUseCaseBuilder() *CloneUseCaseBuilder {
	return &CloneUseCaseBuilder{}
}

// WithService sets the clone service
func (b *CloneUseCaseBuilder) WithService(service domain.CloneService) *CloneUseCaseBuilder {
	b.service = service
	return b
}

// WithFileReader sets the file reader
func (b *CloneUseCaseBuilder) WithFileReader(fileReader domain.FileReader) *CloneUseCaseBuilder {
	b.fileReader = fileReader
	return b
}

// WithFormatter sets the output formatter
func (b *CloneUseCaseBuilder) WithFormatter(formatter domain.CloneOutputFormatter) *CloneUseCaseBuilder {
	b.formatter = formatter
	return b
}

// WithConfigLoader sets the configuration loader
func (b *CloneUseCaseBuilder) WithConfigLoader(configLoader domain.CloneConfigurationLoader) *CloneUseCaseBuilder {
	b.configLoader = configLoader
	return b
}

// WithOutputWriter sets the report writer
func (b *CloneUseCaseBuilder) WithOutputWriter(output domain.ReportWriter) *CloneUseCaseBuilder {
	b.output = output
	return b
}

// Build creates the CloneUseCase with the configured dependencies
func (b *CloneUseCaseBuilder) Build() (*CloneUseCase, error) {
	if b.service == nil {
		return nil, fmt.Errorf("clone service is required")
	}
	if b.fileReader == nil {
		return nil, fmt.Errorf("file reader is required")
	}
	if b.formatter == nil {
		return nil, fmt.Errorf("output formatter is required")
	}
	if b.configLoader == nil {
		return nil, fmt.Errorf("configuration loader is required")
	}

	uc := NewCloneUseCase(
		b.service,
		b.fileReader,
		b.formatter,
		b.configLoader,
	)
	if b.output != nil {
		uc.output = b.output
	}
	return uc, nil
}
