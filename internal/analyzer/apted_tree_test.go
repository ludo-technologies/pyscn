package analyzer

import (
	"testing"

	"github.com/ludo-technologies/pyscn/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestNewTreeNode(t *testing.T) {
	node := NewTreeNode(1, "test")
	assert.Equal(t, 1, node.ID)
	assert.Equal(t, "test", node.Label)
	assert.Empty(t, node.Children)
}

func TestAddChild(t *testing.T) {
	parent := NewTreeNode(1, "parent")
	child := NewTreeNode(2, "child")
	parent.AddChild(child)

	assert.Len(t, parent.Children, 1)
	assert.Equal(t, parent, child.Parent)
}

func TestIsLeaf(t *testing.T) {
	leaf := NewTreeNode(1, "leaf")
	assert.True(t, leaf.IsLeaf())

	parent := NewTreeNode(1, "parent")
	parent.AddChild(NewTreeNode(2, "child"))
	assert.False(t, parent.IsLeaf())
}

func TestSize_SingleNode(t *testing.T) {
	node := NewTreeNode(1, "root")
	assert.Equal(t, 1, node.Size())
}

func TestSize_TreeWithChildren(t *testing.T) {
	//     root
	//    /    \
	//  child1  child2
	//    |
	//  grandchild
	root := NewTreeNode(1, "root")
	child1 := NewTreeNode(2, "child1")
	child2 := NewTreeNode(3, "child2")
	grandchild := NewTreeNode(4, "grandchild")

	root.AddChild(child1)
	root.AddChild(child2)
	child1.AddChild(grandchild)

	assert.Equal(t, 4, root.Size())
}

func TestSizeWithDepthLimit_apted(t *testing.T) {
	tests := []struct {
		name     string
		maxDepth int
		expected int
	}{
		{
			name:     "negative depth returns 1",
			maxDepth: -1,
			expected: 1,
		},
		{
			name:     "zero depth returns 1",
			maxDepth: 0,
			expected: 1,
		},
		{
			name:     "depth 1 includes direct children",
			maxDepth: 1,
			expected: 3, // root + child1 + child2
		},
		{
			name:     "depth 2 includes all nodes",
			maxDepth: 2,
			expected: 4, // root + child1 + child2 + grandchild
		},
		{
			name:     "large depth includes all nodes",
			maxDepth: 100,
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//     root
			//    /    \
			//  child1  child2
			//    |
			//  grandchild
			root := NewTreeNode(1, "root")
			child1 := NewTreeNode(2, "child1")
			child2 := NewTreeNode(3, "child2")
			grandchild := NewTreeNode(4, "grandchild")

			root.AddChild(child1)
			root.AddChild(child2)
			child1.AddChild(grandchild)

			assert.Equal(t, tt.expected, root.SizeWithDepthLimit(tt.maxDepth))
		})
	}
}

func TestHeight_Leaf(t *testing.T) {
	leaf := NewTreeNode(1, "leaf")
	assert.Equal(t, 0, leaf.Height())
}

func TestHeight_Tree(t *testing.T) {
	root := NewTreeNode(1, "root")
	child := NewTreeNode(2, "child")
	grandchild := NewTreeNode(3, "grandchild")

	root.AddChild(child)
	child.AddChild(grandchild)

	assert.Equal(t, 2, root.Height())
}

func TestHeightWithDepthLimit_apted(t *testing.T) {
	tests := []struct {
		name     string
		maxDepth int
		expected int
	}{
		{
			name:     "negative depth returns 0",
			maxDepth: -1,
			expected: 0,
		},
		{
			name:     "zero depth returns 0",
			maxDepth: 0,
			expected: 0,
		},
		{
			name:     "depth 1 includes direct child",
			maxDepth: 1,
			expected: 1,
		},
		{
			name:     "depth 2 includes all nodes",
			maxDepth: 2,
			expected: 2,
		},
		{
			name:     "large depth includes all nodes",
			maxDepth: 100,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := NewTreeNode(1, "root")
			child := NewTreeNode(2, "child")
			grandchild := NewTreeNode(3, "grandchild")

			root.AddChild(child)
			child.AddChild(grandchild)

			assert.Equal(t, tt.expected, root.HeightWithDepthLimit(tt.maxDepth))
		})
	}
}

