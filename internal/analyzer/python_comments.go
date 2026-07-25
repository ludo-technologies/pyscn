package analyzer

import (
	"strings"
	"unicode"

	coreclone "github.com/ludo-technologies/polyscan/core/clone"
)

// removePythonComments strips Python comments (# line comments and docstring
// blocks) while preserving quoted content. It is injected into the core
// textual similarity analyzer as its language-specific CommentStripper.
func removePythonComments(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inMultilineString := false
	stringDelimiter := ""

	for _, line := range lines {
		// Handle multi-line strings (''' or """)
		processedLine := processLineForComments(line, &inMultilineString, &stringDelimiter)
		if processedLine != "" || !inMultilineString {
			result = append(result, processedLine)
		}
	}

	return strings.Join(result, "\n")
}

// processLineForComments processes a single line, removing comments
func processLineForComments(line string, inMultilineString *bool, stringDelimiter *string) string {
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
	return removeLineComment(line)
}

// removeLineComment removes # comments from a line, respecting strings
func removeLineComment(line string) string {
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

// fragmentHashNormalizer applies the default Type-1 normalization (remove
// comments, collapse whitespace) for fragment fingerprinting, independent of
// any per-detector similarity configuration.
var fragmentHashNormalizer = coreclone.NewTextualSimilarityAnalyzer(removePythonComments)
