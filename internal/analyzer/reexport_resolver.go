package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/ludo-technologies/pyscn/internal/parser"
)

// ReExportEntry represents a single re-exported name from an __init__.py
type ReExportEntry struct {
	Name         string // The exported name (e.g., "SomeClass")
	SourceModule string // The actual source module (e.g., "pkg_a.module_x")
	SourceName   string // The name in the source module (may differ if aliased)
}

// ReExportMap holds all exports from a package's __init__.py
type ReExportMap struct {
	PackageName string                    // The package name (e.g., "pkg_a")
	Exports     map[string]*ReExportEntry // name -> source info
	AllDeclared []string                  // Names in __all__ if declared
	HasAllDecl  bool                      // True if __all__ is explicitly declared
}

// ReExportResolver resolves re-exports in __init__.py files
type ReExportResolver struct {
	projectRoot string
	cache       map[string]*ReExportMap // package name -> re-export map
}

// NewReExportResolver creates a new resolver
func NewReExportResolver(projectRoot string) *ReExportResolver {
	return &ReExportResolver{
		projectRoot: projectRoot,
		cache:       make(map[string]*ReExportMap),
	}
}

// GetReExportMap returns the re-export map for a package (cached)
func (r *ReExportResolver) GetReExportMap(packageName string) (*ReExportMap, error) {
	// Check cache first
	if cached, exists := r.cache[packageName]; exists {
		return cached, nil
	}

	// Find and parse the __init__.py file
	initPath := r.findInitFile(packageName)
	if initPath == "" {
		// No __init__.py found - cache empty result
		r.cache[packageName] = nil
		return nil, nil
	}

	// Parse the __init__.py file
	reExportMap, err := r.parseInitFile(initPath, packageName)
	if err != nil {
		// Cache nil on error to avoid repeated parsing failures
		r.cache[packageName] = nil
		return nil, err
	}

	r.cache[packageName] = reExportMap
	return reExportMap, nil
}

// ResolveReExport resolves an imported name to its actual source module
// Returns (sourceModule, found)
func (r *ReExportResolver) ResolveReExport(packageName, importedName string) (string, bool) {
	reExportMap, err := r.GetReExportMap(packageName)
	if err != nil || reExportMap == nil {
		return "", false
	}

	entry, exists := reExportMap.Exports[importedName]
	if !exists {
		return "", false
	}

	return entry.SourceModule, true
}

// findInitFile finds the __init__.py file for a package
func (r *ReExportResolver) findInitFile(packageName string) string {
	// Convert package name to path
	packagePath := filepath.Join(r.projectRoot, strings.ReplaceAll(packageName, ".", string(filepath.Separator)))

	initPath := filepath.Join(packagePath, "__init__.py")
	if _, err := os.Stat(initPath); err == nil {
		return initPath
	}

	return ""
}

// parseInitFile parses an __init__.py file and extracts re-export information
func (r *ReExportResolver) parseInitFile(filePath, packageName string) (*ReExportMap, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	p := parser.New()
	ctx := context.Background()
	result, err := p.Parse(ctx, content)
	if err != nil {
		return nil, err
	}

	exports := make(map[string]*ReExportEntry)

	// Extract re-exports from import statements
	r.extractReExports(result.AST, packageName, exports)

	// Extract __all__ declaration if present
	allNames, hasAll := r.extractAllDeclaration(result.AST)

	// If __all__ is declared, filter exports to only include names in __all__
	if hasAll && len(allNames) > 0 {
		filteredExports := make(map[string]*ReExportEntry)
		for _, name := range allNames {
			if entry, exists := exports[name]; exists {
				filteredExports[name] = entry
			}
		}
		exports = filteredExports
	}

	return &ReExportMap{
		PackageName: packageName,
		Exports:     exports,
		AllDeclared: allNames,
		HasAllDecl:  hasAll,
	}, nil
}

// extractReExports extracts re-export entries from import statements
func (r *ReExportResolver) extractReExports(ast *parser.Node, packageName string, exports map[string]*ReExportEntry) {
	if ast == nil {
		return
	}

	// Walk the AST looking for ImportFrom statements
	r.walkNode(ast, func(node *parser.Node) bool {
		if node.Type == parser.NodeImportFrom {
			r.processImportFrom(node, packageName, exports)
		}
		return true
	})
}

