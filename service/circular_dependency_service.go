package service

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// CircularDependencyService provides an external call interface for detecting circular dependencies.
type CircularDependencyService struct {
	importPositions map[string]map[string]Position
}

// NewCircularDependencyService creates a new instance of CircularDependencyService.
func NewCircularDependencyService() *CircularDependencyService {
	return &CircularDependencyService{
		importPositions: make(map[string]map[string]Position), // file -> imported module -> position
	}
}

// GetImportPositions returns the map of import positions to be used by external functions for reporting
func (s *CircularDependencyService) GetImportPositions() map[string]map[string]Position {
	return s.importPositions
}

// Node represents a node (module or file) in the dependency graph.
// It tracks the name, edges to dependencies, visitation state for DFS, stack state, and parent node.
type Node struct {
	name    string
	edges   []*Node
	visited bool
	inStack bool
	parent  *Node
}

// Graph represents the dependency graph structure composed of nodes.
type Graph struct {
	nodes map[string]*Node
}

// NewGraph creates and initializes a new empty dependency graph.
func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[string]*Node),
	}
}

// AddEdge adds a directed edge in the graph from 'from' node to 'to' node,
// representing that 'from' depends on 'to'.
func (g *Graph) AddEdge(from, to string) {
	fromNode := g.getNode(from)
	toNode := g.getNode(to)
	fromNode.edges = append(fromNode.edges, toNode)
}

// getNode retrieves a node by name from the graph; if not present, creates a new node.
func (g *Graph) getNode(name string) *Node {
	if node, ok := g.nodes[name]; ok {
		return node
	}
	node := &Node{name: name}
	g.nodes[name] = node
	return node
}

type Position struct {
	Line   int
	Column int
}

// DetectCyclesWithStart performs cycle detection for the entire graph,
// returning all detected cycles as slices of node names.
// Cycles are rotated such that the cycle starts at 'desiredStart' node if present.
func (g *Graph) DetectCyclesWithStart(desiredStart string) [][]string {
	var cycles [][]string
	// Reset visitation and stack state
	for _, node := range g.nodes {
		node.visited = false
		node.inStack = false
		node.parent = nil
	}
	// Run DFS to find cycles starting from unvisited nodes
	for _, node := range g.nodes {
		if !node.visited {
			g.dfs(node, &cycles, desiredStart)
		}
	}
	return cycles
}

// dfs performs a Depth-First Search starting from the given node to detect all cycles in the graph.
// It keeps track of visited nodes and the nodes currently in the recursion stack (`inStack`) to identify cycles.
// Arguments:
// - node: the current node being visited
// - cycles: pointer to a slice collecting all detected cycles as slices of node names
// - desiredStart: a node name used to rotate the detected cycle so that it starts with this node (if present).
//
// Workflow:
//  1. Mark the current node as visited and mark it as being in the recursion stack.
//  2. For each neighbor of this node:
//     a. If the neighbor has not been visited, recursively call dfs on it, setting the current node as its parent.
//     b. If the neighbor is in the recursion stack (meaning a back edge found), we detected a cycle.
//     We backtrack from the current node up through its parents until reaching the neighbor,
//     collecting the sequence of node names that form the cycle.
//     c. Add the neighbor again at the end to close the cycle.
//     d. Rotate the cycle slice so that it starts with the desiredStart node, if it exists in the cycle.
//     e. Append this cycle to the list of detected cycles.
//  3. After exploring all neighbors, mark the current node as no longer in the recursion stack.
func (g *Graph) dfs(node *Node, cycles *[][]string, desiredStart string) {
	node.visited = true
	node.inStack = true
	// Explore neighbors (dependencies)
	for _, neighbor := range node.edges {
		if !neighbor.visited {
			neighbor.parent = node // track parent to reconstruct cycle
			g.dfs(neighbor, cycles, desiredStart)
		} else if neighbor.inStack {
			// Back edge found, cycle detected
			cycle := []string{neighbor.name}
			for current := node; current != nil && current != neighbor; current = current.parent {
				cycle = append(cycle, current.name)
			}
			// Close cycle by appending start node again
			cycle = append(cycle, neighbor.name)
			// Rotate cycle to start at desiredStart node if present
			rotatedCycle := rotateCycleToStart(cycle, desiredStart)
			*cycles = append(*cycles, rotatedCycle)
		}
	}
	node.inStack = false
}

// rotateCycleToStart rotates a cycle slice so that it starts at 'start' node if it exists; else returns original.
func rotateCycleToStart(cycle []string, start string) []string {
	var idx int = -1
	for i, v := range cycle {
		if v == start {
			idx = i
			break
		}
	}
	if idx == -1 {
		return cycle
	}
	n := len(cycle)
	rotated := make([]string, n)
	for i := 0; i < n; i++ {
		rotated[i] = cycle[(idx+i)%n]
	}
	return rotated
}

// DetectCycles performs cycle detection with no specific start node (default).
func (g *Graph) DetectCycles() [][]string {
	return g.DetectCyclesWithStart("")
}

// DetectCycles parses Python module files from input paths, builds the dependency graph by reading import statements,
// and detects circular dependencies, returning the cycles found.
// 'desiredStart' specifies the node to use as cycle start when outputting cycles.
// 'prefix' filters modules; only imports starting with 'prefix' are considered.
func (service *CircularDependencyService) DetectCycles(paths []string, desiredStart string, prefix string) ([][]string, error) {
	service.importPositions = make(map[string]map[string]Position)
	graph := NewGraph()

	// Simple regular expression matching import and from-import statements
	importRegex := regexp.MustCompile(`^(?:from|import)\s+([a-zA-Z0-9_.]+)`)

	for _, root := range paths {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("Error accessing path %q: %v\n", path, err)
				return nil
			}
			if info.IsDir() || !strings.HasSuffix(info.Name(), ".py") {
				return nil
			}

			relPath, err := filepath.Rel(root, path)
			if err != nil {
				relPath = path
			}
			relPath = filepath.ToSlash(relPath)

			file, err := os.Open(path)
			if err != nil {
				return nil
			}
			defer file.Close()

			lineNum := 0
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				lineNum++
				line := scanner.Text()
				if matches := importRegex.FindStringSubmatch(line); len(matches) == 2 {
					imported := matches[1]
					// Convert dotted module path to relative file path, e.g., 'player.module' => 'player/module.py'
					importedRelPath := strings.ReplaceAll(imported, ".", "/") + ".py"

					// Only consider dependencies starting with the prefix string
					if strings.HasPrefix(importedRelPath, prefix) {
						// Remove prefix from path to normalize and add edge from current file to imported file
						importedRelPath = importedRelPath[len(prefix):]
						if _, ok := service.importPositions[relPath]; !ok {
							service.importPositions[relPath] = map[string]Position{}
						}
						service.importPositions[relPath][importedRelPath] = Position{Line: lineNum, Column: 1}
						graph.AddEdge(relPath, importedRelPath)
					}
				}
			}
			if err := scanner.Err(); err != nil {
				fmt.Printf("Error reading file %q: %v\n", path, err)
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Error walking root path %q: %v\n", root, err)
		}
	}
	return graph.DetectCyclesWithStart(desiredStart), nil
}
