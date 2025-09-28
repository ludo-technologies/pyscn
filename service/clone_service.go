package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/analyzer"
	"github.com/ludo-technologies/pyscn/internal/constants"
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
	var fileErrors []error
	successfulFiles := 0

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
			fileErrors = append(fileErrors, fmt.Errorf("failed to read file %s: %w", filePath, err))
			continue
		}

		// Parse Python file
		parseResult, err := pyParser.Parse(ctx, content)
		if err != nil {
			fileErrors = append(fileErrors, fmt.Errorf("failed to parse file %s: %w", filePath, err))
			continue
		}

		// Validate parse result
		if parseResult == nil || parseResult.AST == nil {
			fileErrors = append(fileErrors, fmt.Errorf("invalid parse result for file %s", filePath))
			continue
		}

		// Count lines for statistics
		linesAnalyzed += len(strings.Split(string(content), "\n"))
		successfulFiles++

		// Extract code fragments from AST
		if parseResult.AST != nil {
			// Convert single AST node to slice for ExtractFragments
			astNodes := []*parser.Node{parseResult.AST}
			fragments := detector.ExtractFragments(astNodes, filePath)
			allFragments = append(allFragments, fragments...)
		}
	}

	// Check if too many files failed
	if len(fileErrors) > 0 {
		// Log warnings for failed files
		for _, err := range fileErrors {
			log.Printf("Warning: %v", err)
		}

		// If all files failed and no fragments were extracted, still return success with empty results
		// This handles cases where all input files are non-existent or invalid
		if successfulFiles == 0 && len(allFragments) == 0 {
			// Continue to return empty results below
			log.Printf("Warning: All %d files could not be processed, returning empty results", len(filePaths))
		} else if successfulFiles > 0 {
			// Some files succeeded, continue with analysis
			failedRatio := float64(len(fileErrors)) / float64(len(filePaths))
			if failedRatio > 0.5 {
				// More than 50% failed but some succeeded - log warning but continue
				log.Printf("Warning: %d out of %d files could not be processed", len(fileErrors), len(filePaths))
			}
		}
	}

	if len(allFragments) == 0 {
		return &domain.CloneResponse{
			Clones:      []*domain.Clone{},
			ClonePairs:  []*domain.ClonePair{},
			CloneGroups: []*domain.CloneGroup{},
			Statistics: &domain.CloneStatistics{
				FilesAnalyzed: len(filePaths),
				LinesAnalyzed: linesAnalyzed,
			},
			Request:  req,
			Duration: time.Since(startTime).Milliseconds(),
			Success:  true,
		}, nil
	}

	// Starting actual clone detection (this is the slow part)

	// Detect clones (use LSH if enabled)
	var clonePairs []*analyzer.ClonePair
	var cloneGroups []*analyzer.CloneGroup
	if detectorConfig.UseLSH {
		clonePairs, cloneGroups = detector.DetectClonesWithLSH(ctx, allFragments)
	} else {
		clonePairs, cloneGroups = detector.DetectClonesWithContext(ctx, allFragments)
	}

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
	statistics := s.createStatistics(domainClones, domainClonePairs, domainCloneGroups, len(filePaths), linesAnalyzed)

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
	if len(fragment1) > constants.DefaultMaxFragmentSize || len(fragment2) > constants.DefaultMaxFragmentSize {
		return 0.0, fmt.Errorf("fragment size exceeds maximum allowed size of %d bytes", constants.DefaultMaxFragmentSize)
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
	kVal := req.KCoreK
	if kVal < 2 {
		kVal = 2
	}

	return &analyzer.CloneDetectorConfig{
		MinLines:           req.MinLines,
		MinNodes:           req.MinNodes,
		Type1Threshold:     req.Type1Threshold,
		Type2Threshold:     req.Type2Threshold,
		Type3Threshold:     req.Type3Threshold,
		Type4Threshold:     req.Type4Threshold,
		MaxEditDistance:    req.MaxEditDistance,
		IgnoreLiterals:     req.IgnoreLiterals,
		IgnoreIdentifiers:  req.IgnoreIdentifiers,
		CostModelType:      "python", // Default to Python cost model
		MaxClonePairs:      10000,    // Default max pairs
		BatchSizeThreshold: 50,       // Default batch size threshold

		// Grouping
		GroupingMode:      groupMode,
		GroupingThreshold: groupThreshold,
		KCoreK:            kVal,
		// LSH
		UseLSH:                 req.UseLSH,
		LSHSimilarityThreshold: req.LSHSimilarityThreshold,
		LSHBands:               req.LSHBands,
		LSHRows:                req.LSHRows,
		LSHMinHashCount:        req.LSHHashes,
	}
}

