package analyzer

import (
	"bytes"
	"strings"
)

// RawMetricsResult contains file-level raw code metrics.
type RawMetricsResult struct {
	FilePath       string
	SLOC           int
	LLOC           int
	CommentLines   int
	DocstringLines int
	BlankLines     int
	TotalLines     int
	CommentRatio   float64
}

// AggregateRawMetrics contains aggregated raw code metrics across files.
type AggregateRawMetrics struct {
	FilesAnalyzed  int
	SLOC           int
	LLOC           int
	CommentLines   int
	DocstringLines int
	BlankLines     int
	TotalLines     int
	CommentRatio   float64
}

type rawStringMode int

const (
	rawStringModeNone rawStringMode = iota
	rawStringModeDocstring
	rawStringModeCode
)

type rawMetricsState struct {
	inMultilineString    bool
	multilineDelimiter   string
	multilineMode        rawStringMode
	moduleDocstringReady bool
	blockDocstringIndent *int
}

type logicalLineState struct {
	inString       bool
	inTripleString bool
	stringQuote    byte
	tripleQuote    string
	escaped        bool
	bracketDepth   int
	lineHasCode    bool
}

// CalculateRawMetrics calculates raw code metrics without requiring AST parsing.
func CalculateRawMetrics(content []byte, filePath string) *RawMetricsResult {
	lines := splitRawLines(content)
	result := &RawMetricsResult{
		FilePath:   filePath,
		TotalLines: len(lines),
	}
	if len(lines) == 0 {
		return result
	}

	state := rawMetricsState{moduleDocstringReady: true}
	docstringLines := make(map[int]bool, len(lines))

	for i, line := range lines {
		state.classifyLine(line, i, docstringLines, result)
	}

	result.LLOC = estimateLogicalLines(lines, docstringLines)

	denominator := result.SLOC + result.CommentLines
	if denominator > 0 {
		result.CommentRatio = float64(result.CommentLines) / float64(denominator)
	}

	return result
}

// CalculateAggregateRawMetrics aggregates raw code metrics across files.
func CalculateAggregateRawMetrics(results []*RawMetricsResult) *AggregateRawMetrics {
	aggregate := &AggregateRawMetrics{
		FilesAnalyzed: len(results),
	}

	for _, result := range results {
		if result == nil {
			continue
		}
		aggregate.SLOC += result.SLOC
		aggregate.LLOC += result.LLOC
		aggregate.CommentLines += result.CommentLines
		aggregate.DocstringLines += result.DocstringLines
		aggregate.BlankLines += result.BlankLines
		aggregate.TotalLines += result.TotalLines
	}

	denominator := aggregate.SLOC + aggregate.CommentLines
	if denominator > 0 {
		aggregate.CommentRatio = float64(aggregate.CommentLines) / float64(denominator)
	}

	return aggregate
}

func splitRawLines(content []byte) []string {
	if len(content) == 0 {
		return nil
	}

	normalized := bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
	normalized = bytes.ReplaceAll(normalized, []byte("\r"), []byte("\n"))

	lines := strings.Split(string(normalized), "\n")
	if len(lines) > 0 && normalized[len(normalized)-1] == '\n' {
		lines = lines[:len(lines)-1]
	}

	return lines
}

func (s *rawMetricsState) classifyLine(line string, lineIndex int, docstringLines map[int]bool, result *RawMetricsResult) {
	if s.inMultilineString {
		switch s.multilineMode {
		case rawStringModeDocstring:
			result.DocstringLines++
			docstringLines[lineIndex] = true
		case rawStringModeCode:
			result.SLOC++
		}

		if strings.Count(line, s.multilineDelimiter)%2 == 1 {
			s.inMultilineString = false
			s.multilineDelimiter = ""
			s.multilineMode = rawStringModeNone
		}
		return
	}

	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		result.BlankLines++
		return
	}

	if strings.HasPrefix(trimmed, "#") {
		result.CommentLines++
		return
	}

	indent := countLeadingIndent(line)
	if delimiter := s.docstringDelimiter(trimmed, indent); delimiter != "" {
		result.DocstringLines++
		docstringLines[lineIndex] = true
		s.moduleDocstringReady = false
		s.blockDocstringIndent = nil

		if strings.Count(line, delimiter)%2 == 1 {
			s.inMultilineString = true
			s.multilineDelimiter = delimiter
			s.multilineMode = rawStringModeDocstring
		}
		return
	}

	result.SLOC++
	s.clearDocstringExpectation()
	s.moduleDocstringReady = false

	if delimiter, ok := findTripleQuoteOutsideComments(line); ok && strings.Count(line, delimiter)%2 == 1 {
		s.inMultilineString = true
		s.multilineDelimiter = delimiter
		s.multilineMode = rawStringModeCode
	}

	if startsDocstringEligibleBlock(trimmed) {
		blockIndent := indent
		s.blockDocstringIndent = &blockIndent
	}
}

