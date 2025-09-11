package analyzer

import (
	"crypto/md5"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/ludo-technologies/pyscn/internal/parser"
)

// FeatureExtractor defines the interface for extracting features from AST trees
type FeatureExtractor interface {
	// ExtractFeatures extracts a set of string features from an AST tree
	ExtractFeatures(ast *TreeNode) ([]string, error)
	
	// ExtractSubtreeHashes extracts hashes of subtrees up to a maximum height
	ExtractSubtreeHashes(ast *TreeNode, maxHeight int) []string
	
	// ExtractNodeSequences extracts k-gram sequences from pre-order traversal
	ExtractNodeSequences(ast *TreeNode, k int) []string
}

// ASTFeatureExtractor implements feature extraction for AST trees
type ASTFeatureExtractor struct {
	maxSubtreeHeight int    // Maximum height for subtree hashing
	kGramSize        int    // Size of k-grams for sequence features
	includeTypes     bool   // Include node types in features
	includeLiterals  bool   // Include literal values in features
	includeStructure bool   // Include structural patterns
}

// NewASTFeatureExtractor creates a new feature extractor with default settings
func NewASTFeatureExtractor() *ASTFeatureExtractor {
	return &ASTFeatureExtractor{
		maxSubtreeHeight: 3,
		kGramSize:        4,
		includeTypes:     true,
		includeLiterals:  false, // Default to false for better generalization
		includeStructure: true,
	}
}

// NewASTFeatureExtractorWithConfig creates a feature extractor with custom configuration
func NewASTFeatureExtractorWithConfig(maxHeight, kGramSize int, includeTypes, includeLiterals, includeStructure bool) *ASTFeatureExtractor {
	return &ASTFeatureExtractor{
		maxSubtreeHeight: maxHeight,
		kGramSize:        kGramSize,
		includeTypes:     includeTypes,
		includeLiterals:  includeLiterals,
		includeStructure: includeStructure,
	}
}

// ExtractFeatures extracts all features from an AST tree
func (fe *ASTFeatureExtractor) ExtractFeatures(ast *TreeNode) ([]string, error) {
	if ast == nil {
		return []string{}, nil
	}

	var features []string

	// Extract subtree hashes
	if fe.includeTypes {
		subtreeHashes := fe.ExtractSubtreeHashes(ast, fe.maxSubtreeHeight)
		for _, hash := range subtreeHashes {
			features = append(features, "subtree:"+hash)
		}
	}

	// Extract k-gram sequences
	if fe.includeTypes {
		sequences := fe.ExtractNodeSequences(ast, fe.kGramSize)
		for _, seq := range sequences {
			features = append(features, "kgram:"+seq)
		}
	}

	// Extract structural patterns
	if fe.includeStructure {
		patterns := fe.extractStructuralPatterns(ast)
		for _, pattern := range patterns {
			features = append(features, "pattern:"+pattern)
		}
	}

	// Extract node type distribution
	if fe.includeTypes {
		typeDistribution := fe.extractNodeTypeDistribution(ast)
		for nodeType, count := range typeDistribution {
			features = append(features, fmt.Sprintf("type_count:%s:%d", nodeType, count))
		}
	}

	// Extract literal features if enabled
	if fe.includeLiterals {
		literals := fe.extractLiterals(ast)
		for _, literal := range literals {
			features = append(features, "literal:"+literal)
		}
	}

	// Sort features for consistent ordering
	sort.Strings(features)

	return features, nil
}

// ExtractSubtreeHashes extracts hashes of subtrees up to maxHeight
func (fe *ASTFeatureExtractor) ExtractSubtreeHashes(ast *TreeNode, maxHeight int) []string {
	if ast == nil || maxHeight <= 0 {
		return []string{}
	}

	var hashes []string
	fe.extractSubtreeHashesRecursive(ast, maxHeight, &hashes)
	
	// Remove duplicates and sort
	uniqueHashes := make(map[string]bool)
	for _, hash := range hashes {
		uniqueHashes[hash] = true
	}
	
	result := make([]string, 0, len(uniqueHashes))
	for hash := range uniqueHashes {
		result = append(result, hash)
	}
	sort.Strings(result)
	
	return result
}

// extractSubtreeHashesRecursive recursively extracts subtree hashes
func (fe *ASTFeatureExtractor) extractSubtreeHashesRecursive(node *TreeNode, maxHeight int, hashes *[]string) {
	if node == nil || maxHeight <= 0 {
		return
	}

	// Generate hash for current subtree
	subtreeHash := fe.computeSubtreeHash(node, maxHeight)
	*hashes = append(*hashes, subtreeHash)

	// Recursively process children
	for _, child := range node.Children {
		fe.extractSubtreeHashesRecursive(child, maxHeight-1, hashes)
	}
}