// convertFragmentsToDomainClones converts analyzer fragments to domain clones
func (s *CloneService) convertFragmentsToDomainClones(fragments []*analyzer.CodeFragment) []*domain.Clone {
	domainClones := make([]*domain.Clone, len(fragments))

	for i, fragment := range fragments {
		domainClones[i] = &domain.Clone{
			ID: i + 1,
			Location: &domain.CloneLocation{
				FilePath:  fragment.Location.FilePath,
				StartLine: fragment.Location.StartLine,
				EndLine:   fragment.Location.EndLine,
				StartCol:  fragment.Location.StartCol,
				EndCol:    fragment.Location.EndCol,
			},
			Content:    fragment.Content,
			Hash:       fragment.Hash,
			Size:       fragment.Size,
			LineCount:  fragment.LineCount,
			Complexity: fragment.Complexity,
		}
	}

	return domainClones
}

// convertClonePairsToDomain converts analyzer clone pairs to domain clone pairs
func (s *CloneService) convertClonePairsToDomain(clonePairs []*analyzer.ClonePair) []*domain.ClonePair {
	domainPairs := make([]*domain.ClonePair, len(clonePairs))

	for i, pair := range clonePairs {
		domainPairs[i] = &domain.ClonePair{
			ID: i + 1,
			Clone1: &domain.Clone{
				Location: &domain.CloneLocation{
					FilePath:  pair.Fragment1.Location.FilePath,
					StartLine: pair.Fragment1.Location.StartLine,
					EndLine:   pair.Fragment1.Location.EndLine,
					StartCol:  pair.Fragment1.Location.StartCol,
					EndCol:    pair.Fragment1.Location.EndCol,
				},
				Size:      pair.Fragment1.Size,
				LineCount: pair.Fragment1.LineCount,
			},
			Clone2: &domain.Clone{
				Location: &domain.CloneLocation{
					FilePath:  pair.Fragment2.Location.FilePath,
					StartLine: pair.Fragment2.Location.StartLine,
					EndLine:   pair.Fragment2.Location.EndLine,
					StartCol:  pair.Fragment2.Location.StartCol,
					EndCol:    pair.Fragment2.Location.EndCol,
				},
				Size:      pair.Fragment2.Size,
				LineCount: pair.Fragment2.LineCount,
			},
			Similarity: pair.Similarity,
			Distance:   pair.Distance,
			Type:       s.convertCloneType(pair.CloneType),
			Confidence: pair.Confidence,
		}
	}

	return domainPairs
}

// convertCloneGroupsToDomain converts analyzer clone groups to domain clone groups
func (s *CloneService) convertCloneGroupsToDomain(cloneGroups []*analyzer.CloneGroup) []*domain.CloneGroup {
	domainGroups := make([]*domain.CloneGroup, len(cloneGroups))

	for i, group := range cloneGroups {
		domainGroup := &domain.CloneGroup{
			ID:         group.ID,
			Type:       s.convertCloneType(group.CloneType),
			Similarity: group.Similarity,
			Size:       group.Size,
			Clones:     []*domain.Clone{},
		}

		// Convert fragments to clones
		for _, fragment := range group.Fragments {
			clone := &domain.Clone{
				Location: &domain.CloneLocation{
					FilePath:  fragment.Location.FilePath,
					StartLine: fragment.Location.StartLine,
					EndLine:   fragment.Location.EndLine,
					StartCol:  fragment.Location.StartCol,
					EndCol:    fragment.Location.EndCol,
				},
				Size:      fragment.Size,
				LineCount: fragment.LineCount,
			}
			domainGroup.AddClone(clone)
		}

		domainGroups[i] = domainGroup
	}

	return domainGroups
}

