package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/analyzer"
	"github.com/ludo-technologies/pyscn/internal/parser"
	"github.com/ludo-technologies/pyscn/internal/version"
)

// DependencyService builds module dependency graphs across Python files
type DependencyService struct {
	parser *parser.Parser
}

// NewDependencyService creates a new dependency analysis service
func NewDependencyService() *DependencyService {
	return &DependencyService{parser: parser.New()}
}

// Analyze computes the dependency graph and cycles for the given request
func (s *DependencyService) Analyze(ctx context.Context, req domain.DependencyRequest) (*domain.DependencyResponse, error) {
	// Collect files
	fr := NewFileReader()
	files, err := fr.CollectPythonFiles(req.Paths, req.Recursive, req.IncludePatterns, req.ExcludePatterns)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, domain.NewAnalysisError("no Python files found", nil)
	}

	// Build local module map from file path -> module name
	roots := determineRoots(req.Paths)
	fileToModule := make(map[string]string)
	moduleToFiles := make(map[string][]string)
	for _, f := range files {
		mod := moduleNameFromPath(f, roots)
		if mod == "" {
			continue
		}
		fileToModule[f] = mod
		moduleToFiles[mod] = append(moduleToFiles[mod], f)
	}

	// Build graph
	g := analyzer.NewDepGraph()
	for mod := range moduleToFiles {
		g.AddNode(mod)
	}

	var warnings []string
	var errors []string

	// Infer a module prefix from roots (e.g., "app") and apply to local modules to better match absolute imports
	modulePrefix := ""
	if p, ok := shouldApplyPrefix(moduleToFiles, roots); ok {
		modulePrefix = p
		moduleToFiles = addPrefixToModules(moduleToFiles, modulePrefix)
		// update fileToModule mapping accordingly
		for file, mod := range fileToModule {
			fileToModule[file] = modulePrefix + "." + mod
		}
		// add nodes again for prefixed names
		g = analyzer.NewDepGraph()
		for mod := range moduleToFiles {
			g.AddNode(mod)
		}
	}

	// Parse each file and extract import edges
	for _, path := range files {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("dependency analysis cancelled: %w", ctx.Err())
		default:
		}

		content, rerr := fr.ReadFile(path)
		if rerr != nil {
			errors = append(errors, fmt.Sprintf("[%s] read failed: %v", path, rerr))
			continue
		}
		pres, perr := s.parser.Parse(ctx, content)
		if perr != nil {
			// non-fatal: continue with others
			warnings = append(warnings, fmt.Sprintf("[%s] parse error: %v", path, perr))
			continue
		}

		fromModule := fileToModule[path]
		if fromModule == "" {
			continue
		}
		// Extract imports recursively
		imports := collectImports(pres.AST)
		// Normalize imports into full module names
		targets := resolveImports(fromModule, imports)

		// Add edges only to known local modules, using longest known prefix fallback
		for _, t := range targets {
			if t == "" || t == fromModule {
				continue
			}
			// Prefer stdlib for plain single-segment imports (e.g., "logging") unless explicitly package-qualified
			if !strings.Contains(t, ".") {
				// try prefixed resolution only if we have a module prefix
				if modulePrefix != "" {
					if local := longestLocalOrPrefixedPrefix(t, moduleToFiles, modulePrefix); local != "" {
						g.AddEdge(fromModule, local)
					}
				}
				continue
			}
			if local := longestLocalOrPrefixedPrefix(t, moduleToFiles, modulePrefix); local != "" {
				g.AddEdge(fromModule, local)
			}
		}
	}

	// Build response
	edges := g.Edges()
	respEdges := make([]domain.DependencyEdge, 0, len(edges))
	for _, e := range edges {
		if e[0] == e[1] { // exclude self-referential edges from response
			continue
		}
		respEdges = append(respEdges, domain.DependencyEdge{From: e[0], To: e[1]})
	}

	// Cycles
	var respCycles []domain.DependencyCycle
	for _, c := range g.Cycles() {
		respCycles = append(respCycles, domain.DependencyCycle{Modules: c})
	}

	// Sort module files for stable output
	for k := range moduleToFiles {
		sort.Strings(moduleToFiles[k])
	}

	resp := &domain.DependencyResponse{
		Modules:     moduleToFiles,
		Edges:       respEdges,
		Cycles:      respCycles,
		Summary:     domain.DependencySummary{Modules: len(moduleToFiles), Edges: len(respEdges), Cycles: len(respCycles), FilesAnalyzed: len(files)},
		Warnings:    warnings,
		Errors:      errors,
		GeneratedAt: time.Now().Format(time.RFC3339),
		Version:     version.Version,
		DOT:         g.ToDOT(),
	}

	// Optional: layer validation if provided
	if req.Architecture != nil && len(req.Architecture.Layers) > 0 && len(req.Architecture.Rules) > 0 {
		assignments := assignLayers(moduleToFiles, req.Architecture.Layers)
		violations := validateLayerRules(assignments, g, req.Architecture.Rules)
		if len(violations) > 0 {
			resp.LayerViolations = violations
			resp.Summary.LayerViolations = len(violations)
		}
	}
	return resp, nil
}

