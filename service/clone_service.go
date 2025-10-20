package service

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/analyzer"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// CloneService implements the domain.CloneService interface
type CloneService struct {
}

// NewCloneService creates a new clone service
func NewCloneService() *CloneService {
	return &CloneService{}
}

// DetectClones performs clone detection on the given request
func (s *CloneService) DetectClones(ctx context.Context, req *domain.CloneRequest) (*domain.CloneResponse, error) {
	// Input validation
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}
	if req == nil {
		return nil, fmt.Errorf("clone request cannot be nil")
	}
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid clone request: %w", err)
	}

	// Use the files already collected by the usecase layer
	// req.Paths now contains actual Python files to analyze
	return s.DetectClonesInFiles(ctx, req.Paths, req)
}

// DetectClonesInFiles performs clone detection on specific files
func (s *CloneService) DetectClonesInFiles(ctx context.Context, filePaths []string, req *domain.CloneRequest) (*domain.CloneResponse, error) {
	// Input validation
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}
	if req == nil {
		return nil, fmt.Errorf("clone request cannot be nil")
	}
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("file paths cannot be empty")
	}

	startTime := time.Now()

	// Apply timeout if specified
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}

	// Progress reporting removed - file parsing is not the bottleneck

	// Create clone detector with configuration
	detectorConfig := s.createDetectorConfig(req)
	detector := analyzer.NewCloneDetector(detectorConfig)

	// Performance optimizations are built into the detector

	// Create Python parser
	pyParser := parser.New()

	// Parse files and extract fragments
	var allFragments []*analyzer.CodeFragment
	linesAnalyzed := 0
	successfullyAnalyzedFilesCount := 0 // Initialize counter for successfully analyzed files

	for _, filePath := range filePaths {
		// Check for context cancellation periodically
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("clone analysis cancelled: %w", ctx.Err())
		default:
		}

		// Progress reporting removed - file parsing is fast

		// Read file content
		content, err := readFileContent(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to read file %s: %v\n", filePath, err)
			continue // Skip files that cannot be read
		}

		// Parse Python file
		parseResult, err := pyParser.Parse(ctx, content)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to parse file %s: %v\n", filePath, err)
			continue // Skip files that cannot be parsed
		}

		// Validate parse result
		if parseResult == nil || parseResult.AST == nil {
			fmt.Fprintf(os.Stderr, "Warning: Invalid parse result for file %s\n", filePath)
			continue // Skip files with invalid parse results
		}

		// If we reached here, the file was successfully read and parsed.
		successfullyAnalyzedFilesCount++

		// Count lines for statistics
		linesAnalyzed += len(strings.Split(string(content), "\n"))

		// Extract code fragments from AST
		astNodes := []*parser.Node{parseResult.AST}
		fragments := detector.ExtractFragments(astNodes, filePath)
		allFragments = append(allFragments, fragments...)
	}

	if len(allFragments) == 0 {
		return &domain.CloneResponse{
			Clones:      []*domain.Clone{},
			ClonePairs:  []*domain.ClonePair{},
			CloneGroups: []*domain.CloneGroup{},
			Statistics: &domain.CloneStatistics{
				FilesAnalyzed: successfullyAnalyzedFilesCount, // Use the new counter
				LinesAnalyzed: linesAnalyzed,
			},
			Request:  req,
			Duration: time.Since(startTime).Milliseconds(),
			Success:  true,
		}, nil
	}

	// Starting actual clone detection (this is the slow part)

	// Determine whether to use LSH based on configuration
	useLSH := domain.ShouldUseLSH(req.LSHEnabled, len(allFragments), req.LSHAutoThreshold)

	// Update detector with LSH decision
	detector.SetUseLSH(useLSH)

	// Detect clones (detector will automatically use LSH or standard algorithm based on UseLSH setting)
	clonePairs, cloneGroups := detector.DetectClonesWithLSH(ctx, allFragments)

	// Convert to domain objects
	domainClones := s.convertFragmentsToDomainClones(allFragments)
	domainClonePairs := s.convertClonePairsToDomain(clonePairs)
	domainCloneGroups := s.convertCloneGroupsToDomain(cloneGroups)

	// Filter results based on request criteria
	domainClonePairs = s.filterClonePairs(domainClonePairs, req)
	domainCloneGroups = s.filterCloneGroups(domainCloneGroups, req)

	// Sort results
	s.sortResults(domainClones, domainClonePairs, domainCloneGroups, req)

	// Create statistics
	statistics := s.createStatistics(domainClones, domainClonePairs, domainCloneGroups, successfullyAnalyzedFilesCount, linesAnalyzed) // Use the new counter

	duration := time.Since(startTime).Milliseconds()
	// s.progress.Complete(fmt.Sprintf("Clone detection completed in %dms. Found %d clone pairs in %d groups.",
	//	duration, len(domainClonePairs), len(domainCloneGroups)))

	return &domain.CloneResponse{
		Clones:      domainClones,
		ClonePairs:  domainClonePairs,
		CloneGroups: domainCloneGroups,
		Statistics:  statistics,
		Request:     req,
		Duration:    duration,
		Success:     true,
	}, nil
}