func TestString_SingleNode(t *testing.T) {
	root := NewTreeNode(1, "root")
	assert.Equal(t, "Node{ID: 1, Label: root, Children: 0}", root.String())
}

func TestString_TreeWithChildren(t *testing.T) {
	//     root
	//    /    \
	//  child1  child2
	//    |
	//  grandchild
	root := NewTreeNode(1, "root")
	child1 := NewTreeNode(2, "child1")
	child2 := NewTreeNode(3, "child2")
	grandchild := NewTreeNode(4, "grandchild")

	root.AddChild(child1)
	root.AddChild(child2)
	child1.AddChild(grandchild)

	assert.Equal(t, "Node{ID: 1, Label: root, Children: 2}", root.String()) // Children = child1 + child2
}

func TestNewTreeConverter(t *testing.T) {
	converter := NewTreeConverter()
	assert.NotNil(t, converter)
}

func TestConvertAST(t *testing.T) {
	converter := NewTreeConverter()

	tests := []struct {
		name      string
		astNode   *parser.Node
		wantNil   bool
		wantLabel string
	}{
		{
			name:    "nil input returns nil",
			astNode: nil,
			wantNil: true,
		},
		{
			name:      "simple node",
			astNode:   &parser.Node{Type: parser.NodeName, Name: "x"},
			wantNil:   false,
			wantLabel: "Name(x)",
		},
		{
			name:      "constant node",
			astNode:   &parser.Node{Type: parser.NodeConstant, Value: 42},
			wantNil:   false,
			wantLabel: "Constant(42)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.ConvertAST(tt.astNode)

			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantLabel, result.Label)
			}
		})
	}
}

func TestPostOrderTraversal(t *testing.T) {
	//     root(2)
	//    /      \
	//  left(0)  right(1)
	root := NewTreeNode(1, "root")
	left := NewTreeNode(2, "left")
	right := NewTreeNode(3, "right")

	root.AddChild(left)
	root.AddChild(right)

	PostOrderTraversal(root)

	assert.Equal(t, 0, left.PostOrderID)
	assert.Equal(t, 1, right.PostOrderID)
	assert.Equal(t, 2, root.PostOrderID)
}

func TestComputeLeftMostLeaves(t *testing.T) {
	t.Run("nil tree", func(t *testing.T) {
		ComputeLeftMostLeaves(nil) // Should not panic
	})

	t.Run("single node", func(t *testing.T) {
		root := NewTreeNode(1, "root")
		PostOrderTraversal(root)
		ComputeLeftMostLeaves(root)

		assert.Equal(t, 0, root.LeftMostLeaf)
	})

	t.Run("tree with children", func(t *testing.T) {
		//       root
		//       /    \
		//   child1  child2
		//     |
		// grandchild
		root := NewTreeNode(1, "root")
		child1 := NewTreeNode(2, "child1")
		child2 := NewTreeNode(3, "child2")
		grandchild := NewTreeNode(4, "grandchild")

		root.AddChild(child1)
		root.AddChild(child2)
		child1.AddChild(grandchild)

		PostOrderTraversal(root)
		ComputeLeftMostLeaves(root)

		// grandchild(0) is left-most leaf of entire tree
		assert.Equal(t, 0, grandchild.LeftMostLeaf, "grandchild's left-most leaf is itself")
		assert.Equal(t, 0, child1.LeftMostLeaf, "child1's left-most leaf is grandchild")
		assert.Equal(t, 2, child2.LeftMostLeaf, "child2's left-most leaf is itself")
		assert.Equal(t, 0, root.LeftMostLeaf, "root's left-most leaf is grandchild")
	})
}