// computeSubtreeHash computes a hash for a subtree
func (fe *ASTFeatureExtractor) computeSubtreeHash(node *TreeNode, maxDepth int) string {
	if node == nil || maxDepth <= 0 {
		return ""
	}

	// Build a string representation of the subtree
	var builder strings.Builder
	fe.buildSubtreeString(node, maxDepth, &builder)
	
	// Compute MD5 hash
	data := []byte(builder.String())
	hash := md5.Sum(data)
	return fmt.Sprintf("%x", hash)
}

// buildSubtreeString builds a string representation of a subtree
func (fe *ASTFeatureExtractor) buildSubtreeString(node *TreeNode, maxDepth int, builder *strings.Builder) {
	if node == nil || maxDepth <= 0 {
		return
	}

	// Add node label
	builder.WriteString(node.Label)
	
	// Add children count for structural information
	builder.WriteString(fmt.Sprintf("(%d)", len(node.Children)))

	// Process children
	for i, child := range node.Children {
		if i > 0 {
			builder.WriteString(",")
		}
		builder.WriteString("[")
		fe.buildSubtreeString(child, maxDepth-1, builder)
		builder.WriteString("]")
	}
}

// ExtractNodeSequences extracts k-gram sequences from pre-order traversal
func (fe *ASTFeatureExtractor) ExtractNodeSequences(ast *TreeNode, k int) []string {
	if ast == nil || k <= 0 {
		return []string{}
	}

	// Get pre-order traversal of node labels
	var nodeSequence []string
	fe.preOrderTraversal(ast, &nodeSequence)

	// Extract k-grams
	var kgrams []string
	for i := 0; i <= len(nodeSequence)-k; i++ {
		kgram := strings.Join(nodeSequence[i:i+k], "_")
		kgrams = append(kgrams, kgram)
	}

	// Remove duplicates and sort
	uniqueKgrams := make(map[string]bool)
	for _, kgram := range kgrams {
		uniqueKgrams[kgram] = true
	}
	
	result := make([]string, 0, len(uniqueKgrams))
	for kgram := range uniqueKgrams {
		result = append(result, kgram)
	}
	sort.Strings(result)
	
	return result
}

// preOrderTraversal performs pre-order traversal and collects node labels
func (fe *ASTFeatureExtractor) preOrderTraversal(node *TreeNode, sequence *[]string) {
	if node == nil {
		return
	}

	// Visit current node
	*sequence = append(*sequence, fe.normalizeLabel(node.Label))

	// Visit children
	for _, child := range node.Children {
		fe.preOrderTraversal(child, sequence)
	}
}

// normalizeLabel normalizes node labels for better matching
func (fe *ASTFeatureExtractor) normalizeLabel(label string) string {
	// Convert to lowercase for case-insensitive matching
	normalized := strings.ToLower(label)
	
	// Remove common prefixes/suffixes that add noise
	normalized = strings.TrimPrefix(normalized, "node")
	normalized = strings.TrimSuffix(normalized, "node")
	
	return normalized
}

// extractStructuralPatterns extracts common structural patterns
func (fe *ASTFeatureExtractor) extractStructuralPatterns(ast *TreeNode) []string {
	var patterns []string
	
	// Count different types of control structures
	controlStructures := fe.countControlStructures(ast)
	for structure, count := range controlStructures {
		patterns = append(patterns, fmt.Sprintf("control:%s:%d", structure, count))
	}
	
	// Extract depth patterns
	depth := ast.Height()
	patterns = append(patterns, fmt.Sprintf("depth:%d", depth))
	
	// Extract branching factor patterns
	avgBranching := fe.computeAverageBranchingFactor(ast)
	patterns = append(patterns, fmt.Sprintf("avg_branching:%.2f", avgBranching))
	
	return patterns
}

// countControlStructures counts different types of control structures
func (fe *ASTFeatureExtractor) countControlStructures(node *TreeNode) map[string]int {
	counts := make(map[string]int)
	fe.countControlStructuresRecursive(node, counts)
	return counts
}