// determineRoots returns directory roots among provided paths
func determineRoots(paths []string) []string {
	var roots []string
	for _, p := range paths {
		// Normalize to absolute & clean; ignore files, keep directories
		ap := p
		if !filepath.IsAbs(ap) {
			if abs, err := filepath.Abs(ap); err == nil {
				ap = abs
			}
		}
		ap = filepath.Clean(ap)
		if st, err := os.Stat(ap); err == nil && st.IsDir() {
			roots = append(roots, ap)
		} else if err == nil {
			// include parent directory of files
			roots = append(roots, filepath.Dir(ap))
		}
	}
	// prefer longer roots first for longest-prefix match
	sort.Slice(roots, func(i, j int) bool { return len(roots[i]) > len(roots[j]) })
	return roots
}

// moduleNameFromPath converts a Python file path to a dotted module name using the first matching root
func moduleNameFromPath(path string, roots []string) string {
	p := filepath.Clean(path)
	// Find the longest root that is a prefix
	var root string
	for _, r := range roots {
		r = filepath.Clean(r)
		if strings.HasPrefix(p, r) {
			root = r
			break
		}
	}
	rel := p
	if root != "" {
		if r, err := filepath.Rel(root, p); err == nil {
			rel = r
		}
	}
	// Drop extension
	base := strings.TrimSuffix(rel, filepath.Ext(rel))
	parts := splitPath(base)
	if len(parts) == 0 {
		return ""
	}
	// Handle __init__ as package module
	if parts[len(parts)-1] == "__init__" {
		parts = parts[:len(parts)-1]
	}
	if len(parts) == 0 {
		// root package __init__.py becomes empty -> use directory name
		return ""
	}
	return strings.Join(parts, ".")
}

func splitPath(p string) []string {
	// normalize separators and split
	p = filepath.ToSlash(p)
	raw := strings.Split(p, "/")
	parts := make([]string, 0, len(raw))
	for _, x := range raw {
		if x == "" || x == "." {
			continue
		}
		parts = append(parts, x)
	}
	return parts
}

// importSpec captures an import statement statically extracted from AST
type importSpec struct {
	kind   string // "import" or "from"
	module string // base module for from-import (may be empty)
	names  []string
	level  int // for relative imports (from ... import ...)
}

