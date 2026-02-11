package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTFIDFIntegration(t *testing.T) {
	p := parser.New()
	ctx := context.Background()

	// Boilerplate code (will appear frequent and get lower weights with TF-IDF)
	code1 := `@dataclass
class UserProfile:
    username: str
    email: str
    created_at: int = 0`

	code2 := `@dataclass
class ProductEntry:
    product_id: str
    sku: str
    price_cents: int = 0`

	// This code has identical boilerplate structure but different "domain" labels
	// Jaccard similarity (default) often gives high similarity because of shared structure labels.
	// TF-IDF should give lower similarity if those common labels are frequent.

	res1, err := p.Parse(ctx, []byte(code1))
	require.NoError(t, err)
	res2, err := p.Parse(ctx, []byte(code2))
	require.NoError(t, err)

	config := DefaultCloneDetectorConfig()
	config.MinLines = 1
	config.MinNodes = 1
	config.UseTFIDF = true // Enable TF-IDF

	detector := NewCloneDetector(config)

	// Simulate IDF computation by giving it a set of files
	// In a real run, DetectClones does this. We'll manually trigger the pre-computation phase
	// by providing some fragments.
	fragments1 := detector.ExtractFragments([]*parser.Node{res1.AST}, "test1.py")
	fragments2 := detector.ExtractFragments([]*parser.Node{res2.AST}, "test2.py")

	// Prepare trees (this is usually done in DetectClones)
	converter := NewTreeConverter()
	for _, f := range fragments1 {
		f.TreeNode = converter.ConvertAST(f.ASTNode)
		PrepareTreeForAPTED(f.TreeNode)
	}
	for _, f := range fragments2 {
		f.TreeNode = converter.ConvertAST(f.ASTNode)
		PrepareTreeForAPTED(f.TreeNode)
	}

	// Manually initialize TF-IDF calculator with these fragments
	allFragments := append(fragments1, fragments2...)
	detector.tfidfCalculator = NewTFIDFCalculator()
	detector.tfidfCalculator.ComputeIDF(allFragments)

	require.NotNil(t, detector.tfidfCalculator)

	// Now compare with TF-IDF enabled
	similarity := detector.analyzer.ComputeSimilarity(fragments1[0], fragments2[0], detector.tfidfCalculator)

	t.Logf("TF-IDF Similarity: %f", similarity)

	// Compare with no TF-IDF (effectively Jaccard)
	similarityNoTFIDF := detector.analyzer.ComputeSimilarity(fragments1[0], fragments2[0], nil)
	t.Logf("Jaccard Similarity: %f", similarityNoTFIDF)

	// TF-IDF should generally be different than Jaccard
	// In this specific case, if @dataclass and other nodes are common, similarity should go DOWN
	assert.NotEqual(t, similarityNoTFIDF, similarity, "TF-IDF similarity should differ from Jaccard")
}

func TestTFIDF_TrueClones(t *testing.T) {
	p := parser.New()
	ctx := context.Background()

	// Identical code but with different variable names (Type-2)
	code1 := `def calculate_total(prices):
    total = 0
    for price in prices:
        total += price
    return total`

	code2 := `def sum_all_items(values):
    s = 0
    for v in values:
        s += v
    return s`

	res1, err := p.Parse(ctx, []byte(code1))
	require.NoError(t, err)
	res2, err := p.Parse(ctx, []byte(code2))
	require.NoError(t, err)

	config := DefaultCloneDetectorConfig()
	config.MinLines = 1
	config.MinNodes = 1
	config.UseTFIDF = true
	detector := NewCloneDetector(config)

	fragments1 := detector.ExtractFragments([]*parser.Node{res1.AST}, "test1.py")
	fragments2 := detector.ExtractFragments([]*parser.Node{res2.AST}, "test2.py")

	converter := NewTreeConverter()
	for _, f := range fragments1 {
		f.TreeNode = converter.ConvertAST(f.ASTNode)
		PrepareTreeForAPTED(f.TreeNode)
	}
	for _, f := range fragments2 {
		f.TreeNode = converter.ConvertAST(f.ASTNode)
		PrepareTreeForAPTED(f.TreeNode)
	}

	// Manually initialize TF-IDF calculator
	allFragments := append(fragments1, fragments2...)
	detector.tfidfCalculator = NewTFIDFCalculator()
	detector.tfidfCalculator.ComputeIDF(allFragments)

	// Now compare
	similarity := detector.analyzer.ComputeSimilarity(fragments1[0], fragments2[0], detector.tfidfCalculator)

	t.Logf("True Clone TF-IDF Similarity: %f", similarity)

	// Since these are structurally identical and only variable names changed,
	// similarity should be very high (1.0 or very close to it depending on normalization)
	assert.Greater(t, similarity, 0.9, "Highly similar files should have high similarity")
}

