package analyzer

import (
	"bytes"
	"container/heap"
	"context"
	"fmt"
	"math"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// CloneType represents different types of code clones
type CloneType int

const (
	// Type1Clone - Identical code fragments (except whitespace and comments)
	Type1Clone CloneType = iota + 1
	// Type2Clone - Syntactically identical but with different identifiers/literals
	Type2Clone
	// Type3Clone - Syntactically similar with small modifications
	Type3Clone
	// Type4Clone - Functionally similar but syntactically different
	Type4Clone
)

// String returns string representation of CloneType
func (ct CloneType) String() string {
	switch ct {
	case Type1Clone:
		return "Type-1 (Identical)"
	case Type2Clone:
		return "Type-2 (Renamed)"
	case Type3Clone:
		return "Type-3 (Near-Miss)"
	case Type4Clone:
		return "Type-4 (Semantic)"
	default:
		return "Unknown"
	}
}

// CodeLocation represents a location in source code
type CodeLocation struct {
	FilePath  string
	StartLine int
	EndLine   int
	StartCol  int
	EndCol    int
}

// String returns string representation of CodeLocation
func (cl *CodeLocation) String() string {
	return fmt.Sprintf("%s:%d:%d-%d:%d", cl.FilePath, cl.StartLine, cl.StartCol, cl.EndLine, cl.EndCol)
}

// CodeFragment represents a fragment of code
type CodeFragment struct {
	Location   *CodeLocation
	ASTNode    *parser.Node
	TreeNode   *TreeNode
	Content    string   // Original source code content
	Hash       string   // Hash for quick comparison
	Size       int      // Number of AST nodes
	LineCount  int      // Number of source lines
	Complexity int      // Cyclomatic complexity (if applicable)
	Features   []string // Detector-populated clone feature cache for this fragment's tree
}

// NewCodeFragment creates a new code fragment
func NewCodeFragment(location *CodeLocation, astNode *parser.Node, content string) *CodeFragment {
	return &CodeFragment{
		Location:  location,
		ASTNode:   astNode,
		Content:   content,
		Size:      calculateASTSize(astNode),
		LineCount: location.EndLine - location.StartLine + 1,
	}
}

// calculateASTSize calculates the number of nodes in an AST
func calculateASTSize(node *parser.Node) int {
	if node == nil {
		return 0
	}

	size := 1
	for _, child := range parser.OrderedChildren(node, nil) {
		size += calculateASTSize(child)
	}

	return size
}

// ClonePair represents a pair of similar code fragments
type ClonePair struct {
	Fragment1  *CodeFragment
	Fragment2  *CodeFragment
	Similarity float64   // Similarity score (0.0 to 1.0)
	Distance   float64   // Edit distance
	CloneType  CloneType // Type of clone detected
	Confidence float64   // Confidence in the detection (0.0 to 1.0)
}

// String returns string representation of ClonePair
func (cp *ClonePair) String() string {
	return fmt.Sprintf("%s clone: %s <-> %s (similarity: %.3f)",
		cp.CloneType.String(),
		cp.Fragment1.Location.String(),
		cp.Fragment2.Location.String(),
		cp.Similarity)
}

// CloneGroup represents a group of similar code fragments
type CloneGroup struct {
	ID         int             // Unique identifier for this group
	Fragments  []*CodeFragment // All fragments in this group
	CloneType  CloneType       // Primary type of clones in this group
	Similarity float64         // Average similarity within the group
	Size       int             // Number of fragments
}

// NewCloneGroup creates a new clone group
func NewCloneGroup(id int) *CloneGroup {
	return &CloneGroup{
		ID:        id,
		Fragments: []*CodeFragment{},
	}
}

// AddFragment adds a fragment to the clone group
func (cg *CloneGroup) AddFragment(fragment *CodeFragment) {
	cg.Fragments = append(cg.Fragments, fragment)
	cg.Size = len(cg.Fragments)
}

// CloneDetectorConfig holds configuration for clone detection
type CloneDetectorConfig struct {
	// Minimum number of lines for a code fragment to be considered
	MinLines int

	// Minimum number of AST nodes for a code fragment
	MinNodes int

	// Similarity thresholds for different clone types
	Type1Threshold float64 // Usually > domain.DefaultType1CloneThreshold
	Type2Threshold float64 // Usually > domain.DefaultType2CloneThreshold
	Type3Threshold float64 // Usually > domain.DefaultType3CloneThreshold
	Type4Threshold float64 // Usually > domain.DefaultType4CloneThreshold

	// Minimum similarity threshold for clone reporting (user-configurable via --clone-threshold)
	SimilarityThreshold float64

	// Maximum edit distance allowed
	MaxEditDistance float64

	// Whether to ignore differences in literals
	IgnoreLiterals bool

	// Whether to ignore differences in identifiers
	IgnoreIdentifiers bool

	// Whether to skip docstrings from AST comparison (default: true)
	// Docstrings are the first Expr(Constant(str)) in function/class/module bodies
	SkipDocstrings bool

	// Cost model to use for APTED
	CostModelType string // "default", "python", "weighted"

	// Performance tuning parameters
	MaxClonePairs      int // Maximum pairs to keep in memory
	BatchSizeThreshold int // Minimum fragments to trigger batching
	BatchSizeLarge     int // Batch size for normal projects
	BatchSizeSmall     int // Batch size for large projects
	LargeProjectSize   int // Fragment count threshold for large projects
	MaxGoroutines      int // Goroutines for parallel pair comparison (0 = all CPUs)

	// Grouping configuration
	GroupingMode      GroupingMode // デフォルト: GroupingModeConnected
	GroupingThreshold float64      // デフォルト: Type3Threshold
	KCoreK            int          // デフォルト: 2

	// LSH Configuration (optional, opt-in)
	UseLSH                 bool    // Enable LSH acceleration
	LSHSimilarityThreshold float64 // Candidate threshold using MinHash similarity
	LSHBands               int     // Number of LSH bands (default: 32)
	LSHRows                int     // Rows per band (default: 4)
	LSHMinHashCount        int     // Number of MinHash functions (default: 128)
	LSHMaxCandidates       int     // Maximum candidates returned per LSH query

	// Multi-dimensional classification (optional, opt-in)
	EnableMultiDimensionalAnalysis bool // Enable multi-dimensional clone type classification
	EnableTextualAnalysis          bool // Enable Type-1 textual analysis (increases memory usage)
	EnableSemanticAnalysis         bool // Enable Type-4 semantic/CFG analysis (increases CPU usage)
	EnableDFAAnalysis              bool // Enable Data Flow Analysis for enhanced Type-4 detection

	// Framework pattern handling (reduces false positives for dataclass, Pydantic, etc.)
	ReduceBoilerplateSimilarity bool    // Apply lower weight to boilerplate nodes (default: true)
	BoilerplateMultiplier       float64 // Cost multiplier for boilerplate nodes (default: 0.1)
}

// DefaultCloneDetectorConfig returns default configuration
func DefaultCloneDetectorConfig() *CloneDetectorConfig {
	return &CloneDetectorConfig{
		MinLines:          5,
		MinNodes:          10,
		Type1Threshold:    domain.DefaultType1CloneThreshold,
		Type2Threshold:    domain.DefaultType2CloneThreshold,
		Type3Threshold:    domain.DefaultType3CloneThreshold,
		Type4Threshold:    domain.DefaultType4CloneThreshold,
		MaxEditDistance:   50.0,
		IgnoreLiterals:    false,
		IgnoreIdentifiers: false,
		SkipDocstrings:    true,
		CostModelType:     "python",
		// Performance parameters
		MaxClonePairs:      10000,
		BatchSizeThreshold: 50,
		BatchSizeLarge:     100,
		BatchSizeSmall:     50,
		LargeProjectSize:   500,

		// Grouping defaults
		GroupingMode:      GroupingModeConnected,
		GroupingThreshold: domain.DefaultType4CloneThreshold,
		KCoreK:            2,

		// LSH defaults (opt-in)
		UseLSH:                 false,
		LSHSimilarityThreshold: 0.50,
		LSHBands:               32,
		LSHRows:                4,
		LSHMinHashCount:        128,
		LSHMaxCandidates:       defaultLSHMaxCandidates,

		// Multi-dimensional classification defaults (opt-in)
		EnableMultiDimensionalAnalysis: false,
		EnableTextualAnalysis:          false,
		EnableSemanticAnalysis:         false,

		// Framework pattern handling defaults (enabled by default to reduce false positives)
		ReduceBoilerplateSimilarity: true,
		BoilerplateMultiplier:       0.1,
	}
}

// jaccardRejectionThreshold is the Jaccard similarity below which pairs are
// rejected without running APTED. At 0.10, virtually no true Type-3/4 clones
// are missed while the vast majority of non-clone pairs are skipped.
const jaccardRejectionThreshold = 0.10

// CloneDetector detects code clones using APTED algorithm
type CloneDetector struct {
	// Embed config fields (private to maintain encapsulation)
	cloneDetectorConfig CloneDetectorConfig

	analyzer          *APTEDAnalyzer
	converter         *TreeConverter
	classifier        *CloneClassifier // Multi-dimensional classifier (optional)
	textualAnalyzer   *TextualSimilarityAnalyzer
	syntacticAnalyzer *SyntacticSimilarityAnalyzer
	featureExtractor  *ASTFeatureExtractor // Source for CodeFragment.Features
	fragments         []*CodeFragment
	clonePairs        []*ClonePair
	cloneGroups       []*CloneGroup
}

// buildCloneCostModel creates the APTED cost model for the given configuration.
func buildCloneCostModel(config *CloneDetectorConfig) CostModel {
	switch config.CostModelType {
	case "default":
		return NewDefaultCostModel()
	case "python":
		// Use boilerplate-aware cost model if enabled
		return NewPythonCostModelWithBoilerplateConfig(
			config.IgnoreLiterals,
			config.IgnoreIdentifiers,
			config.ReduceBoilerplateSimilarity,
			config.BoilerplateMultiplier,
		)
	case "weighted":
		baseCostModel := NewPythonCostModelWithBoilerplateConfig(
			config.IgnoreLiterals,
			config.IgnoreIdentifiers,
			config.ReduceBoilerplateSimilarity,
			config.BoilerplateMultiplier,
		)
		return NewWeightedCostModel(1.0, 1.0, 0.8, baseCostModel)
	default:
		return NewPythonCostModel()
	}
}

// buildCloneClassifier creates the multi-dimensional classifier, or nil if disabled.
func buildCloneClassifier(config *CloneDetectorConfig) *CloneClassifier {
	if !config.EnableMultiDimensionalAnalysis {
		return nil
	}
	return NewCloneClassifier(&CloneClassifierConfig{
		Type1Threshold:         config.Type1Threshold,
		Type2Threshold:         config.Type2Threshold,
		Type3Threshold:         config.Type3Threshold,
		Type4Threshold:         config.Type4Threshold,
		EnableTextualAnalysis:  config.EnableTextualAnalysis,
		EnableSemanticAnalysis: config.EnableSemanticAnalysis,
		EnableDFAAnalysis:      config.EnableDFAAnalysis,
	})
}

// NewCloneDetector creates a new clone detector with the given configuration
func NewCloneDetector(config *CloneDetectorConfig) *CloneDetector {
	// If DFA is enabled, automatically enable multi-dimensional analysis and semantic analysis
	if config.EnableDFAAnalysis {
		config.EnableMultiDimensionalAnalysis = true
		config.EnableSemanticAnalysis = true
	}

	return &CloneDetector{
		cloneDetectorConfig: *config,
		analyzer:            NewAPTEDAnalyzer(buildCloneCostModel(config)),
		converter:           NewTreeConverterWithConfig(config.SkipDocstrings),
		classifier:          buildCloneClassifier(config),
		textualAnalyzer:     NewTextualSimilarityAnalyzer(),
		syntacticAnalyzer:   NewSyntacticSimilarityAnalyzer(),
		featureExtractor:    NewASTFeatureExtractor(),
		fragments:           []*CodeFragment{},
		clonePairs:          []*ClonePair{},
		cloneGroups:         []*CloneGroup{},
	}
}

// newWorkerDetector returns a shallow copy of cd with private instances of the
// stateful analyzers (the APTED analyzer's scratch buffers, the classifier's
// internal analyzers, and the syntactic analyzer's tree converter), so that
// fragment comparisons can run concurrently across goroutines. Immutable state
// (config, fragments, textual analyzer, feature extractor) is shared.
func (cd *CloneDetector) newWorkerDetector() *CloneDetector {
	w := *cd
	w.analyzer = NewAPTEDAnalyzer(buildCloneCostModel(&cd.cloneDetectorConfig))
	w.classifier = buildCloneClassifier(&cd.cloneDetectorConfig)
	w.syntacticAnalyzer = NewSyntacticSimilarityAnalyzer()
	return &w
}

// effectiveWorkers returns how many goroutines to use for n independent work
// items, honoring MaxGoroutines (0 means use all available CPUs).
func (cd *CloneDetector) effectiveWorkers(n int) int {
	workers := cd.cloneDetectorConfig.MaxGoroutines
	if workers <= 0 {
		workers = runtime.GOMAXPROCS(0)
	}
	if workers > n {
		workers = n
	}
	if workers < 1 {
		workers = 1
	}
	return workers
}

// runParallelIndexed distributes indices [0, n) across workers goroutines.
// Each goroutine receives a private worker detector plus its worker slot, so
// callers can keep per-worker state without locking. With workers <= 1 the
// indices run inline on cd itself.
func (cd *CloneDetector) runParallelIndexed(ctx context.Context, workers, n int, fn func(wd *CloneDetector, worker, index int)) {
	if workers <= 1 {
		for i := 0; i < n; i++ {
			if isCancelled(ctx) {
				return
			}
			fn(cd, 0, i)
		}
		return
	}

	var next atomic.Int64
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wd := cd.newWorkerDetector()
			for {
				i := int(next.Add(1)) - 1
				if i >= n || isCancelled(ctx) {
					return
				}
				fn(wd, w, i)
			}
		}()
	}
	wg.Wait()
}