// convertCloneType converts analyzer clone type to domain clone type
func (s *CloneService) convertCloneType(cloneType analyzer.CloneType) domain.CloneType {
	switch cloneType {
	case analyzer.Type1Clone:
		return domain.Type1Clone
	case analyzer.Type2Clone:
		return domain.Type2Clone
	case analyzer.Type3Clone:
		return domain.Type3Clone
	case analyzer.Type4Clone:
		return domain.Type4Clone
	default:
		return domain.Type1Clone
	}
}

// filterClonePairs filters clone pairs based on request criteria
func (s *CloneService) filterClonePairs(pairs []*domain.ClonePair, req *domain.CloneRequest) []*domain.ClonePair {
	var filtered []*domain.ClonePair

	for _, pair := range pairs {
		// Filter by similarity range
		if pair.Similarity < req.MinSimilarity || pair.Similarity > req.MaxSimilarity {
			continue
		}

		// Filter by clone types
		typeEnabled := false
		for _, enabledType := range req.CloneTypes {
			if pair.Type == enabledType {
				typeEnabled = true
				break
			}
		}
		if !typeEnabled {
			continue
		}

		filtered = append(filtered, pair)
	}

	return filtered
}

// filterCloneGroups filters clone groups based on request criteria
func (s *CloneService) filterCloneGroups(groups []*domain.CloneGroup, req *domain.CloneRequest) []*domain.CloneGroup {
	var filtered []*domain.CloneGroup

	for _, group := range groups {
		// Filter by similarity range
		if group.Similarity < req.MinSimilarity || group.Similarity > req.MaxSimilarity {
			continue
		}

		// Filter by clone types
		typeEnabled := false
		for _, enabledType := range req.CloneTypes {
			if group.Type == enabledType {
				typeEnabled = true
				break
			}
		}
		if !typeEnabled {
			continue
		}

		filtered = append(filtered, group)
	}

	return filtered
}

// sortResults sorts the results based on request criteria
func (s *CloneService) sortResults(clones []*domain.Clone, pairs []*domain.ClonePair, groups []*domain.CloneGroup, req *domain.CloneRequest) {
	// Implementation would depend on the specific sort criteria
	// For now, we'll keep the default ordering from the detector
}

// createStatistics creates clone detection statistics
func (s *CloneService) createStatistics(clones []*domain.Clone, pairs []*domain.ClonePair, groups []*domain.CloneGroup, filesAnalyzed, linesAnalyzed int) *domain.CloneStatistics {
	stats := domain.NewCloneStatistics()
	stats.TotalClones = len(clones)
	stats.TotalClonePairs = len(pairs)
	stats.TotalCloneGroups = len(groups)
	stats.FilesAnalyzed = filesAnalyzed
	stats.LinesAnalyzed = linesAnalyzed

	// Count by type
	for _, pair := range pairs {
		typeStr := pair.Type.String()
		stats.ClonesByType[typeStr]++
	}

	// Calculate average similarity
	if len(pairs) > 0 {
		totalSimilarity := 0.0
		for _, pair := range pairs {
			totalSimilarity += pair.Similarity
		}
		stats.AverageSimilarity = totalSimilarity / float64(len(pairs))
	}

	return stats
}

// readFileContent reads the content of a file
func readFileContent(filePath string) ([]byte, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return content, nil
}