// processImportFrom processes a single "from ... import ..." statement
func (r *ReExportResolver) processImportFrom(node *parser.Node, packageName string, exports map[string]*ReExportEntry) {
	// Get the module being imported from
	module := node.Module
	if module == "" {
		return
	}

	// Determine the source module
	var sourceModule string
	if node.Level > 0 {
		// Relative import: from .module import name
		// Level 1 means from current package, level 2 means from parent, etc.
		if node.Level == 1 && module != "" {
			sourceModule = packageName + "." + module
		} else if node.Level == 1 && module == "" {
			// from . import something - importing from same package
			sourceModule = packageName
		} else {
			// Handle deeper relative imports
			parts := strings.Split(packageName, ".")
			if node.Level <= len(parts) {
				parentPkg := strings.Join(parts[:len(parts)-node.Level+1], ".")
				if module != "" {
					sourceModule = parentPkg + "." + module
				} else {
					sourceModule = parentPkg
				}
			}
		}
	} else {
		// Absolute import - check if it's re-exporting from within the same package
		if strings.HasPrefix(module, packageName+".") {
			sourceModule = module
		} else {
			// External import - not a re-export within the package
			return
		}
	}

	if sourceModule == "" || sourceModule == packageName {
		return
	}

	// Process each imported name
	// For aliased imports like "from .module import X as Y":
	//   - node.Names contains [X] (original name)
	//   - node.Children contains Alias node with Name="X", Value="Y" (alias)
	for _, name := range node.Names {
		if name == "*" {
			// Wildcard import - we can't track individual names without parsing the source module
			// NOTE: Wildcard re-exports are not supported
			continue
		}

		// Check for aliases in the children (Alias nodes)
		exportedName := name
		sourceName := name
		for _, child := range node.Children {
			if child.Type == parser.NodeAlias && child.Name == name {
				// Found the Alias node for this import
				// child.Name is the original name, child.Value is the alias (if any)
				if aliasName, ok := child.Value.(string); ok && aliasName != "" {
					exportedName = aliasName
				}
				break
			}
		}

		exports[exportedName] = &ReExportEntry{
			Name:         exportedName,
			SourceModule: sourceModule,
			SourceName:   sourceName,
		}
	}
}

// extractAllDeclaration extracts names from __all__ = [...] assignment
func (r *ReExportResolver) extractAllDeclaration(ast *parser.Node) ([]string, bool) {
	var allNames []string
	hasAll := false

	if ast == nil {
		return allNames, hasAll
	}

	r.walkNode(ast, func(node *parser.Node) bool {
		if node.Type == parser.NodeAssign {
			// Check if this is __all__ = [...]
			for _, target := range node.Targets {
				if target.Type == parser.NodeName && target.Name == "__all__" {
					hasAll = true
					// Extract list elements from the children/value
					allNames = r.extractListElements(node)
					return false // Stop walking
				}
			}
		}
		return true
	})

	return allNames, hasAll
}

// extractListElements extracts string elements from a list assignment
func (r *ReExportResolver) extractListElements(assignNode *parser.Node) []string {
	var elements []string

	// The Value field contains the list/tuple node
	valueNode, ok := assignNode.Value.(*parser.Node)
	if !ok || valueNode == nil {
		return elements
	}

	if valueNode.Type == parser.NodeList || valueNode.Type == parser.NodeTuple {
		for _, child := range valueNode.Children {
			if child.Type == parser.NodeConstant {
				if str, ok := child.Value.(string); ok {
					elements = append(elements, str)
				}
			}
		}
	}

	return elements
}

// walkNode recursively walks the AST and calls the visitor function
func (r *ReExportResolver) walkNode(node *parser.Node, visitor func(*parser.Node) bool) {
	if node == nil {
		return
	}

	if !visitor(node) {
		return
	}

	// Walk all children
	for _, child := range node.GetChildren() {
		r.walkNode(child, visitor)
	}
}