// SetUseLSH enables or disables LSH acceleration for clone detection
func (cd *CloneDetector) SetUseLSH(enabled bool) {
	cd.cloneDetectorConfig.UseLSH = enabled
}

// SetBatchSizeLarge sets the batch size for normal projects (used in testing)
func (cd *CloneDetector) SetBatchSizeLarge(size int) {
	cd.cloneDetectorConfig.BatchSizeLarge = size
}

// ExtractFragments extracts code fragments from AST nodes
func (cd *CloneDetector) ExtractFragments(astNodes []*parser.Node, filePath string) []*CodeFragment {
	var fragments []*CodeFragment

	for _, node := range astNodes {
		cd.extractFragmentsRecursive(node, filePath, &fragments)
	}

	return fragments
}

// ExtractFragmentsWithSource extracts code fragments from AST nodes with source content.
// Source content is needed for Type-1 clone classification and optional report output.
func (cd *CloneDetector) ExtractFragmentsWithSource(astNodes []*parser.Node, filePath string, sourceCode []byte) []*CodeFragment {
	var fragments []*CodeFragment
	lines := splitLines(sourceCode)

	for _, node := range astNodes {
		cd.extractFragmentsRecursiveWithSource(node, filePath, lines, &fragments)
	}

	return fragments
}

