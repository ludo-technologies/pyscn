package analyzer

import (
	"context"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCloneClassifier tests the multi-dimensional clone classifier
func TestCloneClassifier(t *testing.T) {
	t.Run("NewCloneClassifier", func(t *testing.T) {
		config := &CloneClassifierConfig{
			Type1Threshold:         domain.DefaultType1CloneThreshold,
			Type2Threshold:         domain.DefaultType2CloneThreshold,
			Type3Threshold:         domain.DefaultType3CloneThreshold,
			Type4Threshold:         domain.DefaultType4CloneThreshold,
			EnableTextualAnalysis:  true,
			EnableSemanticAnalysis: true,
		}

		classifier := NewCloneClassifier(config)
		require.NotNil(t, classifier)
		assert.Equal(t, domain.DefaultType1CloneThreshold, classifier.type1Threshold)
		assert.Equal(t, domain.DefaultType2CloneThreshold, classifier.type2Threshold)
		assert.Equal(t, domain.DefaultType3CloneThreshold, classifier.type3Threshold)
		assert.Equal(t, domain.DefaultType4CloneThreshold, classifier.type4Threshold)
	})

	t.Run("ClassifyCloneWithNilFragments", func(t *testing.T) {
		config := &CloneClassifierConfig{
			Type1Threshold: 0.95,
			Type2Threshold: 0.85,
			Type3Threshold: 0.80,
			Type4Threshold: 0.75,
		}
		classifier := NewCloneClassifier(config)

		// Both nil
		result := classifier.ClassifyClone(nil, nil)
		assert.Nil(t, result)

		// One nil
		fragment := &CodeFragment{Content: "test"}
		result = classifier.ClassifyClone(fragment, nil)
		assert.Nil(t, result)

		result = classifier.ClassifyClone(nil, fragment)
		assert.Nil(t, result)
	})

	t.Run("ClassifyCloneSimple", func(t *testing.T) {
		config := &CloneClassifierConfig{
			Type1Threshold: 0.95,
			Type2Threshold: 0.85,
			Type3Threshold: 0.80,
			Type4Threshold: 0.75,
		}
		classifier := NewCloneClassifier(config)

		// Test with nil fragments
		cloneType, similarity, confidence := classifier.ClassifyCloneSimple(nil, nil)
		assert.Equal(t, CloneType(0), cloneType)
		assert.Equal(t, 0.0, similarity)
		assert.Equal(t, 0.0, confidence)
	})
}

// TestTextualSimilarityAnalyzer tests textual similarity analysis
func TestTextualSimilarityAnalyzer(t *testing.T) {
	t.Run("NewTextualSimilarityAnalyzer", func(t *testing.T) {
		analyzer := NewTextualSimilarityAnalyzer()
		require.NotNil(t, analyzer)
		assert.Equal(t, "textual", analyzer.GetName())
	})

	t.Run("IdenticalContent", func(t *testing.T) {
		analyzer := NewTextualSimilarityAnalyzer()

		f1 := &CodeFragment{Content: "def hello():\n    return 'world'"}
		f2 := &CodeFragment{Content: "def hello():\n    return 'world'"}

		similarity := analyzer.ComputeSimilarity(f1, f2)
		assert.Equal(t, 1.0, similarity)
	})

	t.Run("WhitespaceDifference", func(t *testing.T) {
		analyzer := NewTextualSimilarityAnalyzer()

		f1 := &CodeFragment{Content: "def hello():\n    return 'world'"}
		f2 := &CodeFragment{Content: "def hello():\n        return 'world'"}

		// After normalization, whitespace differences should result in high similarity
		similarity := analyzer.ComputeSimilarity(f1, f2)
		assert.Greater(t, similarity, 0.9)
	})

	t.Run("CommentRemoval", func(t *testing.T) {
		analyzer := NewTextualSimilarityAnalyzer()

		f1 := &CodeFragment{Content: "def hello():\n    return 'world'"}
		f2 := &CodeFragment{Content: "def hello():  # greeting\n    return 'world'"}

		// After comment removal, should be identical
		similarity := analyzer.ComputeSimilarity(f1, f2)
		assert.Equal(t, 1.0, similarity)
	})

	t.Run("EmptyContent", func(t *testing.T) {
		analyzer := NewTextualSimilarityAnalyzer()

		// Both empty
		f1 := &CodeFragment{Content: ""}
		f2 := &CodeFragment{Content: ""}
		similarity := analyzer.ComputeSimilarity(f1, f2)
		assert.Equal(t, 1.0, similarity)

		// One empty
		f3 := &CodeFragment{Content: "def hello(): pass"}
		similarity = analyzer.ComputeSimilarity(f1, f3)
		assert.Equal(t, 0.0, similarity)
	})

	t.Run("NilFragment", func(t *testing.T) {
		analyzer := NewTextualSimilarityAnalyzer()

		f1 := &CodeFragment{Content: "test"}
		similarity := analyzer.ComputeSimilarity(f1, nil)
		assert.Equal(t, 0.0, similarity)

		similarity = analyzer.ComputeSimilarity(nil, f1)
		assert.Equal(t, 0.0, similarity)
	})

	t.Run("LevenshteinSimilarity", func(t *testing.T) {
		analyzer := NewTextualSimilarityAnalyzer()

		f1 := &CodeFragment{Content: "def calculate_sum(a, b):"}
		f2 := &CodeFragment{Content: "def calculate_total(x, y):"}

		similarity := analyzer.ComputeSimilarity(f1, f2)
		// Should be similar but not identical
		assert.Greater(t, similarity, 0.5)
		assert.Less(t, similarity, 1.0)
	})
}

// TestSyntacticSimilarityAnalyzer tests syntactic similarity analysis
func TestSyntacticSimilarityAnalyzer(t *testing.T) {
	t.Run("NewSyntacticSimilarityAnalyzer", func(t *testing.T) {
		analyzer := NewSyntacticSimilarityAnalyzer()
		require.NotNil(t, analyzer)
		assert.Equal(t, "syntactic", analyzer.GetName())
	})

	t.Run("WithOptions", func(t *testing.T) {
		analyzer := NewSyntacticSimilarityAnalyzerWithOptions(true, false)
		require.NotNil(t, analyzer)
	})

	t.Run("NilFragments", func(t *testing.T) {
		analyzer := NewSyntacticSimilarityAnalyzer()

		similarity := analyzer.ComputeSimilarity(nil, nil)
		assert.Equal(t, 0.0, similarity)

		fragment := &CodeFragment{}
		similarity = analyzer.ComputeSimilarity(fragment, nil)
		assert.Equal(t, 0.0, similarity)
	})

	t.Run("ComputeDistance", func(t *testing.T) {
		analyzer := NewSyntacticSimilarityAnalyzer()

		// Test with nil
		distance := analyzer.ComputeDistance(nil, nil)
		assert.Equal(t, 0.0, distance)
	})

	t.Run("JaccardSimilarity_Type2Clones", func(t *testing.T) {
		// Type-2 clones: same structure, different identifiers/literals
		// These should have high similarity
		analyzer := NewSyntacticSimilarityAnalyzer()
		converter := NewTreeConverter()
		p := parser.New()
		ctx := context.Background()

		code1 := `def foo(x):
    return x + 1`
		code2 := `def bar(y):
    return y + 2`

		result1, err := p.Parse(ctx, []byte(code1))
		require.NoError(t, err)
		result2, err := p.Parse(ctx, []byte(code2))
		require.NoError(t, err)

		f1 := &CodeFragment{TreeNode: converter.ConvertAST(result1.AST)}
		f2 := &CodeFragment{TreeNode: converter.ConvertAST(result2.AST)}

		PrepareTreeForAPTED(f1.TreeNode)
		PrepareTreeForAPTED(f2.TreeNode)

		similarity := analyzer.ComputeSimilarity(f1, f2)
		// Type-2 clones should have high similarity (>= 0.80)
		assert.GreaterOrEqual(t, similarity, 0.80,
			"Type-2 clones (renamed identifiers/literals) should have high similarity")
	})

	// True positive test: Type-2 clones SHOULD be detected via CloneClassifier
	t.Run("Type2Clone_TruePositive_DetectedViaClassifier", func(t *testing.T) {
		// Type-2 clones: structurally identical code with only identifier/literal differences.
		// These MUST be detected as Type-2 clones through the full CloneClassifier flow.
		converter := NewTreeConverter()
		p := parser.New()
		ctx := context.Background()

		// Two functions with identical structure, only names and literals differ
		code1 := `def calculate_sum(a, b):
    result = a + b
    if result > 100:
        return result * 2
    return result`

		code2 := `def compute_total(x, y):
    value = x + y
    if value > 200:
        return value * 3
    return value`

		result1, err := p.Parse(ctx, []byte(code1))
		require.NoError(t, err)
		result2, err := p.Parse(ctx, []byte(code2))
		require.NoError(t, err)

		f1 := &CodeFragment{TreeNode: converter.ConvertAST(result1.AST)}
		f2 := &CodeFragment{TreeNode: converter.ConvertAST(result2.AST)}

		PrepareTreeForAPTED(f1.TreeNode)
		PrepareTreeForAPTED(f2.TreeNode)

		// Use CloneClassifier with default thresholds
		classifier := NewCloneClassifier(&CloneClassifierConfig{
			Type1Threshold: domain.DefaultType1CloneThreshold,
			Type2Threshold: domain.DefaultType2CloneThreshold,
			Type3Threshold: domain.DefaultType3CloneThreshold,
			Type4Threshold: domain.DefaultType4CloneThreshold,
		})

		result := classifier.ClassifyClone(f1, f2)

		// These ARE Type-2 clones and MUST be detected
		require.NotNil(t, result, "Type-2 clones should be detected")
		assert.Equal(t, Type2Clone, result.CloneType,
			"Structurally identical code with different identifiers should be Type-2 clone")
		assert.GreaterOrEqual(t, result.Similarity, domain.DefaultType2CloneThreshold,
			"Type-2 clone similarity should meet threshold")
	})

	t.Run("JaccardSimilarity_DifferentStructures", func(t *testing.T) {
		// Structurally different code should have low similarity
		analyzer := NewSyntacticSimilarityAnalyzer()
		converter := NewTreeConverter()
		p := parser.New()
		ctx := context.Background()

		code1 := `def foo(x):
    return x + 1`
		code2 := `def bar(items):
    for item in items:
        print(item)
    return len(items)`

		result1, err := p.Parse(ctx, []byte(code1))
		require.NoError(t, err)
		result2, err := p.Parse(ctx, []byte(code2))
		require.NoError(t, err)

		f1 := &CodeFragment{TreeNode: converter.ConvertAST(result1.AST)}
		f2 := &CodeFragment{TreeNode: converter.ConvertAST(result2.AST)}

		PrepareTreeForAPTED(f1.TreeNode)
		PrepareTreeForAPTED(f2.TreeNode)

		similarity := analyzer.ComputeSimilarity(f1, f2)
		// Different structures should have low similarity (< 0.50)
		assert.Less(t, similarity, 0.50,
			"Structurally different code should have low similarity")
	})

	t.Run("GetExtractor", func(t *testing.T) {
		analyzer := NewSyntacticSimilarityAnalyzer()
		extractor := analyzer.GetExtractor()
		assert.NotNil(t, extractor)
	})

	// Issue #292: False positives for structurally different dataclasses
	t.Run("Type2Clone_DifferentDataclasses_Issue292", func(t *testing.T) {
		// Regression test for issue #292: Different dataclasses were incorrectly
		// identified as Type-2 clones with 97%+ similarity due to APTED treating
		// structurally similar but semantically different code as near-identical.
		//
		// This test verifies the full CloneClassifier flow to ensure these
		// different dataclasses are NOT classified as Type-2 clones.
		converter := NewTreeConverter()
		p := parser.New()
		ctx := context.Background()

		// Simplified version of ScopeConfig from issue #292
		code1 := `@dataclass
class ScopeConfig:
    allowlist: frozenset = None
    before_hooks: tuple = ()
    after_hooks: tuple = ()
    timeout: float = None

    def __post_init__(self):
        self.timeout = validate_timeout(self.timeout)`

		// Simplified version of InMemoryMetrics from issue #292
		code2 := `@dataclass
class InMemoryMetrics:
    counters: dict = field(default_factory=dict)
    histograms: dict = field(default_factory=dict)
    _lock: Lock = field(default_factory=Lock)

    def inc_counter(self, name, value, labels):
        with self._lock:
            self.counters[name] = self.counters.get(name, 0) + value

    def observe_histogram(self, name, value, labels):
        with self._lock:
            if name not in self.histograms:
                self.histograms[name] = []
            self.histograms[name].append(value)

    def reset(self):
        with self._lock:
            self.counters.clear()
            self.histograms.clear()`

		result1, err := p.Parse(ctx, []byte(code1))
		require.NoError(t, err)
		result2, err := p.Parse(ctx, []byte(code2))
		require.NoError(t, err)

		f1 := &CodeFragment{TreeNode: converter.ConvertAST(result1.AST)}
		f2 := &CodeFragment{TreeNode: converter.ConvertAST(result2.AST)}

		PrepareTreeForAPTED(f1.TreeNode)
		PrepareTreeForAPTED(f2.TreeNode)

		// Use CloneClassifier with default thresholds to test actual Type-2 detection
		classifier := NewCloneClassifier(&CloneClassifierConfig{
			Type1Threshold: domain.DefaultType1CloneThreshold,
			Type2Threshold: domain.DefaultType2CloneThreshold,
			Type3Threshold: domain.DefaultType3CloneThreshold,
			Type4Threshold: domain.DefaultType4CloneThreshold,
		})

		result := classifier.ClassifyClone(f1, f2)

		// Issue #292: These were incorrectly reported as 97.6% similar Type-2 clones.
		// With Jaccard coefficient, they should NOT be classified as Type-2 clones.
		if result != nil && result.CloneType == Type2Clone {
			t.Errorf("Different dataclasses should NOT be classified as Type-2 clones (issue #292), "+
				"got similarity: %.1f%%", result.Similarity*100)
		}

		// Also verify the raw syntactic similarity is well below the threshold
		syntacticAnalyzer := NewSyntacticSimilarityAnalyzer()
		similarity := syntacticAnalyzer.ComputeSimilarity(f1, f2)
		assert.Less(t, similarity, domain.DefaultType2CloneThreshold,
			"Syntactic similarity should be below Type-2 threshold (issue #292)")
	})

	// Issue #292: False positives for classes with different method counts
	t.Run("Type2Clone_DifferentClassStructures_Issue292", func(t *testing.T) {
		// Regression test for issue #292: TracingHook vs TestMetricsHook
		// were incorrectly identified as 98.9% similar Type-2 clones.
		//
		// This test verifies the full CloneClassifier flow to ensure these
		// different classes are NOT classified as Type-2 clones.
		converter := NewTreeConverter()
		p := parser.New()
		ctx := context.Background()

		// Simplified version of TracingHook from issue #292
		code1 := `class TracingHook:
    def __init__(self, tracer, record_output=True):
        self._tracer = tracer
        self._record_output = record_output
        self._active_spans = {}
        self._lock = Lock()

    def __call__(self, event):
        if event.phase == "start":
            self._handle_start(event)
        elif event.phase == "exit":
            self._handle_exit(event)

    def _handle_start(self, event):
        if event.pid is None:
            return
        attrs = self._build_attributes(event)
        span = self._tracer.start_span(event.program, attrs)
        with self._lock:
            self._active_spans[event.pid] = span

    def _handle_exit(self, event):
        if event.pid is None:
            return
        with self._lock:
            span = self._active_spans.pop(event.pid, None)
        if span:
            span.end()`

		// Simplified version of TestMetricsHook from issue #292
		code2 := `class TestMetricsHook:
    def test_increments_counter(self):
        metrics = InMemoryMetrics()
        hook = MetricsHook(metrics)
        cmd.run_sync()
        assert metrics.counters.get("total") == 1.0

    def test_counts_output_lines(self):
        metrics = InMemoryMetrics()
        hook = MetricsHook(metrics)
        cmd.run_sync()
        assert metrics.counters.get("stdout") == 2.0

    def test_records_duration(self):
        metrics = InMemoryMetrics()
        hook = MetricsHook(metrics)
        cmd.run_sync()
        durations = metrics.histograms.get("duration", [])
        assert len(durations) == 1`

		result1, err := p.Parse(ctx, []byte(code1))
		require.NoError(t, err)
		result2, err := p.Parse(ctx, []byte(code2))
		require.NoError(t, err)

		f1 := &CodeFragment{TreeNode: converter.ConvertAST(result1.AST)}
		f2 := &CodeFragment{TreeNode: converter.ConvertAST(result2.AST)}

		PrepareTreeForAPTED(f1.TreeNode)
		PrepareTreeForAPTED(f2.TreeNode)

		// Use CloneClassifier with default thresholds to test actual Type-2 detection
		classifier := NewCloneClassifier(&CloneClassifierConfig{
			Type1Threshold: domain.DefaultType1CloneThreshold,
			Type2Threshold: domain.DefaultType2CloneThreshold,
			Type3Threshold: domain.DefaultType3CloneThreshold,
			Type4Threshold: domain.DefaultType4CloneThreshold,
		})

		result := classifier.ClassifyClone(f1, f2)

		// Issue #292: These were incorrectly reported as 98.9% similar Type-2 clones.
		// With Jaccard coefficient, they should NOT be classified as Type-2 clones.
		if result != nil && result.CloneType == Type2Clone {
			t.Errorf("Different class structures should NOT be classified as Type-2 clones (issue #292), "+
				"got similarity: %.1f%%", result.Similarity*100)
		}

		// Also verify the raw syntactic similarity is well below the threshold
		syntacticAnalyzer := NewSyntacticSimilarityAnalyzer()
		similarity := syntacticAnalyzer.ComputeSimilarity(f1, f2)
		assert.Less(t, similarity, domain.DefaultType2CloneThreshold,
			"Syntactic similarity should be below Type-2 threshold (issue #292)")
	})
}

// TestStructuralSimilarityAnalyzer tests structural similarity analysis
func TestStructuralSimilarityAnalyzer(t *testing.T) {
	t.Run("NewStructuralSimilarityAnalyzer", func(t *testing.T) {
		analyzer := NewStructuralSimilarityAnalyzer()
		require.NotNil(t, analyzer)
		assert.Equal(t, "structural", analyzer.GetName())
		assert.NotNil(t, analyzer.GetAnalyzer())
	})

	t.Run("WithCostModel", func(t *testing.T) {
		costModel := NewPythonCostModel()
		analyzer := NewStructuralSimilarityAnalyzerWithCostModel(costModel)
		require.NotNil(t, analyzer)
	})

	t.Run("NilFragments", func(t *testing.T) {
		analyzer := NewStructuralSimilarityAnalyzer()

		similarity := analyzer.ComputeSimilarity(nil, nil)
		assert.Equal(t, 0.0, similarity)
	})

	t.Run("ComputeDistance", func(t *testing.T) {
		analyzer := NewStructuralSimilarityAnalyzer()

		distance := analyzer.ComputeDistance(nil, nil)
		assert.Equal(t, 0.0, distance)
	})
}

// TestSemanticSimilarityAnalyzer tests semantic similarity analysis
func TestSemanticSimilarityAnalyzer(t *testing.T) {
	t.Run("NewSemanticSimilarityAnalyzer", func(t *testing.T) {
		analyzer := NewSemanticSimilarityAnalyzer()
		require.NotNil(t, analyzer)
		assert.Equal(t, "semantic", analyzer.GetName())
	})

	t.Run("NilFragments", func(t *testing.T) {
		analyzer := NewSemanticSimilarityAnalyzer()

		similarity := analyzer.ComputeSimilarity(nil, nil)
		assert.Equal(t, 0.0, similarity)

		fragment := &CodeFragment{}
		similarity = analyzer.ComputeSimilarity(fragment, nil)
		assert.Equal(t, 0.0, similarity)
	})

	t.Run("CFGFeatureExtraction", func(t *testing.T) {
		analyzer := NewSemanticSimilarityAnalyzer()

		// Test with nil CFG
		features := analyzer.ExtractFeatures(nil)
		require.NotNil(t, features)
		assert.Equal(t, 0, features.BlockCount)
		assert.Equal(t, 0, features.EdgeCount)
	})
}

// TestCloneDetectorWithMultiDimensionalAnalysis tests the integrated classifier
func TestCloneDetectorWithMultiDimensionalAnalysis(t *testing.T) {
	t.Run("EnableMultiDimensionalAnalysis", func(t *testing.T) {
		config := DefaultCloneDetectorConfig()
		config.EnableMultiDimensionalAnalysis = true
		config.EnableTextualAnalysis = true
		config.EnableSemanticAnalysis = true

		detector := NewCloneDetector(config)
		require.NotNil(t, detector)
		require.NotNil(t, detector.classifier)
	})

	t.Run("DisabledByDefault", func(t *testing.T) {
		config := DefaultCloneDetectorConfig()
		detector := NewCloneDetector(config)
		require.NotNil(t, detector)
		assert.Nil(t, detector.classifier)
	})

	t.Run("IdenticalFunctionsWithClassifier", func(t *testing.T) {
		config := DefaultCloneDetectorConfig()
		config.EnableMultiDimensionalAnalysis = true
		config.MinLines = 1
		config.MinNodes = 1

		detector := NewCloneDetector(config)

		// Create identical function fragments with more lines
		source1 := `def calculate(x, y):
    result = x + y
    return result

def helper():
    pass`
		source2 := `def calculate(x, y):
    result = x + y
    return result

def helper():
    pass`

		ctx := context.Background()
		p := parser.New()
		result1, err1 := p.Parse(ctx, []byte(source1))
		result2, err2 := p.Parse(ctx, []byte(source2))
		require.NoError(t, err1)
		require.NoError(t, err2)

		fragments1 := detector.ExtractFragments([]*parser.Node{result1.AST}, "test1.py")
		fragments2 := detector.ExtractFragments([]*parser.Node{result2.AST}, "test2.py")

		require.NotEmpty(t, fragments1)
		require.NotEmpty(t, fragments2)

		// Prepare tree nodes
		converter := NewTreeConverter()
		for _, f := range fragments1 {
			f.TreeNode = converter.ConvertAST(f.ASTNode)
			PrepareTreeForAPTED(f.TreeNode)
		}
		for _, f := range fragments2 {
			f.TreeNode = converter.ConvertAST(f.ASTNode)
			PrepareTreeForAPTED(f.TreeNode)
		}

		// Compare with classifier
		result := detector.classifier.ClassifyClone(fragments1[0], fragments2[0])
		require.NotNil(t, result)
		assert.Greater(t, result.Similarity, 0.9)
	})
}

// TestExtractFragmentsWithSource tests source content extraction
func TestExtractFragmentsWithSource(t *testing.T) {
	t.Run("ExtractWithContent", func(t *testing.T) {
		config := DefaultCloneDetectorConfig()
		config.EnableTextualAnalysis = true
		config.MinLines = 1
		config.MinNodes = 1

		detector := NewCloneDetector(config)

		source := `def hello():
    x = 1
    y = 2
    return x + y

def goodbye():
    a = 3
    b = 4
    return a + b`

		ctx := context.Background()
		p := parser.New()
		result, err := p.Parse(ctx, []byte(source))
		require.NoError(t, err)

		fragments := detector.ExtractFragmentsWithSource([]*parser.Node{result.AST}, "test.py", []byte(source))

		// Should have extracted function definitions
		require.NotEmpty(t, fragments)

		// Check that content is populated
		for _, f := range fragments {
			assert.NotEmpty(t, f.Content, "Fragment content should be populated")
		}
	})

	t.Run("ExtractWithoutContent", func(t *testing.T) {
		config := DefaultCloneDetectorConfig()
		config.EnableTextualAnalysis = false
		config.MinLines = 1
		config.MinNodes = 1

		detector := NewCloneDetector(config)

		source := `def hello():
    x = 1
    y = 2
    return x + y`

		ctx := context.Background()
		p := parser.New()
		result, err := p.Parse(ctx, []byte(source))
		require.NoError(t, err)

		// Even with source provided, content should not be populated when disabled
		fragments := detector.ExtractFragmentsWithSource([]*parser.Node{result.AST}, "test.py", []byte(source))
		require.NotEmpty(t, fragments)

		for _, f := range fragments {
			assert.Empty(t, f.Content, "Fragment content should be empty when textual analysis is disabled")
		}
	})
}

// TestCFGFeatureComparison tests CFG feature comparison for semantic similarity
func TestCFGFeatureComparison(t *testing.T) {
	t.Run("IdenticalFeatures", func(t *testing.T) {
		analyzer := NewSemanticSimilarityAnalyzer()

		f1 := &CFGFeatures{
			BlockCount:       5,
			EdgeCount:        6,
			CyclomaticNumber: 3,
			EdgeTypeCounts: map[EdgeType]int{
				EdgeNormal:    3,
				EdgeCondTrue:  1,
				EdgeCondFalse: 1,
			},
		}
		f2 := &CFGFeatures{
			BlockCount:       5,
			EdgeCount:        6,
			CyclomaticNumber: 3,
			EdgeTypeCounts: map[EdgeType]int{
				EdgeNormal:    3,
				EdgeCondTrue:  1,
				EdgeCondFalse: 1,
			},
		}

		similarity := analyzer.compareCFGFeatures(f1, f2)
		assert.Equal(t, 1.0, similarity)
	})

	t.Run("DifferentFeatures", func(t *testing.T) {
		analyzer := NewSemanticSimilarityAnalyzer()

		f1 := &CFGFeatures{
			BlockCount:       5,
			EdgeCount:        6,
			CyclomaticNumber: 3,
			EdgeTypeCounts: map[EdgeType]int{
				EdgeNormal:    3,
				EdgeCondTrue:  1,
				EdgeCondFalse: 1,
			},
		}
		f2 := &CFGFeatures{
			BlockCount:       10,
			EdgeCount:        15,
			CyclomaticNumber: 7,
			EdgeTypeCounts: map[EdgeType]int{
				EdgeNormal: 10,
				EdgeLoop:   2,
			},
		}

		similarity := analyzer.compareCFGFeatures(f1, f2)
		assert.Less(t, similarity, 0.8)
	})

	t.Run("NilFeatures", func(t *testing.T) {
		analyzer := NewSemanticSimilarityAnalyzer()

		f1 := &CFGFeatures{BlockCount: 5}
		similarity := analyzer.compareCFGFeatures(nil, f1)
		assert.Equal(t, 0.0, similarity)

		similarity = analyzer.compareCFGFeatures(f1, nil)
		assert.Equal(t, 0.0, similarity)
	})

	t.Run("EmptyFeatures", func(t *testing.T) {
		analyzer := NewSemanticSimilarityAnalyzer()

		f1 := &CFGFeatures{BlockCount: 0, EdgeTypeCounts: map[EdgeType]int{}}
		f2 := &CFGFeatures{BlockCount: 0, EdgeTypeCounts: map[EdgeType]int{}}

		similarity := analyzer.compareCFGFeatures(f1, f2)
		assert.Equal(t, 1.0, similarity)
	})
}
