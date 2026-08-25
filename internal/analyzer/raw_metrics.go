package analyzer

import (
	"bytes"
	"strings"

	"github.com/ludo-technologies/pyscn/internal/parser"
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

	docstringLines map[int]bool

	// slocPrefix[i] holds the number of source lines among the first i lines of
	// the file, so any line range can be measured without re-scanning content.
	slocPrefix []int
}

// Clone returns an independent raw-metrics result.
func (r *RawMetricsResult) Clone() *RawMetricsResult {
	if r == nil {
		return nil
	}
	cloned := *r
	cloned.docstringLines = make(map[int]bool, len(r.docstringLines))
	for line, included := range r.docstringLines {
		cloned.docstringLines[line] = included
	}
	cloned.slocPrefix = append([]int(nil), r.slocPrefix...)
	return &cloned
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
	slocPrefix := make([]int, len(lines)+1)

	for i, line := range lines {
		state.classifyLine(line, i, docstringLines, result)
		slocPrefix[i+1] = result.SLOC
	}
	result.docstringLines = docstringLines
	result.slocPrefix = slocPrefix

	denominator := result.SLOC + result.CommentLines
	if denominator > 0 {
		result.CommentRatio = float64(result.CommentLines) / float64(denominator)
	}

	return result
}

// CalculateAggregateRawMetrics aggregates raw code metrics across files.
func CalculateAggregateRawMetrics(results []*RawMetricsResult) *AggregateRawMetrics {
	aggregate := &AggregateRawMetrics{}

	for _, result := range results {
		if result == nil {
			continue
		}
		aggregate.FilesAnalyzed++
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

// PopulateLogicalLines updates raw metrics with LLOC derived from the parsed AST.
func PopulateLogicalLines(result *RawMetricsResult, ast *parser.Node) {
	if result == nil || ast == nil {
		return
	}

	result.LLOC = countLogicalBody(ast.Body, result.docstringLines, isDocstringBodyOwner(ast))
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

func countLogicalBody(nodes []*parser.Node, docstringLines map[int]bool, allowDocstring bool) int {
	total := 0
	docstringSkipped := false

	for _, node := range nodes {
		if node == nil || isLogicalSeparator(node) {
			continue
		}

		if allowDocstring && !docstringSkipped && isDocstringNode(node, docstringLines) {
			docstringSkipped = true
			continue
		}

		docstringSkipped = true
		total += countLogicalNode(node, docstringLines)
	}

	return total
}

func countLogicalNode(node *parser.Node, docstringLines map[int]bool) int {
	if node == nil || isLogicalSeparator(node) {
		return 0
	}

	if node.Type == parser.NodeElseClause || node.Type == parser.NodeBlock {
		return countLogicalBody(node.Body, docstringLines, false)
	}

	total := 1
	total += countLogicalBody(node.Body, docstringLines, isDocstringBodyOwner(node))
	total += countLogicalBody(node.Orelse, docstringLines, false)
	total += countLogicalBody(node.Finalbody, docstringLines, false)
	total += countLogicalBody(node.Handlers, docstringLines, false)

	return total
}

func isDocstringBodyOwner(node *parser.Node) bool {
	if node == nil {
		return false
	}

	switch node.Type {
	case parser.NodeModule, parser.NodeFunctionDef, parser.NodeAsyncFunctionDef, parser.NodeClassDef:
		return true
	default:
		return false
	}
}

func isDocstringNode(node *parser.Node, docstringLines map[int]bool) bool {
	if node == nil || node.Location.StartLine == 0 || len(docstringLines) == 0 {
		return false
	}

	return docstringLines[node.Location.StartLine-1]
}

func isLogicalSeparator(node *parser.Node) bool {
	return node != nil && string(node.Type) == ";"
}

// FunctionSLOC returns the source lines of code within the 1-indexed, inclusive
// line range [startLine, endLine], using the classification computed for the
// whole file. Comments, blank lines and docstrings are excluded, exactly as in
// the file-level SLOC metric.
//
// The range is measured verbatim, so lines belonging to definitions nested
// inside it (inner functions, classes) count toward the enclosing range as
// well: the value reflects the physical length of the definition.
func (r *RawMetricsResult) FunctionSLOC(startLine, endLine int) int {
	if r == nil || len(r.slocPrefix) == 0 || startLine <= 0 || startLine > endLine {
		return 0
	}

	lastLine := len(r.slocPrefix) - 1
	if startLine > lastLine {
		return 0
	}
	if endLine > lastLine {
		endLine = lastLine
	}

	return r.slocPrefix[endLine] - r.slocPrefix[startLine-1]
}
