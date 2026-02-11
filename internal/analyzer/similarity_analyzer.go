package analyzer

// SimilarityAnalyzer defines the interface for computing similarity between code fragments.
// Each clone type should have its own analyzer implementation.
type SimilarityAnalyzer interface {
	// ComputeSimilarity returns a similarity score between 0.0 and 1.0
	ComputeSimilarity(f1, f2 *CodeFragment, calc *TFIDFCalculator) float64

	// ComputeDistance returns a distance score (0.0 for identical, higher for different)
	ComputeDistance(f1, f2 *CodeFragment) float64

	// GetName returns the name of this analyzer
	GetName() string
}

// CloneClassifier orchestrates multi-dimensional clone classification.
// It uses different analyzers for each clone type and applies a cascading
// classification approach from fastest (Type-1) to slowest (Type-4).
type CloneClassifier struct {
	textualAnalyzer    SimilarityAnalyzer // Type-1: exact text comparison
	syntacticAnalyzer  SimilarityAnalyzer // Type-2: normalized AST comparison
	structuralAnalyzer SimilarityAnalyzer // Type-3: AST with edit distance
	semanticAnalyzer   SimilarityAnalyzer // Type-4: control flow analysis

	// Thresholds for each clone type
	type1Threshold float64
	type2Threshold float64
	type3Threshold float64
	type4Threshold float64

	// Configuration
	enableTextualAnalysis  bool
	enableSemanticAnalysis bool
}

// CloneClassifierConfig holds configuration for the clone classifier
type CloneClassifierConfig struct {
	Type1Threshold         float64
	Type2Threshold         float64
	Type3Threshold         float64
	Type4Threshold         float64
	EnableTextualAnalysis  bool
	EnableSemanticAnalysis bool
	EnableDFAAnalysis      bool // Enable Data Flow Analysis for enhanced Type-4 detection
}

// NewCloneClassifier creates a new multi-dimensional clone classifier
func NewCloneClassifier(config *CloneClassifierConfig) *CloneClassifier {
	classifier := &CloneClassifier{
		type1Threshold:         config.Type1Threshold,
		type2Threshold:         config.Type2Threshold,
		type3Threshold:         config.Type3Threshold,
		type4Threshold:         config.Type4Threshold,
		enableTextualAnalysis:  config.EnableTextualAnalysis,
		enableSemanticAnalysis: config.EnableSemanticAnalysis,
	}

	// Initialize analyzers
	// Type-1: Textual similarity (only if enabled due to memory overhead)
	if config.EnableTextualAnalysis {
		classifier.textualAnalyzer = NewTextualSimilarityAnalyzer()
	}

	// Type-2: Syntactic similarity (uses APTED with identifier/literal normalization)
	classifier.syntacticAnalyzer = NewSyntacticSimilarityAnalyzer()

	// Type-3: Structural similarity (standard APTED)
	classifier.structuralAnalyzer = NewStructuralSimilarityAnalyzer()

	// Type-4: Semantic similarity (CFG-based, only if enabled)
	if config.EnableSemanticAnalysis {
		if config.EnableDFAAnalysis {
			// Use DFA-enhanced semantic analyzer
			classifier.semanticAnalyzer = NewSemanticSimilarityAnalyzerWithDFA()
		} else {
			classifier.semanticAnalyzer = NewSemanticSimilarityAnalyzer()
		}
	}

	return classifier
}

// ClassificationResult holds the result of clone classification
type ClassificationResult struct {
	CloneType  CloneType
	Similarity float64
	Confidence float64
	Analyzer   string
}

