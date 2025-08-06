// Package parser provides Python code parsing capabilities using tree-sitter.
//
// This package wraps the tree-sitter Go bindings to parse Python source code
// into Abstract Syntax Trees (AST). It provides utilities for traversing,
// querying, and analyzing Python code structure.
//
// Key features:
//   - Fast and accurate Python parsing using tree-sitter
//   - Support for Python 3.8+ syntax
//   - Error-tolerant parsing with syntax error detection
//   - Tree traversal and node searching utilities
//   - Cross-platform compatibility
//
// Basic usage:
//
//	p := parser.New()
//	result, err := p.Parse(ctx, []byte("def hello(): pass"))
//	if err != nil {
//	    // Handle parsing error
//	}
//	// Use result.RootNode to traverse the AST
package parser
