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