func TestAddChild_NilChild(t *testing.T) {
	parent := NewTreeNode(1, "parent")
	parent.AddChild(nil)
	assert.Empty(t, parent.Children)
}

func TestComputeKeyRoots(t *testing.T) {
	t.Run("nil tree", func(t *testing.T) {
		result := ComputeKeyRoots(nil)
		assert.Empty(t, result)
	})

	t.Run("single node", func(t *testing.T) {
		root := NewTreeNode(1, "root")
		PostOrderTraversal(root)
		ComputeLeftMostLeaves(root)

		keyRoots := ComputeKeyRoots(root)
		assert.Equal(t, []int{0}, keyRoots)
		assert.True(t, root.KeyRoot)
	})

	t.Run("tree with children", func(t *testing.T) {
		//       root(3)
		//       /    \
		//   child1(1) child2(2)
		//     |
		// grandchild(0)
		root := NewTreeNode(1, "root")
		child1 := NewTreeNode(2, "child1")
		child2 := NewTreeNode(3, "child2")
		grandchild := NewTreeNode(4, "grandchild")

		root.AddChild(child1)
		root.AddChild(child2)
		child1.AddChild(grandchild)

		PostOrderTraversal(root)
		ComputeLeftMostLeaves(root)
		keyRoots := ComputeKeyRoots(root)

		// Key roots are nodes whose left-most leaf hasn't been visited
		assert.Contains(t, keyRoots, 2, "child2 should be a key root")
		assert.Contains(t, keyRoots, 3, "root should be a key root")
		assert.Len(t, keyRoots, 2)
	})
}

func TestPrepareTreeForAPTED(t *testing.T) {
	t.Run("nil tree", func(t *testing.T) {
		keyRoots := PrepareTreeForAPTED(nil)
		assert.Empty(t, keyRoots)
	})

	t.Run("complete preparation", func(t *testing.T) {
		//       root
		//       /    \
		//   child1  child2
		//     |
		// grandchild
		root := NewTreeNode(1, "root")
		child1 := NewTreeNode(2, "child1")
		child2 := NewTreeNode(3, "child2")
		grandchild := NewTreeNode(4, "grandchild")

		root.AddChild(child1)
		root.AddChild(child2)
		child1.AddChild(grandchild)

		keyRoots := PrepareTreeForAPTED(root)

		// Verify PostOrderIDs were assigned
		assert.Equal(t, 0, grandchild.PostOrderID)
		assert.Equal(t, 1, child1.PostOrderID)
		assert.Equal(t, 2, child2.PostOrderID)
		assert.Equal(t, 3, root.PostOrderID)

		// Verify LeftMostLeaves were computed
		assert.Equal(t, 0, root.LeftMostLeaf)
		assert.Equal(t, 0, child1.LeftMostLeaf)
		assert.Equal(t, 2, child2.LeftMostLeaf)

		// Verify KeyRoots were identified
		assert.NotEmpty(t, keyRoots)
		assert.True(t, root.KeyRoot)
	})
}

func TestGetNodeByPostOrderID(t *testing.T) {
	t.Run("nil tree", func(t *testing.T) {
		result := GetNodeByPostOrderID(nil, 0)
		assert.Nil(t, result)
	})

	t.Run("find nodes by ID", func(t *testing.T) {
		root := NewTreeNode(1, "root")
		child1 := NewTreeNode(2, "child1")
		child2 := NewTreeNode(3, "child2")

		root.AddChild(child1)
		root.AddChild(child2)

		PostOrderTraversal(root)

		// Find each node by its PostOrderID
		found0 := GetNodeByPostOrderID(root, 0)
		assert.NotNil(t, found0)
		assert.Equal(t, "child1", found0.Label)

		found1 := GetNodeByPostOrderID(root, 1)
		assert.NotNil(t, found1)
		assert.Equal(t, "child2", found1.Label)

		found2 := GetNodeByPostOrderID(root, 2)
		assert.NotNil(t, found2)
		assert.Equal(t, "root", found2.Label)
	})

	t.Run("non-existent ID", func(t *testing.T) {
		root := NewTreeNode(1, "root")
		PostOrderTraversal(root)

		result := GetNodeByPostOrderID(root, 999)
		assert.Nil(t, result)
	})
}

