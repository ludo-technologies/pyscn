package analyzer

import (
	"context"
	"sort"
	"strings"
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

func TestCloneDetector_DetectClonesWithLSH_CandidateCapDoesNotDropExactFamily(t *testing.T) {
	fragments := make([]*CodeFragment, 0, 9)
	for i := 0; i < 9; i++ {
		fragments = append(fragments, &CodeFragment{
			Location:  &CodeLocation{FilePath: "exact.py", StartLine: i*10 + 1, EndLine: i*10 + 5},
			TreeNode:  buildSimpleTree("FunctionDef", "If", "Return"),
			Size:      5,
			LineCount: 5,
		})
	}

	cfg := DefaultCloneDetectorConfig()
	cfg.UseLSH = true
	cfg.MinNodes = 1
	cfg.LSHSimilarityThreshold = 0.5
	cfg.LSHMaxCandidates = 8

	pairs, groups := NewCloneDetector(cfg).DetectClonesWithLSH(context.Background(), fragments)
	if len(pairs) == 0 || len(groups) == 0 {
		t.Fatalf("candidate cap dropped every exact clone candidate")
	}
}

func TestCloneDetector_DetectClonesWithLSH_MatchesStandardOnBenchmarkFixture(t *testing.T) {
	fragments := buildCloneBenchmarkFragments(4, 4, 8)

	standardDetector := NewCloneDetector(cloneBenchmarkConfig(false))
	standardPairs, standardGroups := standardDetector.DetectClones(fragments)
	if len(standardPairs) == 0 || len(standardGroups) == 0 {
		t.Fatalf("expected standard benchmark fixture run to produce pairs and groups")
	}

	lshDetector := NewCloneDetector(cloneBenchmarkConfig(true))
	lshPairs, lshGroups := lshDetector.DetectClonesWithLSH(context.Background(), fragments)
	if len(lshPairs) == 0 || len(lshGroups) == 0 {
		t.Fatalf("expected LSH benchmark fixture run to produce pairs and groups")
	}

	assertClonePairSnapshotsEqual(t, snapshotClonePairs(standardPairs), snapshotClonePairs(lshPairs))
	assertCloneGroupSnapshotsEqual(t, snapshotCloneGroups(standardGroups), snapshotCloneGroups(lshGroups))
}

type clonePairSnapshot struct {
	members    string
	similarity float64
	cloneType  CloneType
}

func snapshotClonePairs(pairs []*ClonePair) []clonePairSnapshot {
	snapshots := make([]clonePairSnapshot, 0, len(pairs))
	for _, pair := range pairs {
		members := []string{
			fragmentID(pair.Fragment1),
			fragmentID(pair.Fragment2),
		}
		sort.Strings(members)
		snapshots = append(snapshots, clonePairSnapshot{
			members:    strings.Join(members, "||"),
			similarity: pair.Similarity,
			cloneType:  pair.CloneType,
		})
	}

	sort.Slice(snapshots, func(i, j int) bool {
		if snapshots[i].members != snapshots[j].members {
			return snapshots[i].members < snapshots[j].members
		}
		if !almostEqual(snapshots[i].similarity, snapshots[j].similarity) {
			return snapshots[i].similarity < snapshots[j].similarity
		}
		return snapshots[i].cloneType < snapshots[j].cloneType
	})

	return snapshots
}

func assertClonePairSnapshotsEqual(t *testing.T, want, got []clonePairSnapshot) {
	t.Helper()

	if len(want) != len(got) {
		t.Fatalf("pair count mismatch: want %d got %d", len(want), len(got))
	}
	for i := range want {
		if want[i] != got[i] {
			t.Fatalf("pair mismatch at %d: want %+v got %+v", i, want[i], got[i])
		}
	}
}

func assertCloneGroupSnapshotsEqual(t *testing.T, want, got []cloneGroupSnapshot) {
	t.Helper()

	if len(want) != len(got) {
		t.Fatalf("group count mismatch: want %d got %d", len(want), len(got))
	}
	for i := range want {
		if want[i] != got[i] {
			t.Fatalf("group mismatch at %d: want %+v got %+v", i, want[i], got[i])
		}
	}
}