// ComputeSimilarity computes similarity between two code fragments
func (s *CloneService) ComputeSimilarity(ctx context.Context, fragment1, fragment2 string) (float64, error) {
	// Input validation
	if fragment1 == "" || fragment2 == "" {
		return 0.0, fmt.Errorf("fragments cannot be empty")
	}

	if ctx == nil {
		return 0.0, fmt.Errorf("context cannot be nil")
	}

	// Check for excessively large fragments to prevent resource exhaustion
	const maxFragmentSize = 1024 * 1024 // 1MB limit
	if len(fragment1) > maxFragmentSize || len(fragment2) > maxFragmentSize {
		return 0.0, fmt.Errorf("fragment size exceeds maximum allowed size of %d bytes", maxFragmentSize)
	}

	// Parse both fragments
	pyParser := parser.New()

	result1, err := pyParser.Parse(ctx, []byte(fragment1))
	if err != nil {
		return 0.0, fmt.Errorf("failed to parse fragment1: %w", err)
	}
	if result1 == nil || result1.AST == nil {
		return 0.0, fmt.Errorf("fragment1 parsing returned nil result or AST")
	}

	result2, err := pyParser.Parse(ctx, []byte(fragment2))
	if err != nil {
		return 0.0, fmt.Errorf("failed to parse fragment2: %w", err)
	}
	if result2 == nil || result2.AST == nil {
		return 0.0, fmt.Errorf("fragment2 parsing returned nil result or AST")
	}

	// Convert AST nodes to tree nodes for APTED
	converter := analyzer.NewTreeConverter()
	tree1 := converter.ConvertAST(result1.AST)
	tree2 := converter.ConvertAST(result2.AST)

	if tree1 == nil || tree2 == nil {
		return 0.0, fmt.Errorf("failed to convert AST to tree nodes")
	}

	// Use APTED to compute similarity
	costModel := analyzer.NewPythonCostModel()
	aptedAnalyzer := analyzer.NewAPTEDAnalyzer(costModel)

	similarity := aptedAnalyzer.ComputeSimilarity(tree1, tree2)
	return similarity, nil
}

// createDetectorConfig creates a clone detector configuration from the domain request
func (s *CloneService) createDetectorConfig(req *domain.CloneRequest) *analyzer.CloneDetectorConfig {
	// Determine grouping defaults
	groupMode := analyzer.GroupingMode(req.GroupMode)
	if groupMode == "" {
		// Use K-Core as default for better performance and quality balance
		groupMode = analyzer.GroupingModeKCore
	}
	groupThreshold := req.GroupThreshold
	if groupThreshold <= 0 {
		groupThreshold = req.SimilarityThreshold
		if groupThreshold <= 0 {
			groupThreshold = req.Type3Threshold
		}
	}
	kVal := req.KC