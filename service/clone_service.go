package service

import (
	"context"
	"fmt"
	"os"
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

	allFragments, filesAnalyzed, linesAnalyzed, nodesAnalyzed, err := s.extractFragmentsFromFiles(ctx, filePaths, detector)
	if err != nil {
		return nil, err
	}

	return s.buildCloneResponse(ctx, startTime, detectorConfig, detector, allFragments, filesAnalyzed, linesAnalyzed, nodesAnalyzed, req)
}

func (s *CloneService) extractFragmentsFromFiles(ctx context.Context, filePaths []string, detector *analyzer.CloneDetector) ([]*analyzer.CodeFragment, int, int, int, error) {
	pyParser := parser.New()
	var allFragments []*analyzer.CodeFragment
	linesAnalyzed := 0
	nodesAnalyzed := 0
	filesAnalyzed := 0

	for _, filePath := range filePaths {
		select {
		case <-ctx.Done():
			return nil, 0, 0, 0, fmt.Errorf("clone analysis cancelled: %w", ctx.Err())
		default:
		}

		content, err := readFileContent(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to read file %s: %v\n", filePath, err)
			continue
		}

		parseResult, err := pyParser.Parse(ctx, content)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to parse file %s: %v\n", filePath, err)
			continue
		}
		if parseResult == nil || parseResult.AST == nil {
			fmt.Fprintf(os.Stderr, "Warning: Invalid parse result for file %s\n", filePath)
			continue
		}

		filesAnalyzed++
		linesAnalyzed += countSourceLines(content)

		statsVisitor := parser.NewStatisticsVisitor()
		parseResult.AST.Accept(statsVisitor)
		nodesAnalyzed += statsVisitor.TotalNodes

		astNodes := []*parser.Node{parseResult.AST}
		fragments := detector.ExtractFragmentsWithSource(astNodes, filePath, content)
		allFragments = append(allFragments, fragments...)
	}

	return allFragments, filesAnalyzed, linesAnalyzed, nodesAnalyzed, nil
}

