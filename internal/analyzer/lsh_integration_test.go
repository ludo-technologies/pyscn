package analyzer

import (
	"context"
	"testing"
)

// buildSimpleTree builds a small ordered tree with labels
func buildSimpleTree(labels ...string) *TreeNode {
	if len(labels) == 0 {
		return nil
	}
	root := NewTreeNode(0, labels[0])
	cur := root
	for i := 1; i < len(labels); i++ {
		n := NewTreeNode(i, labels[i])
		cur.AddChild(n)
		cur = n
	}
	PrepareTreeForAPTED(root)
	return root
}

func TestCloneDetector_DetectClonesWithLSH_Simple(t *testing.T) {
	// Two similar trees; one dissimilar
	t1 := buildSimpleTree("FunctionDef", "If", "Return")
	t2 := buildSimpleTree("FunctionDef", "If", "Return")
	t3 := buildSimpleTree("ClassDef", "Assign", "Attribute")

	f1 := &CodeFragment{Location: &CodeLocation{FilePath: "A.py", StartLine: 1, EndLine: 10}, TreeNode: t1, Size: 5, LineCount: 5}
	f2 := &CodeFragment{Location: &CodeLocation{FilePath: "B.py", StartLine: 1, EndLine: 8}, TreeNode: t2, Size: 5, LineCount: 5}
	f3 := &CodeFragment{Location: &CodeLocation{FilePath: "C.py", StartLine: 1, EndLine: 6}, TreeNode: t3, Size: 5, LineCount: 5}

	cfg := DefaultCloneDetectorConfig()
	cfg.UseLSH = true
	cfg.MinNodes = 1
	cfg.LSHSimilarityThreshold = 0.2
	cfg.LSHBands = 32
	cfg.LSHRows = 4
	cfg.LSHMinHashCount = 128

	det := NewCloneDetector(cfg)
	pairs, _ := det.DetectClonesWithLSH(context.Background(), []*CodeFragment{f1, f2, f3})
	if len(pairs) == 0 {
		// As a sanity check, verify MinHash similarity and LSH candidate retrieval
		ext := NewASTFeatureExtractor()
		feats1, _ := ext.ExtractFeatures(f1.TreeNode)
		feats2, _ := ext.ExtractFeatures(f2.TreeNode)
		mh := NewMinHasher(128)
		s1 := mh.ComputeSignature(feats1)
		s2 := mh.ComputeSignature(feats2)
		est := mh.EstimateJaccardSimilarity(s1, s2)
		lsh := NewLSHIndex(32, 4)
		_ = lsh.AddFragment("A.py:1-10", s1)
		_ = lsh.AddFragment("B.py:1-8", s2)
		_ = lsh.BuildIndex()
		cands := lsh.FindCandidates(s1)
		t.Fatalf("expected at least one clone pair, got 0 (minhash est=%.3f, cands=%v)", est, cands)
	}
}