// ClassifyClone determines the clone type using cascading analysis.
// It returns the clone type, similarity score, and confidence.
// Classification order: Type-1 (fastest) -> Type-2 -> Type-3 -> Type-4 (slowest)
func (c *CloneClassifier) ClassifyClone(f1, f2 *CodeFragment, calc *TFIDFCalculator) *ClassificationResult {
	// Early validation
	if f1 == nil || f2 == nil {
		return nil
	}

	// Step 1: Type-1 check (textual comparison - fastest)
	if c.textualAnalyzer != nil && c.enableTextualAnalysis {
		textualSim := c.textualAnalyzer.ComputeSimilarity(f1, f2, calc)
		if textualSim >= c.type1Threshold {
			return &ClassificationResult{
				CloneType:  Type1Clone,
				Similarity: textualSim,
				Confidence: 1.0, // Highest confidence for exact textual match
				Analyzer:   c.textualAnalyzer.GetName(),
			}
		}
	}

	// Step 2: Type-2 check (syntactic with normalization)
	if c.syntacticAnalyzer != nil {
		syntacticSim := c.syntacticAnalyzer.ComputeSimilarity(f1, f2, calc)
		if syntacticSim >= c.type2Threshold {
			return &ClassificationResult{
				CloneType:  Type2Clone,
				Similarity: syntacticSim,
				Confidence: 0.95,
				Analyzer:   c.syntacticAnalyzer.GetName(),
			}
		}
	}

	// Step 3: Type-3 check (structural APTED)
	// Cache the structural similarity for potential reuse in fallback
	var structuralSim float64
	if c.structuralAnalyzer != nil {
		structuralSim = c.structuralAnalyzer.ComputeSimilarity(f1, f2, calc)
		if structuralSim >= c.type3Threshold {
			return &ClassificationResult{
				CloneType:  Type3Clone,
				Similarity: structuralSim,
				Confidence: 0.9,
				Analyzer:   c.structuralAnalyzer.GetName(),
			}
		}
	}

	// Step 4: Type-4 check (semantic/CFG - slowest)
	if c.semanticAnalyzer != nil && c.enableSemanticAnalysis {
		semanticSim := c.semanticAnalyzer.ComputeSimilarity(f1, f2, calc)
		if semanticSim >= c.type4Threshold {
			return &ClassificationResult{
				CloneType:  Type4Clone,
				Similarity: semanticSim,
				Confidence: 0.85,
				Analyzer:   c.semanticAnalyzer.GetName(),
			}
		}
	}

	// Fallback: check if cached structural similarity meets Type-4 threshold
	// This maintains backward compatibility with single-metric classification
	if c.structuralAnalyzer != nil && structuralSim >= c.type4Threshold {
		return &ClassificationResult{
			CloneType:  Type4Clone,
			Similarity: structuralSim,
			Confidence: 0.8,
			Analyzer:   c.structuralAnalyzer.GetName(),
		}
	}

	return nil // Not a clone
}

// ClassifyCloneSimple is a simplified version that returns just CloneType, similarity, and confidence.
// This is for backward compatibility with existing code.
func (c *CloneClassifier) ClassifyCloneSimple(f1, f2 *CodeFragment) (CloneType, float64, float64) {
	result := c.ClassifyClone(f1, f2, nil)
	if result == nil {
		return 0, 0.0, 0.0
	}
	return result.CloneType, result.Similarity, result.Confidence
}

// SetTextualAnalyzer sets the textual similarity analyzer (for testing)
func (c *CloneClassifier) SetTextualAnalyzer(analyzer SimilarityAnalyzer) {
	c.textualAnalyzer = analyzer
}

// SetSyntacticAnalyzer sets the syntactic similarity analyzer (for testing)
func (c *CloneClassifier) SetSyntacticAnalyzer(analyzer SimilarityAnalyzer) {
	c.syntacticAnalyzer = analyzer
}

// SetStructuralAnalyzer sets the structural similarity analyzer (for testing)
func (c *CloneClassifier) SetStructuralAnalyzer(analyzer SimilarityAnalyzer) {
	c.structuralAnalyzer = analyzer
}

// SetSemanticAnalyzer sets the semantic similarity analyzer (for testing)
func (c *CloneClassifier) SetSemanticAnalyzer(analyzer SimilarityAnalyzer) {
	c.semanticAnalyzer = analyzer
}