func TestTFIDF_PartialSimilarity(t *testing.T) {
	p := parser.New()
	ctx := context.Background()

	// Two functions that share some logic (a loop summing things)
	// but have different surrounds and different details
	// Two functions that share some logic (a loop summing things)
	// but have different surrounds and different details
	code1 := `def sum_positive_numbers(arr):
    res = 0
    for x in arr:
        if x > 0:
            res += x
    return res`

	code2 := `def complex_process(data):
    s = 0
    for val in data:
        if val > 0:
            s += val
    try:
        save_to_db(s)
    except Exception:
        log_error()
    return s`

	res1, err := p.Parse(ctx, []byte(code1))
	require.NoError(t, err)
	res2, err := p.Parse(ctx, []byte(code2))
	require.NoError(t, err)

	config := DefaultCloneDetectorConfig()
	config.MinLines = 1
	config.MinNodes = 1
	config.UseTFIDF = true
	detector := NewCloneDetector(config)

	fragments1 := detector.ExtractFragments([]*parser.Node{res1.AST}, "test1.py")
	fragments2 := detector.ExtractFragments([]*parser.Node{res2.AST}, "test2.py")

	converter := NewTreeConverter()
	for _, f := range fragments1 {
		f.TreeNode = converter.ConvertAST(f.ASTNode)
		PrepareTreeForAPTED(f.TreeNode)
	}
	for _, f := range fragments2 {
		f.TreeNode = converter.ConvertAST(f.ASTNode)
		PrepareTreeForAPTED(f.TreeNode)
	}

	// Manually initialize TF-IDF calculator
	allFragments := append(fragments1, fragments2...)
	detector.tfidfCalculator = NewTFIDFCalculator()
	detector.tfidfCalculator.ComputeIDF(allFragments)

	// Now compare
	similarity := detector.analyzer.ComputeSimilarity(fragments1[0], fragments2[0], detector.tfidfCalculator)

	t.Logf("Partial Clone TF-IDF Similarity: %f", similarity)

	// Should be in (0, 1) range
	assert.True(t, similarity > 0.0 && similarity < 1.0, "Partial similarity should be between 0 and 1")
	assert.Greater(t, similarity, 0.4, "These should be fairly similar due to the shared loop/if structure")
}

func TestTFIDF_RealWorldAmbiguous(t *testing.T) {
	p := parser.New()
	ctx := context.Background()

	// Read real test files
	pathA := filepath.Join("..", "..", "testdata", "python", "clones", "type4", "find_max_a.py")
	pathB := filepath.Join("..", "..", "testdata", "python", "clones", "type4", "find_max_b.py")

	codeA, err := os.ReadFile(pathA)
	require.NoError(t, err)
	codeB, err := os.ReadFile(pathB)
	require.NoError(t, err)

	resA, err := p.Parse(ctx, codeA)
	require.NoError(t, err)
	resB, err := p.Parse(ctx, codeB)
	require.NoError(t, err)

	config := DefaultCloneDetectorConfig()
	config.MinLines = 3
	config.MinNodes = 5
	config.UseTFIDF = true
	detector := NewCloneDetector(config)

	fragmentsA := detector.ExtractFragments([]*parser.Node{resA.AST}, pathA)
	fragmentsB := detector.ExtractFragments([]*parser.Node{resB.AST}, pathB)

	require.NotEmpty(t, fragmentsA)
	require.NotEmpty(t, fragmentsB)

	converter := NewTreeConverter()
	for _, f := range append(fragmentsA, fragmentsB...) {
		f.TreeNode = converter.ConvertAST(f.ASTNode)
		PrepareTreeForAPTED(f.TreeNode)
	}

	// Manually initialize TF-IDF calculator
	allFragments := append(fragmentsA, fragmentsB...)
	detector.tfidfCalculator = NewTFIDFCalculator()
	detector.tfidfCalculator.ComputeIDF(allFragments)

	foundValidComparison := false
	for _, fA := range fragmentsA {
		// Only look at find_maximum or find_min_max
		nameA := fA.ASTNode.Name
		for _, fB := range fragmentsB {
			if fB.ASTNode.Name == nameA {
				similarity := detector.analyzer.ComputeSimilarity(fA, fB, detector.tfidfCalculator)
				t.Logf("Function %s Real-World Similarity: %f", nameA, similarity)

				// Ensure it's not exactly 0.5 or 1.0 or 0.0
				assert.True(t, similarity > 0.0 && similarity < 1.0)
				assert.NotEqual(t, 0.5, similarity)
				foundValidComparison = true
			}
		}
	}
	assert.True(t, foundValidComparison, "Should have found matching function names to compare")
}

func TestCloneDetector_UseTFIDF_Selection(t *testing.T) {
	config := DefaultCloneDetectorConfig()

	t.Run("Default_Analyzer_Is_APTED", func(t *testing.T) {
		config.UseTFIDF = false
		detector := NewCloneDetector(config)
		assert.Equal(t, "apted", detector.analyzer.GetName())
	})

	t.Run("TFIDF_Enabled_Analyzer_Is_Syntactic", func(t *testing.T) {
		config.UseTFIDF = true
		detector := NewCloneDetector(config)
		assert.Equal(t, "syntactic", detector.analyzer.GetName())
	})
}
