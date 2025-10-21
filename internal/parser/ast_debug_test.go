package parser

import (
	"context"
	"strings"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
)

// TestDebugTreeSitterFinally dumps the tree-sitter AST structure
// to understand how finally clauses are represented
func TestDebugTreeSitterFinally(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "SimpleTryFinally",
			source: `
try:
    print("try")
finally:
    print("finally")
`,
		},
		{
			name: "TryExceptFinally",
			source: `
try:
    risky()
except ValueError:
    handle()
finally:
    cleanup()
`,
		},
		{
			name: "TryExceptElseFinally",
			source: `
try:
    operation()
except Exception:
    handle()
else:
    success()
finally:
    cleanup()
`,
		},
	}

	parser := New()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := []byte(tt.source)
			result, err := parser.Parse(ctx, source)
			if err != nil {
				t.Fatalf("Parse() error: %v", err)
			}

			t.Logf("\n=== Tree-sitter AST for %s ===", tt.name)
			dumpTreeSitterNode(t, result.RootNode, source, 0)

			// Find try_statement nodes
			tryNodes := parser.FindNodes(result.RootNode, "try_statement")
			t.Logf("\n=== Found %d try_statement nodes ===", len(tryNodes))

			for i, tryNode := range tryNodes {
				t.Logf("\n--- Try Statement #%d ---", i+1)
				t.Logf("Type: %s", tryNode.Type())
				t.Logf("Children count: %d", tryNode.ChildCount())

				// Enumerate all children
				for j := 0; j < int(tryNode.ChildCount()); j++ {
					child := tryNode.Child(j)
					if child != nil {
						t.Logf("  Child %d: type='%s', named=%v, text='%s'",
							j,
							child.Type(),
							child.IsNamed(),
							truncateText(child.Content(source), 30),
						)
					}
				}

				// Check for named fields
				t.Logf("\n--- Try Statement Fields ---")
				fieldNames := []string{"body", "alternative", "handlers", "finally"}
				for _, fieldName := range fieldNames {
					fieldNode := tryNode.ChildByFieldName(fieldName)
					if fieldNode != nil {
						t.Logf("  Field '%s': type='%s', text='%s'",
							fieldName,
							fieldNode.Type(),
							truncateText(fieldNode.Content(source), 30),
						)
					} else {
						t.Logf("  Field '%s': <nil>", fieldName)
					}
				}

				// Look for finally-related child nodes
				t.Logf("\n--- Looking for finally-related nodes ---")
				for j := 0; j < int(tryNode.ChildCount()); j++ {
					child := tryNode.Child(j)
					if child != nil && strings.Contains(strings.ToLower(child.Type()), "finally") {
						t.Logf("  Found: type='%s' at child index %d", child.Type(), j)
						dumpTreeSitterNode(t, child, source, 2)
					}
				}
			}
		})
	}
}

// dumpTreeSitterNode recursively dumps a tree-sitter node structure
func dumpTreeSitterNode(t *testing.T, node *sitter.Node, source []byte, depth int) {
	if node == nil {
		return
	}

	indent := strings.Repeat("  ", depth)
	named := ""
	if !node.IsNamed() {
		named = " (anonymous)"
	}

	text := truncateText(node.Content(source), 40)
	t.Logf("%s%s%s: %q", indent, node.Type(), named, text)

	// Only recurse to reasonable depth to avoid too much output
	if depth < 3 {
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child != nil {
				dumpTreeSitterNode(t, child, source, depth+1)
			}
		}
	} else if node.ChildCount() > 0 {
		t.Logf("%s  ... (%d children)", indent, node.ChildCount())
	}
}

// truncateText truncates text to maxLen characters
func truncateText(text string, maxLen int) string {
	text = strings.ReplaceAll(text, "\n", "\\n")
	text = strings.ReplaceAll(text, "\t", "\\t")
	if len(text) > maxLen {
		return text[:maxLen] + "..."
	}
	return text
}

// TestCompareExceptAndFinally compares how except and finally clauses are structured
func TestCompareExceptAndFinally(t *testing.T) {
	source := []byte(`
try:
    operation()
except ValueError:
    handle_value()
except TypeError:
    handle_type()
finally:
    cleanup()
`)

	parser := New()
	ctx := context.Background()
	result, err := parser.Parse(ctx, source)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	tryNodes := parser.FindNodes(result.RootNode, "try_statement")
	if len(tryNodes) == 0 {
		t.Fatal("No try_statement found")
	}

	tryNode := tryNodes[0]
	t.Logf("\n=== Comparing except_clause and finally_clause ===")

	// Find except_clause nodes
	exceptCount := 0
	finallyCount := 0

	for i := 0; i < int(tryNode.ChildCount()); i++ {
		child := tryNode.Child(i)
		if child == nil {
			continue
		}

		switch child.Type() {
		case "except_clause":
			exceptCount++
			t.Logf("\nexcept_clause #%d:", exceptCount)
			t.Logf("  Type: %s", child.Type())
			t.Logf("  Has 'body' field: %v", child.ChildByFieldName("body") != nil)
			if bodyNode := child.ChildByFieldName("body"); bodyNode != nil {
				t.Logf("  Body type: %s", bodyNode.Type())
			}
			t.Logf("  Child count: %d", child.ChildCount())

		case "finally_clause":
			finallyCount++
			t.Logf("\nfinally_clause #%d:", finallyCount)
			t.Logf("  Type: %s", child.Type())
			t.Logf("  Has 'body' field: %v", child.ChildByFieldName("body") != nil)
			if bodyNode := child.ChildByFieldName("body"); bodyNode != nil {
				t.Logf("  Body type: %s", bodyNode.Type())
			}
			t.Logf("  Child count: %d", child.ChildCount())
		}
	}

	t.Logf("\n=== Summary ===")
	t.Logf("except_clause count: %d", exceptCount)
	t.Logf("finally_clause count: %d", finallyCount)

	if finallyCount == 0 {
		t.Errorf("❌ No finally_clause found! This is the root cause of the bug.")
		t.Logf("\nAll children of try_statement:")
		for i := 0; i < int(tryNode.ChildCount()); i++ {
			child := tryNode.Child(i)
			if child != nil {
				t.Logf("  [%d] %s", i, child.Type())
			}
		}
	} else {
		t.Logf("✅ finally_clause nodes found")
	}
}