func TestGetSubtreeNodes(t *testing.T) {
	t.Run("nil tree", func(t *testing.T) {
		result := GetSubtreeNodes(nil)
		assert.Empty(t, result)
	})

	t.Run("single node", func(t *testing.T) {
		root := NewTreeNode(1, "root")
		nodes := GetSubtreeNodes(root)

		assert.Len(t, nodes, 1)
		assert.Equal(t, root, nodes[0])
	})

	t.Run("tree with children", func(t *testing.T) {
		//       root
		//       /    \
		//   child1  child2
		//     |
		// grandchild
		root := NewTreeNode(1, "root")
		child1 := NewTreeNode(2, "child1")
		child2 := NewTreeNode(3, "child2")
		grandchild := NewTreeNode(4, "grandchild")

		root.AddChild(child1)
		root.AddChild(child2)
		child1.AddChild(grandchild)

		nodes := GetSubtreeNodes(root)

		assert.Len(t, nodes, 4)
		// Verify all nodes are included
		labels := make([]string, len(nodes))
		for i, node := range nodes {
			labels[i] = node.Label
		}
		assert.Contains(t, labels, "root")
		assert.Contains(t, labels, "child1")
		assert.Contains(t, labels, "child2")
		assert.Contains(t, labels, "grandchild")
	})

	t.Run("subtree from child node", func(t *testing.T) {
		root := NewTreeNode(1, "root")
		child1 := NewTreeNode(2, "child1")
		child2 := NewTreeNode(3, "child2")
		grandchild := NewTreeNode(4, "grandchild")

		root.AddChild(child1)
		root.AddChild(child2)
		child1.AddChild(grandchild)

		// Get subtree starting from child1
		nodes := GetSubtreeNodes(child1)

		assert.Len(t, nodes, 2) // child1 and grandchild
		labels := make([]string, len(nodes))
		for i, node := range nodes {
			labels[i] = node.Label
		}
		assert.Contains(t, labels, "child1")
		assert.Contains(t, labels, "grandchild")
		assert.NotContains(t, labels, "root")
		assert.NotContains(t, labels, "child2")
	})
}

func TestNewTreeConverterWithConfig(t *testing.T) {
	t.Run("default converter does not skip docstrings", func(t *testing.T) {
		converter := NewTreeConverter()
		assert.NotNil(t, converter)
		assert.False(t, converter.skipDocstrings)
	})

	t.Run("converter with skip docstrings enabled", func(t *testing.T) {
		converter := NewTreeConverterWithConfig(true)
		assert.NotNil(t, converter)
		assert.True(t, converter.skipDocstrings)
	})

	t.Run("converter with skip docstrings disabled", func(t *testing.T) {
		converter := NewTreeConverterWithConfig(false)
		assert.NotNil(t, converter)
		assert.False(t, converter.skipDocstrings)
	})
}