// countControlStructuresRecursive recursively counts control structures
func (fe *ASTFeatureExtractor) countControlStructuresRecursive(node *TreeNode, counts map[string]int) {
	if node == nil {
		return
	}

	// Check if this node represents a control structure
	label := strings.ToLower(node.Label)
	switch {
	case strings.Contains(label, "if"):
		counts["if"]++
	case strings.Contains(label, "for"):
		counts["for"]++
	case strings.Contains(label, "while"):
		counts["while"]++
	case strings.Contains(label, "try"):
		counts["try"]++
	case strings.Contains(label, "function"):
		counts["function"]++
	case strings.Contains(label, "class"):
		counts["class"]++
	}

	// Recursively process children
	for _, child := range node.Children {
		fe.countControlStructuresRecursive(child, counts)
	}
}

// computeAverageBranchingFactor computes the average branching factor
func (fe *ASTFeatureExtractor) computeAverageBranchingFactor(node *TreeNode) float64 {
	totalNodes := 0
	totalChildren := 0
	fe.computeBranchingFactorRecursive(node, &totalNodes, &totalChildren)
	
	if totalNodes == 0 {
		return 0.0
	}
	return float64(totalChildren) / float64(totalNodes)
}

// computeBranchingFactorRecursive recursively computes branching factor statistics
func (fe *ASTFeatureExtractor) computeBranchingFactorRecursive(node *TreeNode, totalNodes, totalChildren *int) {
	if node == nil {
		return
	}

	*totalNodes++
	*totalChildren += len(node.Children)

	for _, child := range node.Children {
		fe.computeBranchingFactorRecursive(child, totalNodes, totalChildren)
	}
}

// extractNodeTypeDistribution extracts the distribution of node types
func (fe *ASTFeatureExtractor) extractNodeTypeDistribution(node *TreeNode) map[string]int {
	distribution := make(map[string]int)
	fe.extractNodeTypeDistributionRecursive(node, distribution)
	return distribution
}

// extractNodeTypeDistributionRecursive recursively extracts node type distribution
func (fe *ASTFeatureExtractor) extractNodeTypeDistributionRecursive(node *TreeNode, distribution map[string]int) {
	if node == nil {
		return
	}

	// Count this node's type
	normalizedLabel := fe.normalizeLabel(node.Label)
	distribution[normalizedLabel]++

	// Recursively process children
	for _, child := range node.Children {
		fe.extractNodeTypeDistributionRecursive(child, distribution)
	}
}

// extractLiterals extracts literal values from the AST
func (fe *ASTFeatureExtractor) extractLiterals(node *TreeNode) []string {
	var literals []string
	fe.extractLiteralsRecursive(node, &literals)
	
	// Remove duplicates and sort
	uniqueLiterals := make(map[string]bool)
	for _, literal := range literals {
		uniqueLiterals[literal] = true
	}
	
	result := make([]string, 0, len(uniqueLiterals))
	for literal := range uniqueLiterals {
		result = append(result, literal)
	}
	sort.Strings(result)
	
	return result
}

// extractLiteralsRecursive recursively extracts literals
func (fe *ASTFeatureExtractor) extractLiteralsRecursive(node *TreeNode, literals *[]string) {
	if node == nil {
		return
	}

	// Check if this node contains a literal value
	if node.OriginalNode != nil {
		literal := fe.extractLiteralFromNode(node.OriginalNode)
		if literal != "" {
			*literals = append(*literals, literal)
		}
	}

	// Recursively process children
	for _, child := range node.Children {
		fe.extractLiteralsRecursive(child, literals)
	}
}

// extractLiteralFromNode extracts literal value from a parser node
func (fe *ASTFeatureExtractor) extractLiteralFromNode(node *parser.Node) string {
	if node == nil {
		return ""
	}

	// Handle different types of literal nodes
	switch node.Type {
	case parser.NodeConstant:
		if node.Value != nil {
			return fe.normalizeLiteral(fmt.Sprintf("%v", node.Value))
		}
	case parser.NodeName:
		if node.Name != "" {
			return "name:" + node.Name
		}
	}

	return ""
}

// normalizeLiteral normalizes literal values
func (fe *ASTFeatureExtractor) normalizeLiteral(literal string) string {
	// For strings, extract only the type information to avoid overfitting
	if strings.HasPrefix(literal, "\"") || strings.HasPrefix(literal, "'") {
		return "string_literal"
	}
	
	// For numbers, check if it's an integer or float
	if _, err := strconv.Atoi(literal); err == nil {
		return "int_literal"
	}
	if _, err := strconv.ParseFloat(literal, 64); err == nil {
		return "float_literal"
	}
	
	// For other literals, return as-is but normalized
	return strings.ToLower(literal)
}