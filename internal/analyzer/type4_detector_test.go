package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/internal/parser"
)

func TestType4CloneDetection(t *testing.T) {
	testDir := "../../testdata/python/clones/type4"

	// Read all Python files
	files, err := filepath.Glob(filepath.Join(testDir, "*.py"))
	if err != nil {
		t.Fatalf("Failed to glob files: %v", err)
	}

	// Create config with DFA enabled
	config := DefaultCloneDetectorConfig()
	config.EnableDFAAnalysis = true
	config.Type4Threshold = 0.70

	detector := NewCloneDetector(config)

	t.Logf("EnableMultiDimensionalAnalysis: %v", detector.cloneDetectorConfig.EnableMultiDimensionalAnalysis)
	t.Logf("EnableSemanticAnalysis: %v", detector.cloneDetectorConfig.EnableSemanticAnalysis)
	t.Logf("EnableDFAAnalysis: %v", detector.cloneDetectorConfig.EnableDFAAnalysis)
	t.Logf("Classifier is set: %v", detector.classifier != nil)

	p := parser.New()
	ctx := context.Background()

	// Extract fragments from all files
	var allFragments []*CodeFragment
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Logf("Failed to read %s: %v", file, err)
			continue
		}

		result, err := p.Parse(ctx, content)
		if err != nil {
			t.Logf("Failed to parse %s: %v", file, err)
			continue
		}

		fragments := detector.ExtractFragmentsWithSource([]*parser.Node{result.AST}, file, content)
		allFragments = append(allFragments, fragments...)
	}

	t.Logf("\nExtracted %d fragments:", len(allFragments))
	for i, f := range allFragments {
		t.Logf("  [%d] %s:%d-%d (nodes:%d, lines:%d)",
			i,
			filepath.Base(f.Location.FilePath),
			f.Location.StartLine,
			f.Location.EndLine,
			f.Size,
			f.LineCount)
	}

	// Run detection
	pairs, groups := detector.DetectClones(allFragments)

	t.Logf("\nDetected %d clone pairs, %d groups:", len(pairs), len(groups))
	for _, pair := range pairs {
		t.Logf("  %s:%d vs %s:%d = %.3f (Type-%d)",
			filepath.Base(pair.Fragment1.Location.FilePath),
			pair.Fragment1.Location.StartLine,
			filepath.Base(pair.Fragment2.Location.FilePath),
			pair.Fragment2.Location.StartLine,
			pair.Similarity,
			pair.CloneType)
	}

	// Test manual comparison of specific fragments
	if len(allFragments) >= 2 {
		t.Logf("\nManual comparison of first two fragments:")
		pair := detector.compareFragments(allFragments[0], allFragments[1])
		if pair != nil {
			t.Logf("  Result: sim=%.3f, type=%d", pair.Similarity, pair.CloneType)
		} else {
			t.Logf("  Result: nil (not a clone)")
		}
	}
}