// extractFragmentsRecursiveWithSource recursively extracts fragments with source content
func (cd *CloneDetector) extractFragmentsRecursiveWithSource(node *parser.Node, filePath string, lines [][]byte, fragments *[]*CodeFragment) {
	if node == nil {
		return
	}

	// Check if this node should be considered as a fragment
	if cd.isFragmentCandidate(node) {
		location := &CodeLocation{
			FilePath:  filePath,
			StartLine: node.Location.StartLine,
			EndLine:   node.Location.EndLine,
			StartCol:  node.Location.StartCol,
			EndCol:    node.Location.EndCol,
		}

		// Extract content from source code if textual analysis is enabled
		content := ""
		if len(lines) > 0 {
			content = cd.extractSourceContent(lines, &node.Location)
		}

		fragment := NewCodeFragment(location, node, content)

		// Filter fragments based on configuration
		if cd.shouldIncludeFragment(fragment) {
			*fragments = append(*fragments, fragment)
		}
	}

	// Recursively process children
	for _, child := range parser.OrderedChildren(node, nil) {
		cd.extractFragmentsRecursiveWithSource(child, filePath, lines, fragments)
	}
}

// extractSourceContent extracts source code content for a given location
func (cd *CloneDetector) extractSourceContent(lines [][]byte, loc *parser.Location) string {
	if loc == nil || len(lines) == 0 {
		return ""
	}

	if loc.StartLine < 1 || loc.EndLine > len(lines) {
		return ""
	}

	// Extract lines from StartLine to EndLine (1-indexed)
	var result []byte
	for i := loc.StartLine - 1; i < loc.EndLine && i < len(lines); i++ {
		result = append(result, lines[i]...)
		if i < loc.EndLine-1 {
			result = append(result, '\n')
		}
	}

	return string(result)
}

