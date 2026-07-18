package analyzer

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/ludo-technologies/pyscn/internal/parser"
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

func buildSimpleAST(types ...parser.NodeType) *parser.Node {
	if len(types) == 0 {
		return nil
	}
	root := parser.NewNode(types[0])
	cur := root
	for i := 1; i < len(types); i++ {
		child := parser.NewNode(types[i])
		cur.AddChild(child)
		cur = child
	}
	return root
}

func prepareTestFragment(t *testing.T, cfg *CloneDetectorConfig, fragment *CodeFragment) {
	t.Helper()
	detector := NewCloneDetector(cfg)
	detector.fragments = []*CodeFragment{fragment}
	detector.prepareFragments()
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
	result := det.DetectClonesWithLSH(context.Background(), []*CodeFragment{f1, f2, f3})
	pairs := result.Pairs
	if len(pairs) == 0 {
		// As a sanity check, verify MinHash similarity and LSH candidate retrieval
		ext := newPythonCloneFeatureExtractor()
		feats1, _ := ext.ExtractFeatures(toCoreTree(f1.TreeNode))
		feats2, _ := ext.ExtractFeatures(toCoreTree(f2.TreeNode))
		mh := NewMinHasher(128)
		s1 := mh.ComputeSignature(feats1)
		s2 := mh.ComputeSignature(feats2)
		est := mh.EstimateJaccardSimilarity(s1, s2)
		lsh := NewLSHIndex(32, 4)
		_ = lsh.AddFragment(0, s1)
		_ = lsh.AddFragment(1, s2)
		_ = lsh.BuildIndex()
		cands := lsh.FindCandidates(s1)
		t.Fatalf("expected at least one clone pair, got 0 (minhash est=%.3f, cands=%v)", est, cands)
	}
}

func TestCloneDetectorPrepareFragmentsBuildsTreeBackedFeatures(t *testing.T) {
	fragment := &CodeFragment{
		Location:  &CodeLocation{FilePath: "tree.py", StartLine: 1, EndLine: 5},
		TreeNode:  buildSimpleTree("FunctionDef", "If", "Return"),
		Size:      3,
		LineCount: 5,
	}

	prepareTestFragment(t, DefaultCloneDetectorConfig(), fragment)

	if len(fragment.Features) == 0 {
		t.Fatalf("expected prepareFragments to populate features for tree-backed fragment")
	}
}

func TestCloneDetectorPrepareFragmentsRefreshesTreeBackedFeatures(t *testing.T) {
	staleFeatures := []string{"stale:feature"}
	fragment := &CodeFragment{
		Location:  &CodeLocation{FilePath: "tree.py", StartLine: 1, EndLine: 5},
		TreeNode:  buildSimpleTree("FunctionDef", "If", "Return"),
		Size:      3,
		LineCount: 5,
		Features:  staleFeatures,
	}

	prepareTestFragment(t, DefaultCloneDetectorConfig(), fragment)

	if slices.Equal(fragment.Features, staleFeatures) {
		t.Fatalf("expected tree-backed fragment features to be refreshed")
	}
	expected, _ := newPythonCloneFeatureExtractor().ExtractFeatures(toCoreTree(fragment.TreeNode))
	if !slices.Equal(fragment.Features, expected) {
		t.Fatalf("expected tree-backed features %v, got %v", expected, fragment.Features)
	}
}

func TestCloneDetectorPrepareFragmentsRefreshesConvertedASTFeatures(t *testing.T) {
	staleFeatures := []string{"stale:feature"}
	fragment := &CodeFragment{
		Location:  &CodeLocation{FilePath: "ast.py", StartLine: 1, EndLine: 5},
		ASTNode:   buildSimpleAST(parser.NodeFunctionDef, parser.NodeIf, parser.NodeReturn),
		Size:      3,
		LineCount: 5,
		Features:  staleFeatures,
	}

	prepareTestFragment(t, DefaultCloneDetectorConfig(), fragment)

	if fragment.TreeNode == nil {
		t.Fatalf("expected prepareFragments to convert AST to tree")
	}
	if slices.Equal(fragment.Features, staleFeatures) {
		t.Fatalf("expected AST conversion to refresh stale features")
	}
	expected, _ := newPythonCloneFeatureExtractor().ExtractFeatures(toCoreTree(fragment.TreeNode))
	if !slices.Equal(fragment.Features, expected) {
		t.Fatalf("expected converted AST features %v, got %v", expected, fragment.Features)
	}
}

func TestCloneDetectorPrepareFragmentsFeatureContractIgnoresLSHRows(t *testing.T) {
	featuresForRows := func(rows int) []string {
		cfg := DefaultCloneDetectorConfig()
		cfg.LSHRows = rows
		fragment := &CodeFragment{
			Location:  &CodeLocation{FilePath: "rows.py", StartLine: 1, EndLine: 5},
			TreeNode:  buildSimpleTree("FunctionDef", "If", "Return"),
			Size:      3,
			LineCount: 5,
		}
		prepareTestFragment(t, cfg, fragment)
		return fragment.Features
	}

	rowsOneFeatures := featuresForRows(1)
	rowsNineFeatures := featuresForRows(9)
	if !slices.Equal(rowsOneFeatures, rowsNineFeatures) {
		t.Fatalf("LSHRows must not change clone features: rows=1 %v rows=9 %v", rowsOneFeatures, rowsNineFeatures)
	}
}

