package analyzer

import (
	"hash/fnv"
	"regexp"
	"strings"
	"unicode"
)

// Precompiled regex for whitespace normalization (avoid recompilation on each call)
var whitespaceRegex = regexp.MustCompile(`\s+`)

// TextualSimilarityAnalyzer computes textual similarity for Type-1 clone detection.
// Type-1 clones are identical code fragments except for whitespace and comments.
type TextualSimilarityAnalyzer struct {
	// Configuration options
	normalizeWhitespace bool
	removeComments      bool
}

// TextualSimilarityConfig holds configuration for textual similarity analysis
type TextualSimilarityConfig struct {
	NormalizeWhitespace bool
	RemoveComments      bool
}

// NewTextualSimilarityAnalyzer creates a new textual similarity analyzer
// with default settings (normalize whitespace and remove comments).
func NewTextualSimilarityAnalyzer() *TextualSimilarityAnalyzer {
	return &TextualSimilarityAnalyzer{
		normalizeWhitespace: true,
		removeComments:      true,
	}
}

// NewTextualSimilarityAnalyzerWithConfig creates a textual similarity analyzer
// with custom configuration.
func NewTextualSimilarityAnalyzerWithConfig(config *TextualSimilarityConfig) *TextualSimilarityAnalyzer {
	return &TextualSimilarityAnalyzer{
		normalizeWhitespace: config.NormalizeWhitespace,
		removeComments:      config.RemoveComments,
	}
}

// ComputeSimilarity computes the textual similarity between two code fragments.
// Returns 1.0 for identical content (after normalization), or a Levenshtein-based
// similarity score for near-matches.
func (t *TextualSimilarityAnalyzer) ComputeSimilarity(f1, f2 *CodeFragment, _ *TFIDFCalculator) float64 {
	if f1 == nil || f2 == nil {
		return 0.0
	}

	// Get normalized content
	content1 := t.normalizeContent(f1.Content)
	content2 := t.normalizeContent(f2.Content)

	// Empty content check
	if content1 == "" && content2 == "" {
		return 1.0 // Both empty = identical
	}
	if content1 == "" || content2 == "" {
		return 0.0 // One empty = completely different
	}

	// Quick hash comparison for identical content
	if t.hashContent(content1) == t.hashContent(content2) {
		return 1.0
	}

	// If not identical, compute string similarity using Levenshtein distance
	return t.computeLevenshteinSimilarity(content1, content2)
}

// normalizeContent normalizes source code content for comparison.
// This removes comments and normalizes whitespace based on configuration.
func (t *TextualSimilarityAnalyzer) normalizeContent(content string) string {
	if content == "" {
		return ""
	}

	result := content

	// Remove Python comments if configured
	if t.removeComments {
		result = t.removePythonComments(result)
	}

	// Normalize whitespace if configured
	if t.normalizeWhitespace {
		result = t.normalizeWhitespaceInContent(result)
	}

	return result
}

// removePythonComments removes Python comments from source code
func (t *TextualSimilarityAnalyzer) removePythonComments(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inMultilineString := false
	stringDelimiter := ""

	for _, line := range lines {
		// Handle multi-line strings (''' or """)
		processedLine := t.processLineForComments(line, &inMultilineString, &stringDelimiter)
		if processedLine != "" || !inMultilineString {
			result = append(result, processedLine)
		}
	}

	return strings.Join(result, "\n")
}

