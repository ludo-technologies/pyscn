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

	coreapted "github.com/ludo-technologies/polyscan/core/apted"
	coreclone "github.com/ludo-technologies/polyscan/core/clone"
	coredomain "github.com/ludo-technologies/polyscan/core/domain"
	corelsh "github.com/ludo-technologies/polyscan/core/lsh"
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
	TreeNode   *coreapted.TreeNode
	Content    string   // Original source code content
	Hash       string   // FNV-64a hex hash of Type-1 normalized content; "" when no source content
	Size       int      // Number of AST nodes
	LineCount  int      // Number of source lines
	Complexity int      // Cyclomatic complexity (if applicable)
	Features   []string // Detector-populated clone feature cache for this fragment's tree

	// belowSizeGate marks a fragment that failed MinLines/MinNodes but was kept
	// as an identical-body candidate. It survives extraction only if
	// RetainIdenticalUndersizedFragments finds it an identical twin.
	belowSizeGate bool

	// id is a detector-assigned identifier used for core/clone grouping.
	id int
	// core caches the core/clone projection of this fragment, populated by
	// prepareFragments so per-pair comparisons don't re-convert the tree.
	core *coreclone.CodeFragment
}

// ItemID returns the fragment's unique ID for core/clone grouping.
func (f *CodeFragment) ItemID() int { return f.id }

// ItemLocation returns the fragment's source location for core/clone grouping.
func (f *CodeFragment) ItemLocation() coreclone.ItemLocation {
	if f == nil || f.Location == nil {
		return coreclone.ItemLocation{}
	}
	return coreclone.ItemLocation{
		FilePath:  f.Location.FilePath,
		StartLine: f.Location.StartLine,
		EndLine:   f.Location.EndLine,
		StartCol:  f.Location.StartCol,
		EndCol:    f.Location.EndCol,
	}
}

// coreFragment returns the core/clone projection of the fragment, using the
// cached conversion when available.
func (f *CodeFragment) coreFragment() *coreclone.CodeFragment {
	if f == nil {
		return nil
	}
	if f.core != nil {
		return f.core
	}
	return toCoreFragment(f, f.id)
}

// toCoreFragment converts a fragment to its core/clone representation.
func toCoreFragment(fragment *CodeFragment, id int) *coreclone.CodeFragment {
	if fragment == nil {
		return nil
	}
	tree := fragment.TreeNode
	if tree == nil && fragment.ASTNode != nil {
		tree = NewTreeConverter().ConvertAST(fragment.ASTNode)
		coreapted.PrepareTreeForAPTED(tree)
	}
	result := &coreclone.CodeFragment{
		ID:         id,
		Content:    fragment.Content,
		Hash:       fragment.Hash,
		ASTNode:    tree,
		NodeCount:  fragment.Size,
		LineCount:  fragment.LineCount,
		Complexity: fragment.Complexity,
		Features:   fragment.Features,
	}
	if fragment.Location != nil {
		result.FilePath = fragment.Location.FilePath
		result.StartLine = fragment.Location.StartLine
		result.EndLine = fragment.Location.EndLine
		result.StartCol = fragment.Location.StartCol
		result.EndCol = fragment.Location.EndCol
	}
	return result
}

// NewCodeFragment creates a new code fragment
func NewCodeFragment(location *CodeLocation, astNode *parser.Node, content string) *CodeFragment {
	return &CodeFragment{
		Location:  location,
		ASTNode:   astNode,
		Content:   content,
		Hash:      fragmentHashNormalizer.HashFragmentContent(content),
		Size:      calculateASTSize(astNode),
		LineCount: location.EndLine - location.StartLine + 1,
	}
}