func TestCloneDetectorDetectClonesWithLSHRefreshesStaleFeatureCache(t *testing.T) {
	fragments := []*CodeFragment{
		{
			Location:  &CodeLocation{FilePath: "cached_a.py", StartLine: 1, EndLine: 5},
			TreeNode:  buildSimpleTree("FunctionDef", "Return", "Name"),
			Size:      3,
			LineCount: 5,
			Features:  []string{"stale:left"},
		},
		{
			Location:  &CodeLocation{FilePath: "cached_b.py", StartLine: 1, EndLine: 5},
			TreeNode:  buildSimpleTree("FunctionDef", "Return", "Name"),
			Size:      3,
			LineCount: 5,
			Features:  []string{"stale:right"},
		},
	}

	cfg := DefaultCloneDetectorConfig()
	cfg.UseLSH = true
	cfg.MinNodes = 1
	cfg.LSHSimilarityThreshold = 1

	hasher := NewMinHasher(cfg.LSHMinHashCount)
	staleSimilarity := hasher.EstimateJaccardSimilarity(
		hasher.ComputeSignature(fragments[0].Features),
		hasher.ComputeSignature(fragments[1].Features),
	)
	if staleSimilarity == 1 {
		t.Fatalf("test setup needs stale features below strict LSH threshold")
	}

	result := NewCloneDetector(cfg).DetectClonesWithLSH(context.Background(), fragments)
	pairs := result.Pairs
	if len(pairs) != 1 {
		t.Fatalf("expected LSH to refresh stale cached features before candidate generation, got %d pairs", len(pairs))
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

	result := NewCloneDetector(cfg).DetectClonesWithLSH(context.Background(), fragments)
	pairs, groups := result.Pairs, result.Groups
	if len(pairs) == 0 || len(groups) == 0 {
		t.Fatalf("candidate cap dropped every exact clone candidate")
	}
	assertGroupContainsAllFragments(t, groups, fragments)
}

func TestCloneDetector_DetectClonesWithLSH_MatchesStandardOnBenchmarkFixture(t *testing.T) {
	fragments := buildCloneBenchmarkFragments(4, 4, 8)

	standardDetector := NewCloneDetector(cloneBenchmarkConfig(false))
	standardResult := standardDetector.DetectClones(fragments)
	standardPairs, standardGroups := standardResult.Pairs, standardResult.Groups
	if len(standardPairs) == 0 || len(standardGroups) == 0 {
		t.Fatalf("expected standard benchmark fixture run to produce pairs and groups")
	}

	lshDetector := NewCloneDetector(cloneBenchmarkConfig(true))
	lshResult := lshDetector.DetectClonesWithLSH(context.Background(), fragments)
	lshPairs, lshGroups := lshResult.Pairs, lshResult.Groups
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

func assertGroupContainsAllFragments(t *testing.T, groups []*CloneGroup, fragments []*CodeFragment) {
	t.Helper()

	want := make(map[*CodeFragment]struct{}, len(fragments))
	for _, fragment := range fragments {
		want[fragment] = struct{}{}
	}

	for _, group := range groups {
		if len(group.Fragments) != len(fragments) {
			continue
		}
		got := make(map[*CodeFragment]struct{}, len(group.Fragments))
		for _, fragment := range group.Fragments {
			got[fragment] = struct{}{}
		}
		if len(got) != len(want) {
			continue
		}
		for fragment := range want {
			if _, ok := got[fragment]; !ok {
				t.Fatalf("group with all fragments is missing %s", fragmentID(fragment))
			}
		}
		return
	}

	t.Fatalf("expected one clone group to contain all %d exact fragments; got %d groups", len(fragments), len(groups))
}

// fragmentID returns a stable identifier for a fragment based on its location.
func fragmentID(f *CodeFragment) string {
	if f == nil || f.Location == nil {
		return fmt.Sprintf("%p", f)
	}
	loc := f.Location
	return fmt.Sprintf("%s|%d|%d|%d|%d", loc.FilePath, loc.StartLine, loc.EndLine, loc.StartCol, loc.EndCol)
}

func almostEqual(a, b float64) bool {
	const eps = 1e-9
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= eps
}

type cloneGroupSnapshot struct {
	members    string
	similarity float64
	cloneType  CloneType
}

func snapshotCloneGroups(groups []*CloneGroup) []cloneGroupSnapshot {
	snapshots := make([]cloneGroupSnapshot, 0, len(groups))
	for _, group := range groups {
		memberIDs := make([]string, 0, len(group.Fragments))
		for _, fragment := range group.Fragments {
			memberIDs = append(memberIDs, fragmentID(fragment))
		}
		sort.Strings(memberIDs)
		snapshots = append(snapshots, cloneGroupSnapshot{
			members:    strings.Join(memberIDs, "||"),
			similarity: group.Similarity,
			cloneType:  group.CloneType,
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