func (s *rawMetricsState) docstringDelimiter(trimmed string, indent int) string {
	delimiter := leadingTripleQuoteDelimiter(trimmed)
	if delimiter == "" {
		return ""
	}

	if s.moduleDocstringReady {
		return delimiter
	}

	if s.blockDocstringIndent != nil && indent > *s.blockDocstringIndent {
		return delimiter
	}

	return ""
}

func (s *rawMetricsState) clearDocstringExpectation() {
	s.blockDocstringIndent = nil
}

func countLeadingIndent(line string) int {
	count := 0
	for _, r := range line {
		if r != ' ' && r != '\t' {
			break
		}
		count++
	}
	return count
}

func startsDocstringEligibleBlock(trimmed string) bool {
	return (strings.HasPrefix(trimmed, "def ") ||
		strings.HasPrefix(trimmed, "async def ") ||
		strings.HasPrefix(trimmed, "class ")) &&
		strings.HasSuffix(trimmed, ":")
}

func leadingTripleQuoteDelimiter(trimmed string) string {
	for _, prefixLen := range []int{0, 1, 2} {
		if prefixLen > len(trimmed)-3 {
			continue
		}

		prefix := trimmed[:prefixLen]
		if !isValidStringPrefix(prefix) {
			continue
		}

		rest := trimmed[prefixLen:]
		switch {
		case strings.HasPrefix(rest, `"""`):
			return `"""`
		case strings.HasPrefix(rest, `'''`):
			return `'''`
		}
	}

	return ""
}

func isValidStringPrefix(prefix string) bool {
	switch prefix {
	case "", "r", "R", "u", "U":
		return true
	default:
		return false
	}
}

func findTripleQuoteOutsideComments(line string) (string, bool) {
	inSingle := false
	inDouble := false
	escaped := false

	for i := 0; i < len(line); i++ {
		ch := line[i]

		if escaped {
			escaped = false
			continue
		}

		if (inSingle || inDouble) && ch == '\\' {
			escaped = true
			continue
		}

		if !inSingle && !inDouble {
			switch {
			case strings.HasPrefix(line[i:], `"""`):
				return `"""`, true
			case strings.HasPrefix(line[i:], `'''`):
				return `'''`, true
			case ch == '#':
				return "", false
			case ch == '\'':
				inSingle = true
			case ch == '"':
				inDouble = true
			}
			continue
		}

		if inSingle && ch == '\'' {
			inSingle = false
		}
		if inDouble && ch == '"' {
			inDouble = false
		}
	}

	return "", false
}

func estimateLogicalLines(lines []string, docstringLines map[int]bool) int {
	state := logicalLineState{}
	statementCount := 0

	for i, line := range lines {
		if docstringLines[i] {
			continue
		}

		lastMeaningful := byte(0)
		for pos := 0; pos < len(line); {
			if state.inTripleString {
				index := strings.Index(line[pos:], state.tripleQuote)
				state.lineHasCode = true
				if index == -1 {
					pos = len(line)
					continue
				}
				pos += index + len(state.tripleQuote)
				state.inTripleString = false
				state.tripleQuote = ""
				continue
			}

			ch := line[pos]

			if state.escaped {
				state.escaped = false
				state.lineHasCode = true
				lastMeaningful = ch
				pos++
				continue
			}

			if state.inString {
				state.lineHasCode = true
				lastMeaningful = ch
				if ch == '\\' {
					state.escaped = true
					pos++
					continue
				}
				if ch == state.stringQuote {
					state.inString = false
				}
				pos++
				continue
			}

			switch {
			case strings.HasPrefix(line[pos:], `"""`):
				state.inTripleString = true
				state.tripleQuote = `"""`
				state.lineHasCode = true
				lastMeaningful = '"'
				pos += 3
				continue
			case strings.HasPrefix(line[pos:], `'''`):
				state.inTripleString = true
				state.tripleQuote = `'''`
				state.lineHasCode = true
				lastMeaningful = '\''
				pos += 3
				continue
			}

			if ch == '#' {
				break
			}

			if ch == '\'' || ch == '"' {
				state.inString = true
				state.stringQuote = ch
				state.lineHasCode = true
				lastMeaningful = ch
				pos++
				continue
			}

			if ch == ' ' || ch == '\t' {
				pos++
				continue
			}

			state.lineHasCode = true
			lastMeaningful = ch

			switch ch {
			case '(', '[', '{':
				state.bracketDepth++
			case ')', ']', '}':
				if state.bracketDepth > 0 {
					state.bracketDepth--
				}
			case ';':
				statementCount++
				state.lineHasCode = false
				lastMeaningful = 0
			}

			pos++
		}

		continued := state.inTripleString || state.inString || state.bracketDepth > 0 || lastMeaningful == '\\'
		if state.lineHasCode && !continued {
			statementCount++
			state.lineHasCode = false
		}
	}

	if state.lineHasCode {
		statementCount++
	}

	return statementCount
}