// collectImports walks AST collecting import specs
func collectImports(ast *parser.Node) []importSpec {
	if ast == nil {
		return nil
	}
	var specs []importSpec
	var walk func(n *parser.Node)
	walk = func(n *parser.Node) {
		if n == nil {
			return
		}
		switch n.Type {
		case parser.NodeImport:
			// Children may contain alias nodes with dotted names
			var names []string
			for _, ch := range n.GetChildren() {
				if ch.Type == parser.NodeAlias && ch.Name != "" {
					names = append(names, ch.Name)
				}
			}
			if len(names) > 0 {
				specs = append(specs, importSpec{kind: "import", names: names})
			}
		case parser.NodeImportFrom:
			// module in n.Module, names in n.Names; level in n.Level
			names := make([]string, 0, len(n.Names))
			names = append(names, n.Names...)
			// also consider alias children
			for _, ch := range n.GetChildren() {
				if ch.Type == parser.NodeAlias && ch.Name != "" {
					names = append(names, ch.Name)
				}
			}
			specs = append(specs, importSpec{kind: "from", module: n.Module, names: names, level: n.Level})
		}
		for _, ch := range n.GetChildren() {
			walk(ch)
		}
	}
	walk(ast)
	return specs
}

// resolveImports converts import specs to absolute dotted module targets from the perspective of fromModule
func resolveImports(fromModule string, specs []importSpec) []string {
	var targets []string
	basePkg := packageOf(fromModule)
	for _, s := range specs {
		switch s.kind {
		case "import":
			for _, name := range s.names {
				if name = strings.TrimSpace(name); name != "" {
					targets = append(targets, name)
				}
			}
		case "from":
			// compute base considering relative level
			base := basePkg
			if s.level > 0 {
				base = ascend(base, s.level)
			}
			if s.module != "" {
				if base != "" {
					base = base + "." + s.module
				} else {
					base = s.module
				}
			}
			if len(s.names) == 0 {
				if base != "" {
					targets = append(targets, base)
				}
				continue
			}
			for _, nm := range s.names {
				nm = strings.TrimSpace(nm)
				if nm == "*" {
					if base != "" {
						targets = append(targets, base)
					}
					continue
				}
				// Conservative: depend on the base package; if submodule likely, include base.sub
				if base != "" {
					// Prefer base.submodule but also include base to avoid missing package-level deps
					targets = append(targets, base)
					targets = append(targets, base+"."+nm)
				} else {
					targets = append(targets, nm)
				}
			}
		}
	}
	// de-duplicate
	uniq := make(map[string]struct{}, len(targets))
	var out []string
	for _, t := range targets {
		if t == "" {
			continue
		}
		if _, ok := uniq[t]; !ok {
			uniq[t] = struct{}{}
			out = append(out, t)
		}
	}
	return out
}

func packageOf(module string) string {
	// package is module without the last segment
	if idx := strings.LastIndex(module, "."); idx >= 0 {
		return module[:idx]
	}
	return ""
}

func ascend(base string, levels int) string {
	if base == "" || levels <= 0 {
		return base
	}
	parts := strings.Split(base, ".")
	if levels >= len(parts) {
		return ""
	}
	return strings.Join(parts[:len(parts)-levels], ".")
}

// longestLocalPrefix finds the longest dotted prefix of name that exists in localModules
func longestLocalPrefix(name string, local map[string][]string) string {
	n := name
	for {
		if _, ok := local[n]; ok {
			return n
		}
		if i := strings.LastIndex(n, "."); i >= 0 {
			n = n[:i]
			continue
		}
		return ""
	}
}

// longestLocalOrPrefixedPrefix tries direct match, then tries with inferred prefix
func longestLocalOrPrefixedPrefix(name string, local map[string][]string, prefix string) string {
	if m := longestLocalPrefix(name, local); m != "" {
		return m
	}
	if prefix != "" && !strings.HasPrefix(name, prefix+".") {
		if m := longestLocalPrefix(prefix+"."+name, local); m != "" {
			return m
		}
	}
	return ""
}

// inferModulePrefix returns the base directory name of the primary root as a heuristic
func inferModulePrefix(roots []string) string {
	if len(roots) == 0 {
		return ""
	}
	base := filepath.Base(roots[0])
	if base == "." || base == ".." || base == string(filepath.Separator) {
		return ""
	}
	return base
}