// processLineForComments processes a single line, removing comments
func (t *TextualSimilarityAnalyzer) processLineForComments(line string, inMultilineString *bool, stringDelimiter *string) string {
	if *inMultilineString {
		// Look for end of multiline string
		if idx := strings.Index(line, *stringDelimiter); idx >= 0 {
			*inMultilineString = false
			return line[idx+len(*stringDelimiter):]
		}
		return "" // Inside multiline string, skip
	}

	// Check for multiline string start
	for _, delim := range []string{`"""`, `'''`} {
		if idx := strings.Index(line, delim); idx >= 0 {
			// Check if it's a docstring (starts at beginning after whitespace)
			trimmed := strings.TrimLeftFunc(line[:idx], unicode.IsSpace)
			if trimmed == "" {
				// Check if same line has closing delimiter
				rest := line[idx+3:]
				if endIdx := strings.Index(rest, delim); endIdx >= 0 {
					// Single line docstring - keep code after it
					return line[:idx] + rest[endIdx+3:]
				}
				*inMultilineString = true
				*stringDelimiter = delim
				return line[:idx]
			}
		}
	}

	// Remove single-line comments (# not inside string)
	result := t.removeLineComment(line)

	return result
}

// removeLineComment removes # comments from a line, respecting strings
func (t *TextualSimilarityAnalyzer) removeLineComment(line string) string {
	inString := false
	stringChar := rune(0)
	escaped := false

	for i, ch := range line {
		if escaped {
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			continue
		}

		if !inString {
			if ch == '"' || ch == '\'' {
				inString = true
				stringChar = ch
			} else if ch == '#' {
				return strings.TrimRightFunc(line[:i], unicode.IsSpace)
			}
		} else {
			if ch == stringChar {
				inString = false
			}
		}
	}

	return line
}

// normalizeWhitespaceInContent normalizes whitespace in content
func (t *TextualSimilarityAnalyzer) normalizeWhitespaceInContent(content string) string {
	// Replace multiple whitespace characters with single space (using precompiled regex)
	content = whitespaceRegex.ReplaceAllString(content, " ")

	// Trim leading and trailing whitespace
	content = strings.TrimSpace(content)

	return content
}

// hashContent computes a FNV-64 hash of the content for quick equality check
func (t *TextualSimilarityAnalyzer) hashContent(content string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(content))
	return h.Sum64()
}

// computeLevenshteinSimilarity computes similarity based on Levenshtein distance.
// Returns a value between 0.0 and 1.0.
func (t *TextualSimilarityAnalyzer) computeLevenshteinSimilarity(s1, s2 string) float64 {
	distance := t.levenshteinDistance(s1, s2)
	maxLen := maxInt(len(s1), len(s2))

	if maxLen == 0 {
		return 1.0
	}

	// Convert distance to similarity
	similarity := 1.0 - float64(distance)/float64(maxLen)
	if similarity < 0.0 {
		return 0.0
	}
	return similarity
}

// levenshteinDistance computes the Levenshtein edit distance between two strings.
// Uses dynamic programming with O(min(m,n)) space optimization.
func (t *TextualSimilarityAnalyzer) levenshteinDistance(s1, s2 string) int {
	// Ensure s1 is the shorter string for space optimization
	if len(s1) > len(s2) {
		s1, s2 = s2, s1
	}

	m := len(s1)
	n := len(s2)

	// Special cases
	if m == 0 {
		return n
	}
	if n == 0 {
		return m
	}

	// Use two rows for space optimization
	prev := make([]int, m+1)
	curr := make([]int, m+1)

	// Initialize first row
	for i := 0; i <= m; i++ {
		prev[i] = i
	}

	// Fill the matrix
	for j := 1; j <= n; j++ {
		curr[0] = j
		for i := 1; i <= m; i++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			curr[i] = min3(
				prev[i]+1,      // deletion
				curr[i-1]+1,    // insertion
				prev[i-1]+cost, // substitution
			)
		}
		prev, curr = curr, prev
	}

	return prev[m]
}

// GetName returns the name of this analyzer
func (t *TextualSimilarityAnalyzer) GetName() string {
	return "textual"
}

// min3 returns the minimum of three integers
func min3(a, b, c int) int {
	if a <= b && a <= c {
		return a
	}
	if b <= c {
		return b
	}
	return c
}

// maxInt returns the maximum of two integers
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
