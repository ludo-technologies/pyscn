package parser

// FixParentReferences traverses the AST and fixes any missing parent references
func FixParentReferences(node *Node) {
	if node == nil {
		return
	}
	
	// Fix direct children
	for _, child := range node.Children {
		if child != nil {
			if child.Parent == nil {
				child.Parent = node
			}
			FixParentReferences(child)
		}
	}
	
	// Fix body nodes
	for _, bodyNode := range node.Body {
		if bodyNode != nil {
			if bodyNode.Parent == nil {
				bodyNode.Parent = node
			}
			FixParentReferences(bodyNode)
		}
	}
	
	// Fix orelse nodes
	for _, orelseNode := range node.Orelse {
		if orelseNode != nil {
			if orelseNode.Parent == nil {
				orelseNode.Parent = node
			}
			FixParentReferences(orelseNode)
		}
	}
	
	// Fix finalbody nodes
	for _, finalNode := range node.Finalbody {
		if finalNode != nil {
			if finalNode.Parent == nil {
				finalNode.Parent = node
			}
			FixParentReferences(finalNode)
		}
	}
	
	// Fix handlers
	for _, handler := range node.Handlers {
		if handler != nil {
			if handler.Parent == nil {
				handler.Parent = node
			}
			FixParentReferences(handler)
		}
	}
	
	// Fix single node references
	if node.Test != nil {
		if node.Test.Parent == nil {
			node.Test.Parent = node
		}
		FixParentReferences(node.Test)
	}
	
	if node.Iter != nil {
		if node.Iter.Parent == nil {
			node.Iter.Parent = node
		}
		FixParentReferences(node.Iter)
	}
	
	if node.Left != nil {
		if node.Left.Parent == nil {
			node.Left.Parent = node
		}
		FixParentReferences(node.Left)
	}
	
	if node.Right != nil {
		if node.Right.Parent == nil {
			node.Right.Parent = node
		}
		FixParentReferences(node.Right)
	}
	
	// Fix targets
	for _, target := range node.Targets {
		if target != nil {
			if target.Parent == nil {
				target.Parent = node
			}
			FixParentReferences(target)
		}
	}
	
	// Fix args
	for _, arg := range node.Args {
		if arg != nil {
			if arg.Parent == nil {
				arg.Parent = node
			}
			FixParentReferences(arg)
		}
	}
	
	// Fix keywords
	for _, kw := range node.Keywords {
		if kw != nil {
			if kw.Parent == nil {
				kw.Parent = node
			}
			FixParentReferences(kw)
		}
	}
	
	// Fix decorators
	for _, dec := range node.Decorator {
		if dec != nil {
			if dec.Parent == nil {
				dec.Parent = node
			}
			FixParentReferences(dec)
		}
	}
	
	// Fix bases
	for _, base := range node.Bases {
		if base != nil {
			if base.Parent == nil {
				base.Parent = node
			}
			FixParentReferences(base)
		}
	}
	
	// Fix value node if it's a Node type
	if valNode, ok := node.Value.(*Node); ok && valNode != nil {
		if valNode.Parent == nil {
			valNode.Parent = node
		}
		FixParentReferences(valNode)
	}
}