func (s *CloneService) buildCloneResponse(
	ctx context.Context,
	startTime time.Time,
	detectorConfig *analyzer.CloneDetectorConfig,
	detector *analyzer.CloneDetector,
	allFragments []*analyzer.CodeFragment,
	filesAnalyzed int,
	linesAnalyzed int,
	nodesAnalyzed int,
	req *domain.CloneRequest,
) (*domain.CloneResponse, error) {
	if len(allFragments) == 0 {
		return &domain.CloneResponse{
			Clones:      []*domain.Clone{},
			ClonePairs:  []*domain.ClonePair{},
			CloneGroups: []*domain.CloneGroup{},
			Statistics: &domain.CloneStatistics{
				TotalFragments: 0,
				FilesAnalyzed:  filesAnalyzed,
				LinesAnalyzed:  linesAnalyzed,
				NodesAnalyzed:  nodesAnalyzed,
			},
			Request:  req,
			Duration: time.Since(startTime).Milliseconds(),
			Success:  true,
		}, nil
	}

	// Determine whether to use LSH based on configuration and estimated exact-pair cost.
	useLSH := domain.ShouldUseLSHWithPairEstimate(
		req.LSHEnabled,
		len(allFragments),
		req.LSHAutoThreshold,
		detectorConfig.MaxClonePairs,
	)

	// Update detector with LSH decision
	detector.SetUseLSH(useLSH)

	// Detect clones (detector will automatically use LSH or standard algorithm based on UseLSH setting)
	detectionResult := detector.DetectClonesWithLSH(ctx, allFragments)

	// Convert to domain objects
	domainClones, fragmentIDs := s.convertFragmentsToDomainClones(allFragments)
	domainClonePairs := s.convertClonePairsToDomain(detectionResult.Pairs, req.ShowContent)
	domainCloneGroups := s.convertCloneGroupsToDomain(detectionResult.Groups, req.ShowContent, fragmentIDs)

	// Filter results based on request criteria
	domainClonePairs = s.filterClonePairs(domainClonePairs, req)
	domainCloneGroups = s.filterCloneGroups(domainCloneGroups, req)

	// Sort results
	s.sortResults(domainClones, domainClonePairs, domainCloneGroups, req)

	// Build statistics from the filtered results so that counts always match the
	// returned clone pairs and groups. The detector's result already separates
	// raw candidates from detected items; we derive final numbers only from the
	// detected items exposed in the response.
	statistics := s.buildCloneStatistics(detectionResult, domainClonePairs, domainCloneGroups, filesAnalyzed, linesAnalyzed, nodesAnalyzed)

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
		// Use Connected as default to ensure all clone pairs form groups
		groupMode = analyzer.GroupingModeConnected
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
		MinLines:            req.MinLines,
		MinNodes:            req.MinNodes,
		Type1Threshold:      req.Type1Threshold,
		Type2Threshold:      req.Type2Threshold,
		Type3Threshold:      req.Type3Threshold,
		Type4Threshold:      req.Type4Threshold,
		SimilarityThreshold: req.SimilarityThreshold, // User-configurable minimum similarity
		MaxEditDistance:     req.MaxEditDistance,
		IgnoreLiterals:      req.IgnoreLiterals,
		IgnoreIdentifiers:   req.IgnoreIdentifiers,
		SkipDocstrings:      req.SkipDocstrings,
		CostModelType:       "python", // Default to Python cost model
		MaxClonePairs:       10000,    // Default max pairs
		BatchSizeThreshold:  50,       // Default batch size threshold

		// Advanced analysis
		EnableDFAAnalysis: req.EnableDFA,

		// Grouping
		GroupingMode:      groupMode,
		GroupingThreshold: groupThreshold,
		KCoreK:            kVal,
		// LSH (UseLSH will be set dynamically based on fragment count)
		UseLSH:                 false, // Will be overridden after fragment extraction
		LSHSimilarityThreshold: req.LSHSimilarityThreshold,
		LSHBands:               req.LSHBands,
		LSHRows:                req.LSHRows,
		LSHMinHashCount:        req.LSHHashes,
	}
}

