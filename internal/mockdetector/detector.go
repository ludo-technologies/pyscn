package mockdetector

import (
	"context"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// Detector analyzes Python code for mock data patterns.
type Detector struct {
	parser     *parser.Parser
	heuristics *Heuristics
}

// NewDetector creates a new Detector instance.
func NewDetector(keywords, domains []string) *Detector {
	return &Detector{
		parser:     parser.New(),
		heuristics: NewHeuristics(keywords, domains),
	}
}

// DetectResult represents the result of detecting mock data in a file.
type DetectResult struct {
	FilePath string
	Findings []domain.MockDataFinding
}

// Detect analyzes Python source code for mock data patterns.
func (d *Detector) Detect(ctx context.Context, source []byte, filePath string) (*DetectResult, error) {
	result, err := d.parser.Parse(ctx, source)
	if err != nil {
		return nil, err
	}

	findings := d.analyzeTree(result.RootNode, source)

	return &DetectResult{
		FilePath: filePath,
		Findings: findings,
	}, nil
}

// analyzeTree walks the AST and collects mock data findings.
func (d *Detector) analyzeTree(root *sitter.Node, source []byte) []domain.MockDataFinding {
	var findings []domain.MockDataFinding

	// Walk all nodes in the tree
	_ = d.walkTree(root, func(node *sitter.Node) error {
		nodeType := node.Type()

		switch nodeType {
		case "string":
			finding := d.analyzeString(node, source)
			if finding != nil {
				findings = append(findings, *finding)
			}

		case "identifier":
			finding := d.analyzeIdentifier(node, source)
			if finding != nil {
				findings = append(findings, *finding)
			}

		case "comment":
			finding := d.analyzeComment(node, source)
			if finding != nil {
				findings = append(findings, *finding)
			}

		case "assignment":
			// Check for mock-related variable assignments
			finding := d.analyzeAssignment(node, source)
			if finding != nil {
				findings = append(findings, *finding)
			}
		}

		return nil
	})

	return findings
}

// walkTree traverses the AST depth-first.
func (d *Detector) walkTree(node *sitter.Node, visitor func(*sitter.Node) error) error {
	if err := visitor(node); err != nil {
		return err
	}

	childCount := int(node.ChildCount())
	for i := 0; i < childCount; i++ {
		child := node.Child(i)
		if err := d.walkTree(child, visitor); err != nil {
			return err
		}
	}

	return nil
}

// analyzeString checks a string literal for mock data patterns.
func (d *Detector) analyzeString(node *sitter.Node, source []byte) *domain.MockDataFinding {
	content := node.Content(source)

	// Remove quotes from string content
	value := extractStringContent(content)
	if value == "" {
		return nil
	}

	matches := d.heuristics.CheckString(value)
	if len(matches) == 0 {
		return nil
	}

	// Return the highest severity match
	match := selectHighestSeverity(matches)

	return &domain.MockDataFinding{
		Location: domain.MockDataLocation{
			FilePath:    "",
			StartLine:   int(node.StartPoint().Row) + 1,
			EndLine:     int(node.EndPoint().Row) + 1,
			StartColumn: int(node.StartPoint().Column),
			EndColumn:   int(node.EndPoint().Column),
		},
		Value:       value,
		Type:        match.Type,
		Severity:    match.Severity,
		Description: match.Description,
		Rationale:   match.Rationale,
		Context:     content,
	}
}

// analyzeIdentifier checks an identifier for mock data patterns.
func (d *Detector) analyzeIdentifier(node *sitter.Node, source []byte) *domain.MockDataFinding {
	name := node.Content(source)

	matches := d.heuristics.CheckIdentifier(name)
	if len(matches) == 0 {
		return nil
	}

	match := matches[0]

	return &domain.MockDataFinding{
		Location: domain.MockDataLocation{
			FilePath:    "",
			StartLine:   int(node.StartPoint().Row) + 1,
			EndLine:     int(node.EndPoint().Row) + 1,
			StartColumn: int(node.StartPoint().Column),
			EndColumn:   int(node.EndPoint().Column),
		},
		Value:        name,
		Type:         match.Type,
		Severity:     match.Severity,
		Description:  match.Description,
		Rationale:    match.Rationale,
		VariableName: name,
	}
}

// analyzeComment checks a comment for placeholder markers.
func (d *Detector) analyzeComment(node *sitter.Node, source []byte) *domain.MockDataFinding {
	comment := node.Content(source)

	match := d.heuristics.CheckComment(comment)
	if match == nil {
		return nil
	}

	return &domain.MockDataFinding{
		Location: domain.MockDataLocation{
			FilePath:    "",
			StartLine:   int(node.StartPoint().Row) + 1,
			EndLine:     int(node.EndPoint().Row) + 1,
			StartColumn: int(node.StartPoint().Column),
			EndColumn:   int(node.EndPoint().Column),
		},
		Value:       comment,
		Type:        match.Type,
		Severity:    match.Severity,
		Description: match.Description,
		Rationale:   match.Rationale,
	}
}

// analyzeAssignment checks for mock-related variable assignments.
func (d *Detector) analyzeAssignment(node *sitter.Node, source []byte) *domain.MockDataFinding {
	// Check if this is an assignment with a mock-related name
	// and a string value that might contain mock data
	leftChild := node.ChildByFieldName("left")
	if leftChild == nil {
		return nil
	}

	// Get the variable name
	varName := leftChild.Content(source)

	// Check if the variable name suggests mock data
	identifierMatches := d.heuristics.CheckIdentifier(varName)
	if len(identifierMatches) == 0 {
		return nil
	}

	// Only flag assignments where both name and context suggest mock data
	rightChild := node.ChildByFieldName("right")
	if rightChild == nil {
		return nil
	}

	// Get the value being assigned
	value := rightChild.Content(source)

	// If the right side is a string, check it too
	if rightChild.Type() == "string" {
		stringContent := extractStringContent(value)
		stringMatches := d.heuristics.CheckString(stringContent)
		if len(stringMatches) > 0 {
			match := selectHighestSeverity(stringMatches)
			return &domain.MockDataFinding{
				Location: domain.MockDataLocation{
					FilePath:    "",
					StartLine:   int(node.StartPoint().Row) + 1,
					EndLine:     int(node.EndPoint().Row) + 1,
					StartColumn: int(node.StartPoint().Column),
					EndColumn:   int(node.EndPoint().Column),
				},
				Value:        value,
				Type:         match.Type,
				Severity:     match.Severity,
				Description:  "Mock data assignment detected",
				Rationale:    match.Rationale + " (assigned to mock variable: " + varName + ")",
				VariableName: varName,
			}
		}
	}

	return nil
}

// extractStringContent removes quotes from a Python string literal.
func extractStringContent(s string) string {
	if len(s) < 2 {
		return s
	}

	// Handle triple-quoted strings
	if len(s) >= 6 && (s[:3] == `"""` || s[:3] == `'''`) {
		return s[3 : len(s)-3]
	}

	// Handle f-strings
	if len(s) >= 2 && (s[0] == 'f' || s[0] == 'F') {
		if len(s) >= 3 {
			return extractStringContent(s[1:])
		}
		return s
	}

	// Handle regular strings (Python only uses single and double quotes)
	if s[0] == '"' || s[0] == '\'' {
		return s[1 : len(s)-1]
	}

	return s
}

// selectHighestSeverity returns the match with the highest severity.
func selectHighestSeverity(matches []Match) Match {
	if len(matches) == 0 {
		return Match{}
	}

	highest := matches[0]
	for _, m := range matches[1:] {
		if m.Severity.Level() > highest.Severity.Level() {
			highest = m
		}
	}
	return highest
}