func TestTreeConverter_SkipDocstrings(t *testing.T) {
	t.Run("skips docstring when configured", func(t *testing.T) {
		// Create a function with a docstring
		// def test_function():
		//     """This is a docstring"""
		//     return 1
		docstringNode := &parser.Node{
			Type:     parser.NodeConstant,
			Value:    "This is a docstring",
			Children: []*parser.Node{},
		}
		exprNode := &parser.Node{
			Type:     parser.NodeExpr,
			Children: []*parser.Node{docstringNode},
		}
		returnNode := &parser.Node{
			Type:     parser.NodeReturn,
			Children: []*parser.Node{},
		}
		funcNode := &parser.Node{
			Type: parser.NodeFunctionDef,
			Name: "test_function",
			Body: []*parser.Node{exprNode, returnNode},
		}

		converterWithSkip := NewTreeConverterWithConfig(true)
		treeWithSkip := converterWithSkip.ConvertAST(funcNode)

		converterWithoutSkip := NewTreeConverterWithConfig(false)
		treeWithoutSkip := converterWithoutSkip.ConvertAST(funcNode)

		// Tree with skip should have fewer children (docstring excluded)
		assert.Less(t, treeWithSkip.Size(), treeWithoutSkip.Size())
	})

	t.Run("does not skip non-docstring expressions", func(t *testing.T) {
		// Create a function with a non-docstring expression first
		// def test_function():
		//     print("hello")  # This is a Call, not a Constant
		//     return 1
		callNode := &parser.Node{
			Type:     parser.NodeCall,
			Name:     "print",
			Children: []*parser.Node{},
		}
		exprNode := &parser.Node{
			Type:     parser.NodeExpr,
			Children: []*parser.Node{callNode},
		}
		returnNode := &parser.Node{
			Type:     parser.NodeReturn,
			Children: []*parser.Node{},
		}
		funcNode := &parser.Node{
			Type: parser.NodeFunctionDef,
			Name: "test_function",
			Body: []*parser.Node{exprNode, returnNode},
		}

		converterWithSkip := NewTreeConverterWithConfig(true)
		treeWithSkip := converterWithSkip.ConvertAST(funcNode)

		converterWithoutSkip := NewTreeConverterWithConfig(false)
		treeWithoutSkip := converterWithoutSkip.ConvertAST(funcNode)

		// Should be same size since the first expression is not a docstring
		assert.Equal(t, treeWithSkip.Size(), treeWithoutSkip.Size())
	})

	t.Run("does not skip docstring at non-first position", func(t *testing.T) {
		// Create a function with docstring at position 1 (not 0)
		// def test_function():
		//     x = 1
		//     """Not a docstring - appears after first statement"""
		assignNode := &parser.Node{
			Type:     parser.NodeAssign,
			Children: []*parser.Node{},
		}
		docstringNode := &parser.Node{
			Type:     parser.NodeConstant,
			Value:    "Not a docstring",
			Children: []*parser.Node{},
		}
		exprNode := &parser.Node{
			Type:     parser.NodeExpr,
			Children: []*parser.Node{docstringNode},
		}
		funcNode := &parser.Node{
			Type: parser.NodeFunctionDef,
			Name: "test_function",
			Body: []*parser.Node{assignNode, exprNode},
		}

		converterWithSkip := NewTreeConverterWithConfig(true)
		treeWithSkip := converterWithSkip.ConvertAST(funcNode)

		converterWithoutSkip := NewTreeConverterWithConfig(false)
		treeWithoutSkip := converterWithoutSkip.ConvertAST(funcNode)

		// Should be same size since the string is not at position 0
		assert.Equal(t, treeWithSkip.Size(), treeWithoutSkip.Size())
	})

	t.Run("handles integer constant at first position", func(t *testing.T) {
		// Create a function with integer constant at first position
		// def test_function():
		//     42  # Not a docstring - it's an integer
		//     return 1
		intNode := &parser.Node{
			Type:     parser.NodeConstant,
			Value:    int64(42), // integer, not string
			Children: []*parser.Node{},
		}
		exprNode := &parser.Node{
			Type:     parser.NodeExpr,
			Children: []*parser.Node{intNode},
		}
		returnNode := &parser.Node{
			Type:     parser.NodeReturn,
			Children: []*parser.Node{},
		}
		funcNode := &parser.Node{
			Type: parser.NodeFunctionDef,
			Name: "test_function",
			Body: []*parser.Node{exprNode, returnNode},
		}

		converterWithSkip := NewTreeConverterWithConfig(true)
		treeWithSkip := converterWithSkip.ConvertAST(funcNode)

		converterWithoutSkip := NewTreeConverterWithConfig(false)
		treeWithoutSkip := converterWithoutSkip.ConvertAST(funcNode)

		// Should be same size since the constant is not a string
		assert.Equal(t, treeWithSkip.Size(), treeWithoutSkip.Size())
	})
}