// convertFragmentsToDomainClones converts analyzer fragments to domain clones.
//
// It also returns a fragment -> id map so that the same fragment receives the
// same id everywhere it appears in the response (top-level clones[] and the
// per-group clone_groups[].clones[]). The id is the fragment's 1-based index in
// the analyzed fragment set, which is response-unique and stable.
func (s *CloneService) convertFragmentsToDomainClones(
	fragments []*analyzer.CodeFragment,
) ([]*domain.Clone, map[*analyzer.CodeFragment]int) {
	domainClones := make([]*domain.Clone, len(fragments))
	fragmentIDs := make(map[*analyzer.CodeFragment]int, len(fragments))

	for i, fragment := range fragments {
		id := i + 1
		fragmentIDs[fragment] = id
		domainClones[i] = &domain.Clone{
			ID: id,
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

	return domainClones, fragmentIDs
}

// convertClonePairsToDomain converts analyzer clone pairs to domain clone pairs.
func (s *CloneService) convertClonePairsToDomain(clonePairs []*analyzer.ClonePair, includeContent bool) []*domain.ClonePair {
	domainPairs := make([]*domain.ClonePair, len(clonePairs))

	for i, pair := range clonePairs {
		clone1 := &domain.Clone{
			Location: &domain.CloneLocation{
				FilePath:  pair.Fragment1.Location.FilePath,
				StartLine: pair.Fragment1.Location.StartLine,
				EndLine:   pair.Fragment1.Location.EndLine,
				StartCol:  pair.Fragment1.Location.StartCol,
				EndCol:    pair.Fragment1.Location.EndCol,
			},
			Hash:      pair.Fragment1.Hash,
			Size:      pair.Fragment1.Size,
			LineCount: pair.Fragment1.LineCount,
		}
		clone2 := &domain.Clone{
			Location: &domain.CloneLocation{
				FilePath:  pair.Fragment2.Location.FilePath,
				StartLine: pair.Fragment2.Location.StartLine,
				EndLine:   pair.Fragment2.Location.EndLine,
				StartCol:  pair.Fragment2.Location.StartCol,
				EndCol:    pair.Fragment2.Location.EndCol,
			},
			Hash:      pair.Fragment2.Hash,
			Size:      pair.Fragment2.Size,
			LineCount: pair.Fragment2.LineCount,
		}
		if includeContent {
			clone1.Content = pair.Fragment1.Content
			clone2.Content = pair.Fragment2.Content
		}

		domainPairs[i] = &domain.ClonePair{
			ID:         i + 1,
			Clone1:     clone1,
			Clone2:     clone2,
			Similarity: pair.Similarity,
			Distance:   pair.Distance,
			Type:       s.convertCloneType(pair.CloneType),
			Confidence: pair.Confidence,
		}
	}

	return domainPairs
}

// convertCloneGroupsToDomain converts analyzer clone groups to domain clone groups.
func (s *CloneService) convertCloneGroupsToDomain(
	cloneGroups []*analyzer.CloneGroup,
	includeContent bool,
	fragmentIDs map[*analyzer.CodeFragment]int,
) []*domain.CloneGroup {
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
				// Reuse the fragment's response-wide id so a group fragment
				// matches the same fragment in top-level clones[] and ids stay
				// unique across the response (not reset per group).
				ID:   fragmentIDs[fragment],
				Type: s.convertCloneType(group.CloneType),
				Location: &domain.CloneLocation{
					FilePath:  fragment.Location.FilePath,
					StartLine: fragment.Location.StartLine,
					EndLine:   fragment.Location.EndLine,
					StartCol:  fragment.Location.StartCol,
					EndCol:    fragment.Location.EndCol,
				},
				Hash:      fragment.Hash,
				Size:      fragment.Size,
				LineCount: fragment.LineCount,
			}
			if includeContent {
				clone.Content = fragment.Content
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
		return domain.Type4Clone
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

// buildCloneStatistics converts analyzer-level detection statistics into the
// domain statistics attached to the response. All counts are derived from the
// filtered results so that reported numbers always match what is returned.
func (s *CloneService) buildCloneStatistics(
	result *analyzer.CloneDetectionResult,
	pairs []*domain.ClonePair,
	groups []*domain.CloneGroup,
	filesAnalyzed, linesAnalyzed, nodesAnalyzed int,
) *domain.CloneStatistics {
	stats := domain.NewCloneStatistics()
	stats.TotalFragments = result.Statistics.TotalFragments
	stats.TotalClones = countUniqueCloneFragments(pairs, groups)
	stats.TotalClonePairs = len(pairs)
	stats.TotalCloneGroups = len(groups)
	stats.FilesAnalyzed = filesAnalyzed
	stats.LinesAnalyzed = linesAnalyzed
	stats.NodesAnalyzed = nodesAnalyzed

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

// countUniqueCloneFragments counts distinct domain clones that participate in
// at least one clone pair or group. Counting detected items separately from the
// raw fragment collection prevents accidentally reporting the candidate count.
func countUniqueCloneFragments(pairs []*domain.ClonePair, groups []*domain.CloneGroup) int {
	type locKey struct {
		file      string
		startLine int
		endLine   int
	}
	seen := make(map[locKey]struct{})
	addClone := func(c *domain.Clone) {
		if c != nil && c.Location != nil {
			seen[locKey{c.Location.FilePath, c.Location.StartLine, c.Location.EndLine}] = struct{}{}
		}
	}
	for _, p := range pairs {
		addClone(p.Clone1)
		addClone(p.Clone2)
	}
	for _, g := range groups {
		for _, c := range g.Clones {
			addClone(c)
		}
	}
	return len(seen)
}

// readFileContent reads the content of a file
func readFileContent(filePath string) ([]byte, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return content, nil
}
