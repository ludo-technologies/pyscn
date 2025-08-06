package parser

import (
	"context"
	"fmt"
	"io"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
)

// Parser provides Python code parsing capabilities using tree-sitter
type Parser struct {
	parser *sitter.Parser
}

// New creates a new Parser instance with Python grammar
func New() *Parser {
	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())
	return &Parser{
		parser: parser,
	}
}

// ParseResult represents the result of parsing Python code
type ParseResult struct {
	Tree       *sitter.Tree
	RootNode   *sitter.Node
	SourceCode []byte
}

// Parse parses Python source code and returns the AST
func (p *Parser) Parse(ctx context.Context, source []byte) (*ParseResult, error) {
	tree, err := p.parser.ParseCtx(ctx, nil, source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source: %w", err)
	}

	rootNode := tree.RootNode()
	if rootNode.HasError() {
		return nil, fmt.Errorf("syntax errors found in source code")
	}

	return &ParseResult{
		Tree:       tree,
		RootNode:   rootNode,
		SourceCode: source,
	}, nil
}

// ParseFile parses a Python file from a reader
func (p *Parser) ParseFile(ctx context.Context, reader io.Reader) (*ParseResult, error) {
	source, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read source: %w", err)
	}

	return p.Parse(ctx, source)
}

// GetNodeText returns the text content of a node
func (p *Parser) GetNodeText(node *sitter.Node, source []byte) string {
	return node.Content(source)
}

// WalkTree traverses the AST and calls the visitor function for each node
func (p *Parser) WalkTree(node *sitter.Node, visitor func(*sitter.Node) error) error {
	if err := visitor(node); err != nil {
		return err
	}

	childCount := int(node.ChildCount())
	for i := 0; i < childCount; i++ {
		child := node.Child(i)
		if err := p.WalkTree(child, visitor); err != nil {
			return err
		}
	}

	return nil
}

// FindNodes finds all nodes of a specific type in the tree
func (p *Parser) FindNodes(node *sitter.Node, nodeType string) []*sitter.Node {
	var nodes []*sitter.Node

	_ = p.WalkTree(node, func(n *sitter.Node) error {
		if n.Type() == nodeType {
			nodes = append(nodes, n)
		}
		return nil
	})

	return nodes
}

// HasSyntaxErrors checks if the parsed tree contains any syntax errors
func (p *Parser) HasSyntaxErrors(node *sitter.Node) bool {
	hasError := false

	_ = p.WalkTree(node, func(n *sitter.Node) error {
		if n.IsError() || n.IsMissing() {
			hasError = true
		}
		return nil
	})

	return hasError
}