// addPrefixToModules rebuilds a module map by prefixing each module name with prefix.
func addPrefixToModules(mods map[string][]string, prefix string) map[string][]string {
	out := make(map[string][]string, len(mods))
	for k, v := range mods {
		if strings.HasPrefix(k, prefix+".") {
			out[k] = v
		} else {
			out[prefix+"."+k] = v
		}
	}
	return out
}

// shouldApplyPrefix decides whether to add a module prefix based on module name shapes.
// If most modules are single-segment (no dots), prefix with the root basename; otherwise, do not prefix.
func shouldApplyPrefix(mods map[string][]string, roots []string) (string, bool) {
	if len(roots) == 0 || len(mods) == 0 {
		return "", false
	}
	base := inferModulePrefix(roots)
	if base == "" {
		return "", false
	}
	total := 0
	dotless := 0
	for k := range mods {
		total++
		if !strings.Contains(k, ".") {
			dotless++
		}
	}
	if total == 0 {
		return "", false
	}
	// apply prefix when majority are dotless (heuristic)
	if float64(dotless)/float64(total) >= 0.6 {
		return base, true
	}
	return "", false
}

// assignLayers maps modules to layer names based on package patterns
func assignLayers(moduleToFiles map[string][]string, layers []domain.ArchitectureLayer) map[string]string {
	// Precompile patterns
	compiled := make([][]*regexp.Regexp, len(layers))
	for i, layer := range layers {
		for _, pat := range layer.Packages {
			if rx := globToRegexp(pat); rx != nil {
				compiled[i] = append(compiled[i], rx)
			}
		}
	}
	assign := make(map[string]string, len(moduleToFiles))
	for mod := range moduleToFiles {
		for i, layer := range layers {
			for _, rx := range compiled[i] {
				if rx.MatchString(mod) {
					assign[mod] = layer.Name
					goto nextMod
				}
			}
		}
	nextMod:
	}
	return assign
}

// validateLayerRules checks edges against allowed layer transitions
func validateLayerRules(assign map[string]string, g *analyzer.DepGraph, rules []domain.ArchitectureRule) []domain.LayerViolation {
	// Build allow map
	allow := make(map[string]map[string]struct{})
	for _, r := range rules {
		set := allow[r.From]
		if set == nil {
			set = make(map[string]struct{})
			allow[r.From] = set
		}
		for _, a := range r.Allow {
			set[a] = struct{}{}
		}
	}
	var out []domain.LayerViolation
	for _, e := range g.Edges() {
		fm, tm := e[0], e[1]
		fl, okf := assign[fm]
		tl, okt := assign[tm]
		if !okf || !okt {
			continue // ignore unassigned modules
		}
		if fl == tl {
			continue // same layer allowed by default
		}
		if set, ok := allow[fl]; ok {
			if _, ok := set[tl]; ok {
				continue
			}
		}
		out = append(out, domain.LayerViolation{FromModule: fm, ToModule: tm, FromLayer: fl, ToLayer: tl})
	}
	return out
}

// globToRegexp converts a simple glob (supports * and ?) on dotted module names into a regexp
// Examples:
//
//	"src.presentation.*" -> ^src\.presentation\.[^.]+$
//	"src.**" -> ^src\..*$
func globToRegexp(glob string) *regexp.Regexp {
	if glob == "" {
		return nil
	}
	// normalize: treat "**" as ".*" across segments; "*" as any chars except dot
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(glob); i++ {
		c := glob[i]
		if c == '.' {
			b.WriteString(`\.`)
			continue
		}
		if c == '*' {
			// check double star
			if i+1 < len(glob) && glob[i+1] == '*' {
				b.WriteString(".*")
				i++
			} else {
				b.WriteString("[^.]+")
			}
			continue
		}
		if c == '?' {
			b.WriteString(".")
			continue
		}
		// escape regex meta
		if strings.ContainsRune(`+()[]{}^$|`, rune(c)) {
			b.WriteByte('\\')
		}
		b.WriteByte(c)
	}
	b.WriteString("$")
	rx, err := regexp.Compile(b.String())
	if err != nil {
		return nil
	}
	return rx
}