// splitLines splits source code into lines
func splitLines(sourceCode []byte) [][]byte {
	if len(sourceCode) == 0 {
		return nil
	}

	lines := make([][]byte, 0, bytes.Count(sourceCode, []byte{'\n'})+1)
	start := 0
	for idx, b := range sourceCode {
		if b == '\n' {
			lines = append(lines, sourceCode[start:idx])
			start = idx + 1
		}
	}
	if start < len(sourceCode) {
		lines = append(lines, sourceCode[start:])
	}
	return lines
}

// extractFragmentsRecursive recursively extracts fragments from AST
func (cd *CloneDetector) extractFragmentsRecursive(node *parser.Node, filePath string, fragments *[]*CodeFragment) {
	if node == nil {
		return
	}

	// Check if this node should be considered as a fragment
	if cd.isFragmentCandidate(node) {
		location := &CodeLocation{
			FilePath:  filePath,
			StartLine: node.Location.StartLine,
			EndLine:   node.Location.EndLine,
			StartCol:  node.Location.StartCol,
			EndCol:    node.Location.EndCol,
		}

		fragment := NewCodeFragment(location, node, "")

		// Filter fragments based on configuration
		if cd.shouldIncludeFragment(fragment) {
			*fragments = append(*fragments, fragment)
		}
	}

	// Recursively process children
	for _, child := range parser.OrderedChildren(node, nil) {
		cd.extractFragmentsRecursive(child, filePath, fragments)
	}
}

// isFragmentCandidate checks if a node should be considered as a fragment candidate
func (cd *CloneDetector) isFragmentCandidate(node *parser.Node) bool {
	switch node.Type {
	// Consider functions, classes, and compound statements as fragment candidates.
	case
		parser.NodeFunctionDef,
		parser.NodeAsyncFunctionDef,
		parser.NodeClassDef,
		parser.NodeFor,
		parser.NodeAsyncFor,
		parser.NodeWhile,
		parser.NodeIf,
		parser.NodeTry,
		parser.NodeWith,
		parser.NodeAsyncWith:
		return true
	}

	return false
}

// shouldIncludeFragment determines if a fragment should be included in analysis
func (cd *CloneDetector) shouldIncludeFragment(fragment *CodeFragment) bool {
	// Check minimum size requirements
	if fragment.Size < cd.cloneDetectorConfig.MinNodes {
		return false
	}

	if fragment.LineCount < cd.cloneDetectorConfig.MinLines {
		return false
	}

	return true
}

// isCancelled checks if the context is cancelled
func isCancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// DetectClones detects clones in the given code fragments
func (cd *CloneDetector) DetectClones(fragments []*CodeFragment) ([]*ClonePair, []*CloneGroup) {
	return cd.DetectClonesWithContext(context.Background(), fragments)
}

// DetectClonesWithContext detects clones with context support for cancellation
func (cd *CloneDetector) DetectClonesWithContext(ctx context.Context, fragments []*CodeFragment) ([]*ClonePair, []*CloneGroup) {
	cd.fragments = fragments
	cd.clonePairs = []*ClonePair{}
	cd.cloneGroups = []*CloneGroup{}

	// Check for cancellation before starting
	if isCancelled(ctx) {
		return cd.clonePairs, cd.cloneGroups
	}

	// Convert AST fragments to tree nodes
	cd.prepareFragments()

	// Check for cancellation after preparation
	if isCancelled(ctx) {
		return cd.clonePairs, cd.cloneGroups
	}

	// Detect clone pairs with context
	cd.detectClonePairsWithContext(ctx)

	// Check for cancellation before grouping
	if isCancelled(ctx) {
		return cd.clonePairs, cd.cloneGroups
	}

	// Group related clones using configured strategy
	// Clamp threshold to [0,1]
	thr := cd.cloneDetectorConfig.GroupingThreshold
	if thr < 0.0 {
		thr = 0.0
	} else if thr > 1.0 {
		thr = 1.0
	}
	k := cd.cloneDetectorConfig.KCoreK
	if k < 2 {
		k = 2
	}
	groupingConfig := GroupingConfig{
		Mode:           cd.cloneDetectorConfig.GroupingMode,
		Threshold:      thr,
		KCoreK:         k,
		Type1Threshold: cd.cloneDetectorConfig.Type1Threshold,
		Type2Threshold: cd.cloneDetectorConfig.Type2Threshold,
		Type3Threshold: cd.cloneDetectorConfig.Type3Threshold,
		Type4Threshold: cd.cloneDetectorConfig.Type4Threshold,
	}
	strategy := CreateGroupingStrategy(groupingConfig)
	cd.groupClonesWithStrategy(strategy)

	return cd.clonePairs, cd.cloneGroups
}

