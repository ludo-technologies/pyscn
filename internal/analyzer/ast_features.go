package analyzer

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strings"
)

// FeatureExtractor converts AST trees into feature sets for Jaccard similarity
type FeatureExtractor interface {
	ExtractFeatures(ast *TreeNode) ([]string, error)
	ExtractSubtreeHashes(ast *TreeNode, maxHeight int) []string
	ExtractNodeSequences(ast *TreeNode, k int) []string
}

// ASTFeatureExtractor implements FeatureExtractor for TreeNode
type ASTFeatureExtractor struct {
	maxSubtreeHeight int  // Default: 3
	kGramSize        int  // Default: 4
	includeTypes     bool // Include node types in features
	includeLiterals  bool // Include literal values
}

// NewASTFeatureExtractor creates a feature extractor with sensible defaults
func NewASTFeatureExtractor() *ASTFeatureExtractor {
	return &ASTFeatureExtractor{
		maxSubtreeHeight: 3,
		kGramSize:        4,
		includeTypes:     true,
		includeLiterals:  false,
	}
}

// WithOptions allows overriding defaults
func (a *ASTFeatureExtractor) WithOptions(maxHeight, k int, includeTypes, includeLiterals bool) *ASTFeatureExtractor {
	if maxHeight > 0 {
		a.maxSubtreeHeight = maxHeight
	}
	if k > 0 {
		a.kGramSize = k
	}
	a.includeTypes = includeTypes
	a.includeLiterals = includeLiterals
	return a
}

// ExtractFeatures builds a mixed set of features from the tree
// - Subtree hashes (bottom-up) up to maxSubtreeHeight
// - k-grams from pre-order traversal labels
// - Node type presence and lightweight distribution markers
// - Structural pattern tokens
func (a *ASTFeatureExtractor) ExtractFeatures(ast *TreeNode) ([]string, error) {
	if ast == nil {
		return []string{}, nil
	}

	// Collect components
	features := make(map[string]struct{})

	// Subtree hashes
	for _, f := range a.ExtractSubtreeHashes(ast, a.maxSubtreeHeight) {
		features[f] = struct{}{}
	}

	// k-grams
	for _, f := range a.ExtractNodeSequences(ast, a.kGramSize) {
		features["kgram:"+f] = struct{}{}
	}

	// Node type distribution and presence
	typeCounts := make(map[string]int)
	preorder := a.preorderLabels(ast)
	for _, lbl := range preorder {
		base := a.baseType(lbl)
		if a.includeTypes && base != "" {
			typeCounts[base]++
			features["type:"+base] = struct{}{}
		}
	}
	// Add binned counts to keep set nature while encoding rough frequencies
	for t, c := range typeCounts {
		bin := a.binCount(c)
		features[fmt.Sprintf("typedist:%s:%s", t, bin)] = struct{}{}
	}

	// Structural patterns
	for _, p := range a.extractPatterns(ast) {
		features["pattern:"+p] = struct{}{}
	}

	// Flatten to slice in stable order for repeatability
	out := make([]string, 0, len(features))
	for f := range features {
		out = append(out, f)
	}
	sort.Strings(out)
	return out, nil
}

// ExtractSubtreeHashes computes bottom-up hashes of subtrees up to maxHeight
func (a *ASTFeatureExtractor) ExtractSubtreeHashes(ast *TreeNode, maxHeight int) []string {
	if ast == nil {
		return []string{}
	}
	var feats []string
	var dfs func(n *TreeNode) (uint64, int)
	dfs = func(n *TreeNode) (uint64, int) {
		if n == nil {
			return 0, -1
		}
		// Compute children's hashes and heights
		childHashes := make([]uint64, 0, len(n.Children))
		maxH := -1
		for _, ch := range n.Children {
			h, height := dfs(ch)
			childHashes = append(childHashes, h)
			if height > maxH {
				maxH = height
			}
		}
		height := maxH + 1
		// Combine label and ordered child hashes
		h := fnv.New64a()
		_, _ = h.Write([]byte(a.canonicalLabel(n.Label)))
		for _, ch := range childHashes { // preserve order
			var b [8]byte
			// write uint64 big endian
			for i := 0; i < 8; i++ {
				b[7-i] = byte(ch >> (uint(8 * i)))
			}
			_, _ = h.Write(b[:])
		}
		hv := h.Sum64()
		if height <= maxHeight {
			feats = append(feats, fmt.Sprintf("sub:%d:%016x", height, hv))
		}
		return hv, height
	}
	_, _ = dfs(ast)
	return feats
}

// ExtractNodeSequences returns k-grams from pre-order traversal labels
func (a *ASTFeatureExtractor) ExtractNodeSequences(ast *TreeNode, k int) []string {
	if ast == nil || k <= 1 {
		return []string{}
	}
	labels := a.preorderLabels(ast)
	if len(labels) < k {
		return []string{}
	}
	grams := make([]string, 0, len(labels)-k+1)
	for i := 0; i <= len(labels)-k; i++ {
		grams = append(grams, strings.Join(labels[i:i+k], ":"))
	}
	return grams
}

// Helpers

func (a *ASTFeatureExtractor) canonicalLabel(lbl string) string {
	// Include or strip literal/value details
	if a.includeLiterals {
		return lbl
	}
	// Strip anything after first '(' to drop literal/name payloads
	if idx := strings.IndexByte(lbl, '('); idx >= 0 {
		return lbl[:idx]
	}
	return lbl
}

func (a *ASTFeatureExtractor) baseType(lbl string) string {
	// Base type is canonical label without payload
	s := a.canonicalLabel(lbl)
	// If canonical label is empty, skip
	return s
}

func (a *ASTFeatureExtractor) preorderLabels(ast *TreeNode) []string {
	labels := []string{}
	var walk func(n *TreeNode)
	walk = func(n *TreeNode) {
		if n == nil {
			return
		}
		labels = append(labels, a.canonicalLabel(n.Label))
		for _, ch := range n.Children {
			walk(ch)
		}
	}
	walk(ast)
	return labels
}

func (a *ASTFeatureExtractor) binCount(c int) string {
	switch {
	case c <= 1:
		return "1"
	case c <= 3:
		return "2-3"
	case c <= 7:
		return "4-7"
	case c <= 15:
		return "8-15"
	default:
		return "16+"
	}
}

func (a *ASTFeatureExtractor) extractPatterns(ast *TreeNode) []string {
	// Simple structural pattern tokens based on presence of common constructs
	// Patterns derived from base labels
	counts := make(map[string]int)
	var walk func(n *TreeNode)
	walk = func(n *TreeNode) {
		if n == nil {
			return
		}
		b := a.baseType(n.Label)
		counts[b]++
		for _, ch := range n.Children {
			walk(ch)
		}
	}
	walk(ast)
	pats := []string{}
	addIf := func(name string) {
		if counts[name] > 0 {
			pats = append(pats, name)
		}
	}
	addIf("If")
	addIf("For")
	addIf("While")
	addIf("Try")
	addIf("With")
	addIf("FunctionDef")
	addIf("ClassDef")
	addIf("Return")
	addIf("Assign")
	addIf("Call")
	addIf("Attribute")
	addIf("Compare")
	sort.Strings(pats)
	return pats
}
