package analyzer

import (
    "fmt"
    "sort"
    "strings"
)

// DepGraph represents a directed graph of module dependencies
type DepGraph struct {
    // adjacency: from -> set(to)
    adjacency map[string]map[string]struct{}
    // nodes tracks all known module names
    nodes map[string]struct{}
}

// NewDepGraph creates an empty dependency graph
func NewDepGraph() *DepGraph {
    return &DepGraph{
        adjacency: make(map[string]map[string]struct{}),
        nodes:     make(map[string]struct{}),
    }
}

// AddNode adds a module node
func (g *DepGraph) AddNode(name string) {
    if name == "" {
        return
    }
    if _, ok := g.nodes[name]; !ok {
        g.nodes[name] = struct{}{}
    }
}

// AddEdge adds a directed edge from -> to
func (g *DepGraph) AddEdge(from, to string) {
    if from == "" || to == "" {
        return
    }
    g.AddNode(from)
    g.AddNode(to)
    if _, ok := g.adjacency[from]; !ok {
        g.adjacency[from] = make(map[string]struct{})
    }
    g.adjacency[from][to] = struct{}{}
}

// Nodes returns all node names (sorted)
func (g *DepGraph) Nodes() []string {
    out := make([]string, 0, len(g.nodes))
    for n := range g.nodes {
        out = append(out, n)
    }
    sort.Strings(out)
    return out
}

// Edges returns all edges as pairs (sorted by from,to)
func (g *DepGraph) Edges() [][2]string {
    var edges [][2]string
    for from, tos := range g.adjacency {
        for to := range tos {
            edges = append(edges, [2]string{from, to})
        }
    }
    sort.Slice(edges, func(i, j int) bool {
        if edges[i][0] == edges[j][0] {
            return edges[i][1] < edges[j][1]
        }
        return edges[i][0] < edges[j][0]
    })
    return edges
}

// StronglyConnectedComponents finds SCCs using Tarjan's algorithm
func (g *DepGraph) StronglyConnectedComponents() [][]string {
    index := 0
    stack := []string{}
    onStack := make(map[string]bool)
    indices := make(map[string]int)
    lowlink := make(map[string]int)
    var sccs [][]string

    var strongconnect func(v string)
    strongconnect = func(v string) {
        indices[v] = index
        lowlink[v] = index
        index++
        stack = append(stack, v)
        onStack[v] = true

        for w := range g.adjacency[v] {
            if _, seen := indices[w]; !seen {
                strongconnect(w)
                if lowlink[w] < lowlink[v] {
                    lowlink[v] = lowlink[w]
                }
            } else if onStack[w] {
                if indices[w] < lowlink[v] {
                    lowlink[v] = indices[w]
                }
            }
        }

        if lowlink[v] == indices[v] {
            // start new SCC
            var comp []string
            for {
                n := len(stack) - 1
                w := stack[n]
                stack = stack[:n]
                onStack[w] = false
                comp = append(comp, w)
                if w == v {
                    break
                }
            }
            if len(comp) > 0 {
                sort.Strings(comp)
                sccs = append(sccs, comp)
            }
        }
    }

    // ensure adjacency entries for nodes with no outgoing edges
    for n := range g.nodes {
        if _, ok := g.adjacency[n]; !ok {
            g.adjacency[n] = make(map[string]struct{})
        }
    }
    // run Tarjan
    for n := range g.nodes {
        if _, seen := indices[n]; !seen {
            strongconnect(n)
        }
    }
    return sccs
}

// Cycles returns only SCCs that represent cycles (size > 1 or self-loop)
func (g *DepGraph) Cycles() [][]string {
    sccs := g.StronglyConnectedComponents()
    var cycles [][]string
    for _, comp := range sccs {
        if len(comp) > 1 {
            cycles = append(cycles, comp)
            continue
        }
        // single node: check self-loop
        v := comp[0]
        if _, ok := g.adjacency[v][v]; ok {
            cycles = append(cycles, []string{v})
        }
    }
    return cycles
}

// ToDOT returns a DOT representation of the graph, highlighting cycle edges in red
func (g *DepGraph) ToDOT() string {
    // Build a set for quick cycle membership
    cycleSet := make(map[string]struct{})
    for _, comp := range g.Cycles() {
        for _, n := range comp {
            cycleSet[n] = struct{}{}
        }
    }

    var b strings.Builder
    b.WriteString("digraph dependencies {\n")
    // nodes
    for _, n := range g.Nodes() {
        if _, inCycle := cycleSet[n]; inCycle {
            fmt.Fprintf(&b, "  \"%s\" [style=filled, fillcolor=\"#ffe6e6\"];\n", n)
        } else {
            fmt.Fprintf(&b, "  \"%s\";\n", n)
        }
    }
    // edges
    for _, e := range g.Edges() {
        color := ""
        if _, a := cycleSet[e[0]]; a {
            if _, bcy := cycleSet[e[1]]; bcy {
                color = " [color=red]"
            }
        }
        fmt.Fprintf(&b, "  \"%s\" -> \"%s\"%s;\n", e[0], e[1], color)
    }
    b.WriteString("}\n")
    return b.String()
}