// DetectClonesWithLSH runs a two-stage pipeline using LSH for candidate generation,
// followed by APTED verification on candidates only. Falls back to exhaustive if misconfigured.
func (cd *CloneDetector) DetectClonesWithLSH(ctx context.Context, fragments []*CodeFragment) ([]*ClonePair, []*CloneGroup) {
	// If not enabled, delegate to standard path
	if cd == nil || !cd.cloneDetectorConfig.UseLSH {
		return cd.DetectClonesWithContext(ctx, fragments)
	}

	cd.fragments = fragments
	cd.clonePairs = []*ClonePair{}
	cd.cloneGroups = []*CloneGroup{}

	if isCancelled(ctx) {
		return cd.clonePairs, cd.cloneGroups
	}

	// Prepare TreeNodes for APTED and feature extraction
	cd.prepareFragments()

	if isCancelled(ctx) {
		return cd.clonePairs, cd.cloneGroups
	}

	// Stage 1: MinHash signatures from prepared clone features
	hasher := NewMinHasher(cd.cloneDetectorConfig.LSHMinHashCount)

	type fragRec struct {
		idx int
		sig *MinHashSignature
	}
	records := make([]fragRec, 0, len(cd.fragments))
	sigByIndex := make(map[int]*MinHashSignature, len(cd.fragments))
	for i, f := range cd.fragments {
		if f == nil || f.TreeNode == nil {
			continue
		}
		sig := hasher.ComputeSignature(f.Features)
		records = append(records, fragRec{idx: i, sig: sig})
		sigByIndex[i] = sig
	}

	// Edge case: if signatures cannot be built, fallback
	if len(records) <= 1 {
		return cd.DetectClonesWithContext(ctx, fragments)
	}

	// Stage 2: LSH indexing
	lsh := NewLSHIndex(cd.cloneDetectorConfig.LSHBands, cd.cloneDetectorConfig.LSHRows).
		WithMaxCandidates(cd.cloneDetectorConfig.LSHMaxCandidates)
	for _, r := range records {
		_ = lsh.AddFragment(r.idx, r.sig)
	}
	_ = lsh.BuildIndex()

	// Stage 3: Candidate generation + APTED verification
	// Use MinHash similarity to filter before expensive APTED
	minhashThreshold := cd.cloneDetectorConfig.LSHSimilarityThreshold
	if minhashThreshold < 0 {
		minhashThreshold = 0
	} else if minhashThreshold > 1 {
		minhashThreshold = 1
	}

	// Collect deduplicated candidate pairs that survive the cheap pre-filters,
	// then run the expensive APTED verification in parallel.
	type candidatePair struct{ a, b int }
	seenPairs := make(map[[2]int]struct{})
	var candidates []candidatePair
	for _, r := range records {
		if isCancelled(ctx) {
			break
		}
		cands := lsh.FindCandidates(r.sig)
		for _, j := range cands {
			i := r.idx
			if j == i || j < 0 || i < 0 {
				continue
			}
			// Deduplicate unordered pair (i<j)
			a, b := i, j
			if a > b {
				a, b = b, a
			}
			key := [2]int{a, b}
			if _, ok := seenPairs[key]; ok {
				continue
			}
			seenPairs[key] = struct{}{}

			f1 := cd.fragments[a]
			f2 := cd.fragments[b]
			if cd.isOverlappingLocation(f1.Location, f2.Location) {
				continue
			}

			// MinHash similarity pre-filter
			est := hasher.EstimateJaccardSimilarity(sigByIndex[a], sigByIndex[b])
			if est < minhashThreshold {
				continue
			}

			if f1.TreeNode == nil || f2.TreeNode == nil {
				continue
			}
			candidates = append(candidates, candidatePair{a: a, b: b})
		}
	}

	// APTED verification on surviving candidates
	verified := make([]*ClonePair, len(candidates))
	workers := cd.effectiveWorkers(len(candidates))
	cd.runParallelIndexed(ctx, workers, len(candidates), func(wd *CloneDetector, _, k int) {
		c := candidates[k]
		pair := wd.compareFragments(cd.fragments[c.a], cd.fragments[c.b])
		if pair != nil && wd.isSignificantClone(pair) {
			verified[k] = pair
		}
	})
	for _, pair := range verified {
		if pair != nil {
			cd.clonePairs = append(cd.clonePairs, pair)
		}
	}

	// Finalize results
	cd.limitAndSortClonePairs(cd.cloneDetectorConfig.MaxClonePairs)

	// Grouping
	thr := cd.cloneDetectorConfig.GroupingThreshold
	if thr < 0.0 {
		thr = 0.0
	} else if thr > 1.0 {
		thr = 1.0
	}
	k := cd.cloneDetectorConfig.KCoreK
	if k < 2 {
		k = 2
	}
	groupingConfig := GroupingConfig{
		Mode:           cd.cloneDetectorConfig.GroupingMode,
		Threshold:      thr,
		KCoreK:         k,
		Type1Threshold: cd.cloneDetectorConfig.Type1Threshold,
		Type2Threshold: cd.cloneDetectorConfig.Type2Threshold,
		Type3Threshold: cd.cloneDetectorConfig.Type3Threshold,
		Type4Threshold: cd.cloneDetectorConfig.Type4Threshold,
	}
	strategy := CreateGroupingStrategy(groupingConfig)
	cd.groupClonesWithStrategy(strategy)

	return cd.clonePairs, cd.cloneGroups
}