// newFragment creates a fragment whose size and line count honor SkipDocstrings,
// so docstrings neither pad MinLines/MinNodes nor appear in the compared tree.
func (cd *CloneDetector) newFragment(location *CodeLocation, astNode *parser.Node, content string) *CodeFragment {
	fragment := NewCodeFragment(location, astNode, content)
	if cd.cloneDetectorConfig.SkipDocstrings {
		fragment.Size = cd.converter.CountNodes(astNode)
		fragment.LineCount -= cd.converter.DocstringLines(astNode)
	}
	return fragment
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

// CloneDetectionStatistics provides statistics computed during clone detection.
// It tracks only the data the detector itself can derive from the fragment and
// pair/group collections. Callers (e.g. service/clone_service.go) augment it
// with file/line/node counts gathered while reading sources.
type CloneDetectionStatistics struct {
	TotalFragments    int
	TotalClones       int
	TotalClonePairs   int
	TotalCloneGroups  int
	ClonesByType      map[string]int
	AverageSimilarity float64
}

// CloneDetectionResult bundles the detected clone pairs and groups with the
// statistics derived from them. Wrapping the output prevents accidental use of
// raw fragment counts where detected clone counts are required.
type CloneDetectionResult struct {
	Pairs      []*ClonePair
	Groups     []*CloneGroup
	Statistics *CloneDetectionStatistics
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
	GroupingMode      GroupingMode // Default: GroupingModeConnected
	GroupingThreshold float64      // Default: Type3Threshold
	KCoreK            int          // Default: 2

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
		MinLines:          domain.DefaultCloneMinLines,
		MinNodes:          domain.DefaultCloneMinNodes,
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
		GroupingMode:      GroupingMode(domain.DefaultCloneGroupMode),
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

// pythonClonePatternNames are the Python AST constructs surfaced as structural
// pattern features for clone feature extraction (core's extractor is
// language-neutral and emits pattern features only for configured names).
var pythonClonePatternNames = []string{
	"If", "For", "While", "Try", "With",
	"FunctionDef", "ClassDef", "Return", "Assign", "Call", "Attribute", "Compare",
}

// pythonLiteralLikeNames are the Python AST node labels that carry identifier
// or literal payloads; their label features are suppressed when literals are
// excluded so renames and literal changes do not perturb the feature set.
var pythonLiteralLikeNames = []string{"Name", "Constant", "Arg", "Keyword"}

// newPythonCloneFeatureExtractor returns a core feature extractor configured
// with Python pattern and literal-like label names.
func newPythonCloneFeatureExtractor() *coreclone.ASTFeatureExtractor {
	return coreclone.NewASTFeatureExtractor().
		WithPatterns(pythonClonePatternNames).
		WithLiteralNames(pythonLiteralLikeNames)
}

// CloneDetector detects code clones using APTED algorithm
type CloneDetector struct {
	// Embed config fields (private to maintain encapsulation)
	cloneDetectorConfig CloneDetectorConfig

	analyzer         *coreapted.APTEDAnalyzer
	converter        *TreeConverter
	classifier       *CloneClassifier // Multi-dimensional classifier (optional)
	textualAnalyzer  *coreclone.TextualSimilarityAnalyzer
	featureExtractor *coreclone.ASTFeatureExtractor // Source for CodeFragment.Features
	pairClassifier   *coreclone.PairClassifier
	fragments        []*CodeFragment
	clonePairs       []*ClonePair
	cloneGroups      []*CloneGroup
}

// buildCloneCostModel creates the APTED cost model for the given configuration.
func buildCloneCostModel(config *CloneDetectorConfig) coreapted.CostModel {
	switch config.CostModelType {
	case "default":
		return coreapted.NewDefaultCostModel()
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
		return coreapted.NewWeightedCostModel(1.0, 1.0, 0.8, baseCostModel)
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

	textualAnalyzer := coreclone.NewTextualSimilarityAnalyzer(removePythonComments)
	pairClassifier := coreclone.NewPairClassifier(coreclone.ClassifierConfig{
		Type1Threshold: config.Type1Threshold, Type2Threshold: config.Type2Threshold,
		Type3Threshold: config.Type3Threshold, Type4Threshold: config.Type4Threshold,
		EnableType1: true, EnableType2: true, EnableType3: true, EnableType4: true,
		JaccardPreFilterThreshold: 0.10,
	}, textualAnalyzer, coreclone.NewSyntacticSimilarityAnalyzerWithExtractor(
		newPythonCloneFeatureExtractor().WithOptions(3, 4, true, false)))

	return &CloneDetector{
		cloneDetectorConfig: *config,
		analyzer:            newAPTEDAnalyzer(buildCloneCostModel(config)),
		converter:           NewTreeConverterWithConfig(config.SkipDocstrings),
		classifier:          buildCloneClassifier(config),
		textualAnalyzer:     textualAnalyzer,
		featureExtractor:    newPythonCloneFeatureExtractor(),
		pairClassifier:      pairClassifier,
		fragments:           []*CodeFragment{},
		clonePairs:          []*ClonePair{},
		cloneGroups:         []*CloneGroup{},
	}
}

// newWorkerDetector returns a shallow copy of cd with private instances of the
// stateful analyzers (the APTED analyzer's scratch buffers and the classifier's
// internal analyzers), so that fragment comparisons can run concurrently across
// goroutines. Immutable state (config, fragments, the core textual/syntactic
// analyzers, pair classifier, and feature extractor) is shared.
func (cd *CloneDetector) newWorkerDetector() *CloneDetector {
	w := *cd
	w.analyzer = newAPTEDAnalyzer(buildCloneCostModel(&cd.cloneDetectorConfig))
	w.classifier = buildCloneClassifier(&cd.cloneDetectorConfig)
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

		cd.collectFragment(cd.newFragment(location, node, content), fragments)
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

		cd.collectFragment(cd.newFragment(location, node, ""), fragments)
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

// collectFragment appends a fragment that passes the MinLines/MinNodes gate.
// A fragment that fails the gate is still kept when it is an identical-body
// candidate, because a byte-identical twin elsewhere in the project is real
// duplication no matter how short the body is. Such a fragment is provisional:
// RetainIdenticalUndersizedFragments drops it again once every file has been
// extracted and no twin has turned up.
func (cd *CloneDetector) collectFragment(fragment *CodeFragment, fragments *[]*CodeFragment) {
	switch {
	case cd.shouldIncludeFragment(fragment):
	case isIdenticalBodyCandidate(fragment):
		fragment.belowSizeGate = true
	default:
		return
	}
	*fragments = append(*fragments, fragment)
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

// identicalBodyMinStatements is how many statements an undersized function body
// must have before an identical twin counts as duplication. A single-statement
// override (pass, return self.x, raise NotImplementedError) is protocol
// boilerplate that repeats across unrelated classes, not copy-paste.
const identicalBodyMinStatements = 2

// isIdenticalBodyCandidate reports whether an undersized fragment is a complete
// function body worth keeping until its twin is known. Requiring a whole
// function excludes partial blocks, whose identical text is rarely actionable
// on its own, and the fragment needs source content because identity is decided
// on the Type-1 normalized hash of that content.
func isIdenticalBodyCandidate(fragment *CodeFragment) bool {
	if fragment.Hash == "" || fragment.ASTNode == nil {
		return false
	}

	switch fragment.ASTNode.Type {
	case parser.NodeFunctionDef, parser.NodeAsyncFunctionDef:
	default:
		return false
	}

	return countBodyStatements(fragment.ASTNode) >= identicalBodyMinStatements
}

// countBodyStatements counts the statements in a body, ignoring a leading
// docstring so that documenting a one-line override does not make it look like
// copy-pasted logic.
func countBodyStatements(node *parser.Node) int {
	count := 0
	for i, stmt := range node.Body {
		if i == 0 && isStringConstantStatement(stmt) {
			continue
		}
		count++
	}
	return count
}

// RetainIdenticalUndersizedFragments finalizes a fragment set gathered across
// files. Fragments held past the size gate during extraction survive only when
// an identical one was found elsewhere; the rest are discarded, leaving the
// MinLines/MinNodes contract intact for everything that is not a duplicate.
func RetainIdenticalUndersizedFragments(fragments []*CodeFragment) []*CodeFragment {
	occurrences := make(map[string]int)
	for _, fragment := range fragments {
		if fragment != nil && fragment.belowSizeGate {
			occurrences[fragment.Hash]++
		}
	}
	if len(occurrences) == 0 {
		return fragments
	}

	retained := make([]*CodeFragment, 0, len(fragments))
	for _, fragment := range fragments {
		if fragment != nil && fragment.belowSizeGate && occurrences[fragment.Hash] < 2 {
			continue
		}
		retained = append(retained, fragment)
	}
	return retained
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
func (cd *CloneDetector) DetectClones(fragments []*CodeFragment) *CloneDetectionResult {
	return cd.DetectClonesWithContext(context.Background(), fragments)
}

// DetectClonesWithContext detects clones with context support for cancellation
func (cd *CloneDetector) DetectClonesWithContext(ctx context.Context, fragments []*CodeFragment) *CloneDetectionResult {
	cd.fragments = fragments
	cd.clonePairs = []*ClonePair{}
	cd.cloneGroups = []*CloneGroup{}

	// Check for cancellation before starting
	if isCancelled(ctx) {
		return cd.buildCloneDetectionResult()
	}

	// Convert AST fragments to tree nodes
	cd.prepareFragments()

	// Check for cancellation after preparation
	if isCancelled(ctx) {
		return cd.buildCloneDetectionResult()
	}

	// Detect clone pairs with context
	cd.detectClonePairsWithContext(ctx)

	// Check for cancellation before grouping
	if isCancelled(ctx) {
		return cd.buildCloneDetectionResult()
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
	cd.groupClones(thr, k)

	return cd.buildCloneDetectionResult()
}

// DetectClonesWithLSH runs a two-stage pipeline using LSH for candidate generation,
// followed by APTED verification on candidates only. Falls back to exhaustive if misconfigured.
func (cd *CloneDetector) DetectClonesWithLSH(ctx context.Context, fragments []*CodeFragment) *CloneDetectionResult {
	// If not enabled, delegate to standard path
	if cd == nil || !cd.cloneDetectorConfig.UseLSH {
		return cd.DetectClonesWithContext(ctx, fragments)
	}

	cd.fragments = fragments
	cd.clonePairs = []*ClonePair{}
	cd.cloneGroups = []*CloneGroup{}

	if isCancelled(ctx) {
		return cd.buildCloneDetectionResult()
	}

	// Prepare TreeNodes for APTED and feature extraction
	cd.prepareFragments()

	if isCancelled(ctx) {
		return cd.buildCloneDetectionResult()
	}

	// Stage 1: MinHash signatures from prepared clone features
	hasher := corelsh.NewMinHasher(cd.cloneDetectorConfig.LSHMinHashCount)

	type fragRec struct {
		idx int
		sig *corelsh.MinHashSignature
	}
	records := make([]fragRec, 0, len(cd.fragments))
	sigByIndex := make(map[int]*corelsh.MinHashSignature, len(cd.fragments))
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
	lsh := newLSHCandidateIndex(cd.cloneDetectorConfig.LSHBands, cd.cloneDetectorConfig.LSHRows, cd.cloneDetectorConfig.LSHMaxCandidates)
	for _, r := range records {
		_ = lsh.AddFragment(r.idx, r.sig)
	}

	// Stage 3: Candidate generation + APTED verification
	// Use MinHash similarity to filter before expensive APTED
	minhashThreshold := cd.cloneDetectorConfig.LSHSimilarityThreshold
	if minhashThreshold < 0 {
		minhashThreshold = 0
	} else if minhashThreshold > 1 {
		minhashThreshold = 1
	}

	// Stream deduplicated candidate pairs that survive the cheap pre-filters
	// to a pool of workers for the expensive APTED verification. Streaming
	// keeps memory bounded: the full candidate list is never materialized.
	type candidatePair struct{ a, b int }
	seenPairs := make(map[[2]int]struct{})

	workers := cd.effectiveWorkers(len(records))
	candCh := make(chan candidatePair, 4*workers)
	verified := make([][]*ClonePair, workers)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wd := cd
			if workers > 1 {
				wd = cd.newWorkerDetector()
			}
			for c := range candCh {
				// Keep draining on cancellation so the producer never blocks.
				if isCancelled(ctx) {
					continue
				}
				pair := wd.compareFragments(cd.fragments[c.a], cd.fragments[c.b])
				if pair != nil && wd.isSignificantClone(pair) {
					verified[w] = append(verified[w], pair)
				}
			}
		}()
	}

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
			candCh <- candidatePair{a: a, b: b}
		}
	}
	close(candCh)
	wg.Wait()

	for _, pairs := range verified {
		cd.clonePairs = append(cd.clonePairs, pairs...)
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
	cd.groupClones(thr, k)

	return cd.buildCloneDetectionResult()
}

// buildCloneDetectionResult builds a CloneDetectionResult from the detector's
// current state. It derives all statistics directly from the detected pairs and
// groups so callers cannot accidentally pass a raw fragment count in their place.
func (cd *CloneDetector) buildCloneDetectionResult() *CloneDetectionResult {
	stats := &CloneDetectionStatistics{
		TotalFragments:   len(cd.fragments),
		TotalClonePairs:  len(cd.clonePairs),
		TotalCloneGroups: len(cd.cloneGroups),
		ClonesByType:     make(map[string]int),
	}

	seenFragments := make(map[*CodeFragment]struct{})
	countFragment := func(f *CodeFragment) {
		if f != nil {
			seenFragments[f] = struct{}{}
		}
	}
	for _, pair := range cd.clonePairs {
		countFragment(pair.Fragment1)
		countFragment(pair.Fragment2)
		stats.ClonesByType[pair.CloneType.String()]++
	}
	for _, group := range cd.cloneGroups {
		if group == nil {
			continue
		}
		for _, f := range group.Fragments {
			countFragment(f)
		}
	}
	stats.TotalClones = len(seenFragments)

	if len(cd.clonePairs) > 0 {
		totalSim := 0.0
		for _, pair := range cd.clonePairs {
			totalSim += pair.Similarity
		}
		stats.AverageSimilarity = totalSim / float64(len(cd.clonePairs))
	}

	return &CloneDetectionResult{
		Pairs:      cd.clonePairs,
		Groups:     cd.cloneGroups,
		Statistics: stats,
	}
}

// prepareFragments converts AST fragments to tree nodes, populates clone
// features, and caches each fragment's core/clone projection so per-pair
// comparisons don't re-convert trees.
func (cd *CloneDetector) prepareFragments() {
	for i, fragment := range cd.fragments {
		if fragment == nil {
			continue
		}
		if fragment.TreeNode == nil && fragment.ASTNode != nil {
			fragment.TreeNode = cd.converter.ConvertAST(fragment.ASTNode)
		}
		if fragment.TreeNode == nil {
			continue
		}
		coreapted.PrepareTreeForAPTED(fragment.TreeNode)
		fragment.id = i
		fragment.core = toCoreFragment(fragment, i)
		features, _ := cd.featureExtractor.ExtractFeatures(fragment.core.ASTNode)
		fragment.Features = features
		fragment.core.Features = features
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
	if !coreclone.ShouldCompareFragments(fragment1.coreFragment(), fragment2.coreFragment()) {
		return nil
	}

	// Jaccard pre-filter: reject clear non-clones before expensive APTED/classifier work.
	// Only used for rejection — all non-rejected pairs proceed to APTED-based classification
	// for accurate clone typing and distance computation.
	if !cd.pairClassifier.PassesJaccardPreFilter(fragment1.coreFragment(), fragment2.coreFragment()) {
		return nil
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
			Fragment1: fragment1,
			Fragment2: fragment2,
			// Semantic Type-4 pairs are never exact textual matches (the
			// classifier checks Type-1 first), so apply the same non-textual
			// cap as classifyClonePair to keep similarity comparable across
			// clone types.
			Similarity: cd.pairClassifier.CapNonTextualSimilarity(result.Similarity),
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

	confidence := coreclone.CalculateConfidence(fragment1.coreFragment(), fragment2.coreFragment(), similarity)

	return &ClonePair{
		Fragment1:  fragment1,
		Fragment2:  fragment2,
		Similarity: similarity,
		Distance:   distance,
		CloneType:  cloneType,
		Confidence: confidence,
	}
}

// classifyClonePair classifies a clone pair via the core pair classifier,
// gating Type-1 on exact textual match and Type-2 on syntactic (normalized
// AST) similarity. Returns the (possibly capped) similarity actually used for
// classification.
func (cd *CloneDetector) classifyClonePair(fragment1, fragment2 *CodeFragment, similarity float64) (CloneType, float64) {
	coreType, capped := cd.pairClassifier.ClassifyPair(fragment1.coreFragment(), fragment2.coreFragment(), similarity)
	return CloneType(coreType), capped
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

	// Fragments kept past the size gate earn their place only as exact
	// duplicates of each other; at that size anything less is noise.
	if pair.Fragment1.belowSizeGate || pair.Fragment2.belowSizeGate {
		return pair.Fragment1.Hash != "" && pair.Fragment1.Hash == pair.Fragment2.Hash
	}

	// Additional filtering based on fragment characteristics
	minSize := math.Min(float64(pair.Fragment1.Size), float64(pair.Fragment2.Size))
	return minSize >= float64(cd.cloneDetectorConfig.MinNodes)
}

func (cd *CloneDetector) groupClones(threshold float64, k int) {
	var strategy coreclone.GroupingStrategy[*CodeFragment]
	if cd.cloneDetectorConfig.GroupingMode == GroupingModeCentroid {
		strategy = newCentroidCompatibilityStrategy(cd, threshold)
	} else {
		strategy = coreclone.NewGroupingStrategy[*CodeFragment](coreclone.GroupingConfig{
			Mode:      cd.cloneDetectorConfig.GroupingMode.coreMode(),
			Threshold: threshold,
			KCoreK:    k,
		})
	}
	cd.groupClonesWithStrategy(strategy, threshold)
}

// groupClonesWithStrategy groups clone pairs using a core/clone grouping
// strategy, then applies the shared dedup and suppression passes.
func (cd *CloneDetector) groupClonesWithStrategy(strategy coreclone.GroupingStrategy[*CodeFragment], threshold float64) {
	if strategy == nil {
		cd.cloneGroups = []*CloneGroup{}
		return
	}

	// Assign deterministic unique item IDs (first appearance order) so the
	// grouping framework can identify fragments regardless of origin.
	seen := make(map[*CodeFragment]struct{}, len(cd.clonePairs)*2)
	nextID := 0
	assignID := func(f *CodeFragment) {
		if _, ok := seen[f]; !ok {
			seen[f] = struct{}{}
			f.id = nextID
			nextID++
		}
	}

	corePairs := make([]*coreclone.ItemPair[*CodeFragment], 0, len(cd.clonePairs))
	originals := make(map[*coreclone.ItemPair[*CodeFragment]]*ClonePair, len(cd.clonePairs))
	for _, pair := range cd.clonePairs {
		if pair == nil || pair.Fragment1 == nil || pair.Fragment2 == nil {
			continue
		}
		assignID(pair.Fragment1)
		assignID(pair.Fragment2)
		converted := &coreclone.ItemPair[*CodeFragment]{
			Item1:      pair.Fragment1,
			Item2:      pair.Fragment2,
			Similarity: pair.Similarity,
			PairType:   coredomain.CloneType(pair.CloneType),
		}
		corePairs = append(corePairs, converted)
		originals[converted] = pair
	}

	groups := strategy.GroupItems(corePairs)
	if cd.cloneDetectorConfig.GroupingMode == GroupingModeStar {
		groups = filterStarGroupsByMedoid(groups, corePairs, threshold)
	}
	memberResult := coreclone.DedupeStrictSubsetGroupMembers(groups, corePairs)
	groupResult := coreclone.DedupeCoveredGroups(memberResult.Groups)
	for key := range memberResult.Suppressed {
		groupResult.Suppressed[key] = struct{}{}
	}

	// Drop the suppressed members' pairs before the groups are validated, so a
	// group that only held together through a suppressed member is re-split
	// rather than reported as one family.
	corePairs = coreclone.FilterPairsWithSuppressedMembers(corePairs, groupResult.Suppressed)
	corePairs = coreclone.FilterSuppressedPairs(corePairs, groupResult.SuppressedPairs)

	groups = splitDisconnectedGroups(groupResult.Groups, corePairs)
	groups = coreclone.FilterGroupsWithoutBackingPairs(groups, corePairs)
	if cd.cloneDetectorConfig.GroupingMode == GroupingModeCentroid {
		cd.refreshCentroidGroupMetadata(groups)
	}
	// k-core deliberately prunes fragments with fewer than k similar
	// neighbours, so its excluded pairs must stay excluded.
	if cd.cloneDetectorConfig.GroupingMode != GroupingModeKCore {
		groups = appendUncoveredPairGroups(groups, corePairs, threshold)
	}

	cd.clonePairs = cd.clonePairs[:0]
	for _, pair := range corePairs {
		cd.clonePairs = append(cd.clonePairs, originals[pair])
	}
	cd.cloneGroups = make([]*CloneGroup, 0, len(groups))
	for _, group := range groups {
		cd.cloneGroups = append(cd.cloneGroups, &CloneGroup{
			ID:         len(cd.cloneGroups),
			Fragments:  group.Items,
			CloneType:  CloneType(group.GroupType),
			Similarity: group.Similarity,
			Size:       len(group.Items),
		})
	}
}

// splitDisconnectedGroups breaks each group into the connected components of
// the surviving pair graph.
//
// Grouping runs before members are suppressed, so a member that is later
// dropped can leave the rest of a group with no pair holding it together. The
// group was still reported whole, which presented fragments as one duplication
// family although no reported pair connects them.
func splitDisconnectedGroups(
	groups []*coreclone.ItemGroup[*CodeFragment],
	pairs []*coreclone.ItemPair[*CodeFragment],
) []*coreclone.ItemGroup[*CodeFragment] {
	if len(groups) == 0 {
		return groups
	}

	adjacency := make(map[[2]int]struct{}, len(pairs))
	for _, pair := range pairs {
		if pair == nil {
			continue
		}
		adjacency[fragmentPairKey(pair.Item1, pair.Item2)] = struct{}{}
	}

	out := make([]*coreclone.ItemGroup[*CodeFragment], 0, len(groups))
	for _, group := range groups {
		if group == nil {
			continue
		}
		components := connectedFragmentComponents(group.Items, adjacency)
		if len(components) < 2 {
			out = append(out, group)
			continue
		}
		// Similarity and type are recomputed downstream from the surviving pairs.
		for _, members := range components {
			if len(members) < 2 {
				continue
			}
			out = append(out, &coreclone.ItemGroup[*CodeFragment]{
				ID:         group.ID,
				Items:      members,
				GroupType:  group.GroupType,
				Similarity: group.Similarity,
			})
		}
	}
	return out
}

// fragmentPairKey identifies a fragment pair independently of its order.
func fragmentPairKey(a, b *CodeFragment) [2]int {
	if a.id > b.id {
		return [2]int{b.id, a.id}
	}
	return [2]int{a.id, b.id}
}

// connectedFragmentComponents partitions items into connected components,
// preserving the input order within and across components.
func connectedFragmentComponents(items []*CodeFragment, adjacency map[[2]int]struct{}) [][]*CodeFragment {
	parent := make([]int, len(items))
	for i := range parent {
		parent[i] = i
	}
	find := func(i int) int {
		for parent[i] != i {
			parent[i] = parent[parent[i]]
			i = parent[i]
		}
		return i
	}
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if _, ok := adjacency[fragmentPairKey(items[i], items[j])]; !ok {
				continue
			}
			if rootI, rootJ := find(i), find(j); rootI != rootJ {
				parent[rootI] = rootJ
			}
		}
	}

	roots := make([]int, 0, len(items))
	buckets := make(map[int][]*CodeFragment, len(items))
	for i, item := range items {
		root := find(i)
		if _, ok := buckets[root]; !ok {
			roots = append(roots, root)
		}
		buckets[root] = append(buckets[root], item)
	}

	components := make([][]*CodeFragment, 0, len(roots))
	for _, root := range roots {
		components = append(components, buckets[root])
	}
	return components
}

// appendUncoveredPairGroups adds a two-member group for every remaining pair
// that no group already covers.
//
// The partitioning strategies (complete linkage, star, centroid) put a fragment
// in at most one group, so a fragment that clones several others keeps only its
// best cluster and its other pairs reach no group at all. Reports that render
// groups would then drop those duplicates entirely, even though the pair was
// detected and scored.
func appendUncoveredPairGroups(
	groups []*coreclone.ItemGroup[*CodeFragment],
	pairs []*coreclone.ItemPair[*CodeFragment],
	threshold float64,
) []*coreclone.ItemGroup[*CodeFragment] {
	// Index which groups hold each fragment. Materialising every member
	// combination instead would cost O(n^2) for a group of n members, most of
	// it recording pairs that were never detected.
	groupsByFragment := make(map[int]map[int]struct{}, len(groups))
	assign := func(fragmentID, groupIndex int) {
		indexes, ok := groupsByFragment[fragmentID]
		if !ok {
			indexes = make(map[int]struct{}, 1)
			groupsByFragment[fragmentID] = indexes
		}
		indexes[groupIndex] = struct{}{}
	}
	for index, group := range groups {
		if group == nil {
			continue
		}
		for _, item := range group.Items {
			assign(item.id, index)
		}
	}

	shareGroup := func(a, b *CodeFragment) bool {
		left, right := groupsByFragment[a.id], groupsByFragment[b.id]
		if len(right) < len(left) {
			left, right = right, left
		}
		for index := range left {
			if _, ok := right[index]; ok {
				return true
			}
		}
		return false
	}

	for _, pair := range pairs {
		if pair == nil || pair.Similarity < threshold {
			continue
		}
		if shareGroup(pair.Item1, pair.Item2) {
			continue
		}

		members := []*CodeFragment{pair.Item1, pair.Item2}
		if members[1].id < members[0].id {
			members[0], members[1] = members[1], members[0]
		}
		index := len(groups)
		groups = append(groups, &coreclone.ItemGroup[*CodeFragment]{
			Items:      members,
			GroupType:  pair.PairType,
			Similarity: pair.Similarity,
		})
		assign(members[0].id, index)
		assign(members[1].id, index)
	}
	return groups
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

// GetStatistics returns clone detection statistics as a map for backward
// compatibility with existing callers and tests. New code should prefer the
// structured CloneDetectionResult returned by DetectClones* methods.
func (cd *CloneDetector) GetStatistics() map[string]interface{} {
	result := cd.buildCloneDetectionResult()
	stats := result.Statistics

	m := make(map[string]interface{})
	m["total_fragments"] = stats.TotalFragments
	m["total_clone_pairs"] = stats.TotalClonePairs
	m["total_clone_groups"] = stats.TotalCloneGroups
	m["clone_types"] = stats.ClonesByType
	m["average_similarity"] = stats.AverageSimilarity

	return m
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