// prepareFragments converts AST fragments to tree nodes and populates clone features.
func (cd *CloneDetector) prepareFragments() {
	for _, fragment := range cd.fragments {
		if fragment == nil {
			continue
		}
		if fragment.TreeNode == nil && fragment.ASTNode != nil {
			fragment.TreeNode = cd.converter.ConvertAST(fragment.ASTNode)
		}
		if fragment.TreeNode == nil {
			continue
		}
		PrepareTreeForAPTED(fragment.TreeNode)
		features, _ := cd.featureExtractor.ExtractFeatures(fragment.TreeNode)
		fragment.Features = features
	}
}

// detectClonePairsWithContext detects pairs with context support
func (cd *CloneDetector) detectClonePairsWithContext(ctx context.Context) {
	n := len(cd.fragments)

	// Early return for small datasets
	if n <= 1 {
		return
	}

	maxPairs := cd.cloneDetectorConfig.MaxClonePairs
	if maxPairs <= 0 {
		maxPairs = 10000
	}

	cd.detectClonePairsParallel(ctx, maxPairs)

	// Sort and limit final results
	cd.limitAndSortClonePairs(maxPairs)
}

// clonePairMinHeap is a min-heap on Similarity: the root is the worst retained
// pair, so a better candidate can replace it in O(log n). Used to keep the
// best maxPairs candidates per worker with bounded memory.
type clonePairMinHeap []*ClonePair

func (h clonePairMinHeap) Len() int           { return len(h) }
func (h clonePairMinHeap) Less(i, j int) bool { return h[i].Similarity < h[j].Similarity }
func (h clonePairMinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *clonePairMinHeap) Push(x any)        { *h = append(*h, x.(*ClonePair)) }
func (h *clonePairMinHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// detectClonePairsParallel enumerates all unordered fragment pairs row by row,
// fanning rows out across workers. Each worker keeps a bounded min-heap of its
// best pairs, so memory stays O(workers * maxPairs) regardless of input size.
// The union of per-worker top-maxPairs always contains the global top-maxPairs.
func (cd *CloneDetector) detectClonePairsParallel(ctx context.Context, maxPairs int) {
	n := len(cd.fragments)
	workers := cd.effectiveWorkers(n - 1)
	heaps := make([]clonePairMinHeap, workers)

	cd.runParallelIndexed(ctx, workers, n, func(wd *CloneDetector, worker, i int) {
		h := &heaps[worker]
		for j := i + 1; j < n; j++ {
			// Once the heap is full, the worst retained similarity becomes a
			// pruning floor for this worker.
			floor := 0.0
			if h.Len() >= maxPairs {
				floor = (*h)[0].Similarity
			}
			pair := wd.tryCreateClonePair(i, j, floor)
			if pair == nil {
				continue
			}
			if h.Len() < maxPairs {
				heap.Push(h, pair)
			} else if pair.Similarity > (*h)[0].Similarity {
				(*h)[0] = pair
				heap.Fix(h, 0)
			}
		}
	})

	total := 0
	for w := range heaps {
		total += heaps[w].Len()
	}
	merged := make([]*ClonePair, 0, total)
	for w := range heaps {
		merged = append(merged, heaps[w]...)
	}
	cd.clonePairs = merged
}

// shouldCompareFragments performs early filtering to determine if two fragments should be compared
func (cd *CloneDetector) shouldCompareFragments(fragment1, fragment2 *CodeFragment) bool {
	// Early filtering: Skip if size difference is too large (>50%)
	sizeDiff := math.Abs(float64(fragment1.Size - fragment2.Size))
	avgSize := float64(fragment1.Size+fragment2.Size) / 2.0
	if avgSize > 0 && sizeDiff/avgSize > 0.5 {
		return false // Too different in size to be clones
	}

	// Early filtering: Skip if line count difference is too large
	lineDiff := math.Abs(float64(fragment1.LineCount - fragment2.LineCount))
	if lineDiff > float64(fragment1.LineCount)*0.5 && lineDiff > float64(fragment2.LineCount)*0.5 {
		return false // Too different in line count
	}

	return true
}

// compareFragments compares two fragments and returns a clone pair if similar.
// Uses a Jaccard pre-filter on pre-computed features to minimize expensive APTED calls.
func (cd *CloneDetector) compareFragments(fragment1, fragment2 *CodeFragment) *ClonePair {
	if fragment1.TreeNode == nil || fragment2.TreeNode == nil {
		return nil
	}

	if cd.usesSemanticClassifier() {
		return cd.compareFragmentsWithClassifier(fragment1, fragment2)
	}

	// Early filtering check
	if !cd.shouldCompareFragments(fragment1, fragment2) {
		return nil
	}

	// Jaccard pre-filter: reject clear non-clones before expensive APTED/classifier work.
	// Only used for rejection — all non-rejected pairs proceed to APTED-based classification
	// for accurate clone typing and distance computation.
	if len(fragment1.Features) > 0 && len(fragment2.Features) > 0 {
		if jaccardSimilarity(fragment1.Features, fragment2.Features) < jaccardRejectionThreshold {
			return nil
		}
	}

	// Use multi-dimensional classifier if enabled
	if cd.classifier != nil && cd.cloneDetectorConfig.EnableMultiDimensionalAnalysis {
		return cd.compareFragmentsWithClassifier(fragment1, fragment2)
	}

	// Fallback to single-metric classification (backward compatible)
	return cd.compareFragmentsSingleMetric(fragment1, fragment2)
}

func (cd *CloneDetector) usesSemanticClassifier() bool {
	return cd.classifier != nil &&
		cd.cloneDetectorConfig.EnableMultiDimensionalAnalysis &&
		cd.cloneDetectorConfig.EnableSemanticAnalysis
}

// compareFragmentsWithClassifier uses the classifier as a gate and APTED for
// final similarity/distance scoring and clone-type classification.
func (cd *CloneDetector) compareFragmentsWithClassifier(fragment1, fragment2 *CodeFragment) *ClonePair {
	result := cd.classifier.ClassifyClone(fragment1, fragment2)
	if result == nil {
		return nil
	}

	// Always run APTED with the detector's cost model (boilerplate-aware) so that
	// Distance is populated and clone type is derived from a consistent metric.
	distance, similarity := cd.analyzer.ComputeDistanceAndSimilarity(fragment1.TreeNode, fragment2.TreeNode)

	if result.CloneType == Type4Clone && result.Analyzer == "semantic" {
		return &ClonePair{
			Fragment1:  fragment1,
			Fragment2:  fragment2,
			Similarity: result.Similarity,
			Distance:   distance,
			CloneType:  result.CloneType,
			Confidence: result.Confidence,
		}
	}

	cloneType, similarity := cd.classifyClonePair(fragment1, fragment2, similarity)
	if cloneType == 0 {
		return nil
	}

	return &ClonePair{
		Fragment1:  fragment1,
		Fragment2:  fragment2,
		Similarity: similarity,
		Distance:   distance,
		CloneType:  cloneType,
		Confidence: result.Confidence,
	}
}

// compareFragmentsSingleMetric uses APTED for clone type classification.
// Jaccard pre-filtering is handled at the compareFragments level.
func (cd *CloneDetector) compareFragmentsSingleMetric(fragment1, fragment2 *CodeFragment) *ClonePair {
	return cd.compareWithAPTED(fragment1, fragment2)
}

// compareWithAPTED uses the APTED algorithm for precise similarity measurement.
func (cd *CloneDetector) compareWithAPTED(fragment1, fragment2 *CodeFragment) *ClonePair {
	distance, similarity := cd.analyzer.ComputeDistanceAndSimilarity(fragment1.TreeNode, fragment2.TreeNode)

	cloneType, similarity := cd.classifyClonePair(fragment1, fragment2, similarity)
	if cloneType == 0 {
		return nil
	}

	confidence := cd.calculateConfidence(fragment1, fragment2, similarity)

	return &ClonePair{
		Fragment1:  fragment1,
		Fragment2:  fragment2,
		Similarity: similarity,
		Distance:   distance,
		CloneType:  cloneType,
		Confidence: confidence,
	}
}

func (cd *CloneDetector) classifyClonePair(fragment1, fragment2 *CodeFragment, similarity float64) (CloneType, float64) {
	if similarity >= cd.cloneDetectorConfig.Type1Threshold && cd.textualAnalyzer.IsExactMatch(fragment1, fragment2) {
		return Type1Clone, similarity
	}

	structuralSimilarity := cd.capNonTextualSimilarity(similarity)
	if structuralSimilarity >= cd.cloneDetectorConfig.Type2Threshold {
		syntacticSimilarity := cd.syntacticAnalyzer.ComputeSimilarity(fragment1, fragment2)
		if syntacticSimilarity >= cd.cloneDetectorConfig.Type2Threshold {
			return Type2Clone, math.Min(structuralSimilarity, syntacticSimilarity)
		}
	}
	if structuralSimilarity >= cd.cloneDetectorConfig.Type3Threshold {
		return Type3Clone, structuralSimilarity
	}
	if structuralSimilarity >= cd.cloneDetectorConfig.Type4Threshold {
		return Type4Clone, structuralSimilarity
	}

	return 0, structuralSimilarity
}

func (cd *CloneDetector) capNonTextualSimilarity(similarity float64) float64 {
	if similarity < cd.cloneDetectorConfig.Type1Threshold {
		return similarity
	}

	capped := math.Nextafter(cd.cloneDetectorConfig.Type1Threshold, 0)
	if capped < cd.cloneDetectorConfig.Type2Threshold {
		return cd.cloneDetectorConfig.Type2Threshold
	}
	return capped
}

// calculateConfidence calculates confidence in clone detection
func (cd *CloneDetector) calculateConfidence(fragment1, fragment2 *CodeFragment, similarity float64) float64 {
	// Base confidence on similarity
	confidence := similarity

	// Increase confidence for larger fragments
	avgSize := float64(fragment1.Size+fragment2.Size) / 2.0
	sizeBonus := math.Min(avgSize/100.0, 0.2) // Up to 20% bonus for large fragments
	confidence += sizeBonus

	// Increase confidence if both fragments have similar complexity
	if fragment1.Complexity > 0 && fragment2.Complexity > 0 {
		complexityRatio := float64(math.Min(float64(fragment1.Complexity), float64(fragment2.Complexity))) /
			float64(math.Max(float64(fragment1.Complexity), float64(fragment2.Complexity)))
		confidence += complexityRatio * 0.1 // Up to 10% bonus for similar complexity
	}

	// Cap confidence at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// isSignificantClone determines if a clone pair is significant enough to report
func (cd *CloneDetector) isSignificantClone(pair *ClonePair) bool {
	// Check minimum similarity threshold
	// Use SimilarityThreshold if set, otherwise fall back to Type4Threshold
	minThreshold := cd.cloneDetectorConfig.SimilarityThreshold
	if minThreshold <= 0 {
		minThreshold = cd.cloneDetectorConfig.Type4Threshold
	}
	if pair.Similarity < minThreshold {
		return false
	}

	// Check maximum distance threshold (0 means no limit)
	if pair.CloneType != Type4Clone && cd.cloneDetectorConfig.MaxEditDistance > 0 && pair.Distance > cd.cloneDetectorConfig.MaxEditDistance {
		return false
	}

	// Additional filtering based on fragment characteristics
	minSize := math.Min(float64(pair.Fragment1.Size), float64(pair.Fragment2.Size))
	return minSize >= float64(cd.cloneDetectorConfig.MinNodes)
}

// groupClonesWithStrategy groups clone pairs using a pluggable strategy.
// This keeps backward compatibility with the existing groupClones method.
//
//nolint:unused // Hook for pluggable grouping; used when strategy is wired via config.
func (cd *CloneDetector) groupClonesWithStrategy(strategy GroupingStrategy) {
	if strategy == nil {
		cd.cloneGroups = []*CloneGroup{}
		return
	}
	dedupeResult := dedupeStrictSubsetGroupMembers(strategy.GroupClones(cd.clonePairs), cd.clonePairs)
	cd.cloneGroups = dedupeResult.groups
	cd.clonePairs = filterClonePairsWithSuppressedMembers(cd.clonePairs, dedupeResult.suppressed)
}

// isSameLocation checks if two locations refer to the same code
func (cd *CloneDetector) isSameLocation(loc1, loc2 *CodeLocation) bool {
	return loc1.FilePath == loc2.FilePath &&
		loc1.StartLine == loc2.StartLine &&
		loc1.EndLine == loc2.EndLine
}

// isOverlappingLocation checks if two locations from the same file overlap
// (one contains or partially contains the other)
func (cd *CloneDetector) isOverlappingLocation(loc1, loc2 *CodeLocation) bool {
	if loc1.FilePath != loc2.FilePath {
		return false
	}
	// Check if ranges overlap: NOT (loc1 ends before loc2 starts OR loc2 ends before loc1 starts)
	return !(loc1.EndLine < loc2.StartLine || loc2.EndLine < loc1.StartLine)
}

// GetStatistics returns clone detection statistics
func (cd *CloneDetector) GetStatistics() map[string]interface{} {
	stats := make(map[string]interface{})

	stats["total_fragments"] = len(cd.fragments)
	stats["total_clone_pairs"] = len(cd.clonePairs)
	stats["total_clone_groups"] = len(cd.cloneGroups)

	// Count by clone type
	typeCounts := make(map[string]int)
	for _, pair := range cd.clonePairs {
		typeCounts[pair.CloneType.String()]++
	}
	stats["clone_types"] = typeCounts

	// Average similarity
	if len(cd.clonePairs) > 0 {
		totalSim := 0.0
		for _, pair := range cd.clonePairs {
			totalSim += pair.Similarity
		}
		stats["average_similarity"] = totalSim / float64(len(cd.clonePairs))
	}

	return stats
}

// tryCreateClonePair attempts to create a clone pair if it meets similarity threshold
func (cd *CloneDetector) tryCreateClonePair(i, j int, minSimilarity float64) *ClonePair {
	fragment1 := cd.fragments[i]
	fragment2 := cd.fragments[j]

	// Skip if both fragments are from the same location or overlap in the same file
	if cd.isOverlappingLocation(fragment1.Location, fragment2.Location) {
		return nil
	}

	// Quick similarity check before expensive computation
	if fragment1.TreeNode == nil || fragment2.TreeNode == nil {
		return nil
	}

	// Full similarity computation (compareFragments already calls shouldCompareFragments)
	pair := cd.compareFragments(fragment1, fragment2)
	if pair != nil && cd.isSignificantClone(pair) && pair.Similarity >= minSimilarity {
		return pair
	}
	return nil
}

// limitAndSortClonePairs ensures final results are sorted and limited
func (cd *CloneDetector) limitAndSortClonePairs(maxPairs int) {
	// Sort clone pairs by similarity (descending)
	sort.Slice(cd.clonePairs, func(i, j int) bool {
		return cd.clonePairs[i].Similarity > cd.clonePairs[j].Similarity
	})

	// Limit the number of pairs to prevent memory issues
	if len(cd.clonePairs) > maxPairs {
		cd.clonePairs = cd.clonePairs[:maxPairs]
	}
}

// Helper functions
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
