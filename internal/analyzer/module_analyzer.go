```go
package analyzer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// ModuleAnalyzer analyzes module-level dependencies and builds dependency graphs
type ModuleAnalyzer struct {
	projectRoot     string
	pythonPath      []string
	excludePatterns []string
	includePatterns []string

	// Module resolution cache
	resolvedModules map[string]string // module name -> file path
	packageCache    map[string]bool   // package name -> is valid package

	// Re-export resolution
	reExportResolver *ReExportResolver // Resolves re-exports in __init__.py files

	// Analysis options
	includeStdLib     bool
	includeThirdParty bool
	followRelative    bool
}

// ModuleAnalysisOptions configures module analysis behavior
type ModuleAnalysisOptions struct {
	ProjectRoot       string   // Project root directory
	PythonPath        []string // Additional Python path entries
	ExcludePatterns   []string // Module patterns to exclude
	IncludePatterns   []string // Module patterns to include (default: ["**/*.py"])
	IncludeStdLib     bool     // Include standard library dependencies
	IncludeThirdParty bool     // Include third-party dependencies
	FollowRelative    bool     // Follow relative imports
}

// DefaultModuleAnalysisOptions returns default analysis options
func DefaultModuleAnalysisOptions() *ModuleAnalysisOptions {
	return &ModuleAnalysisOptions{
		IncludePatterns:   []string{"**/*.py"},
		ExcludePatterns:   []string{"test_*.py", "*_test.py", "__pycache__", "*.pyc"},
		IncludeStdLib:     false,
		IncludeThirdParty: true,
		FollowRelative:    true,
	}
}

// NewModuleAnalyzer creates a new module analyzer
func NewModuleAnalyzer(options *ModuleAnalysisOptions) (*ModuleAnalyzer, error) {
	if options == nil {
		options = DefaultModuleAnalysisOptions()
	}

	// Determine project root
	projectRoot := options.ProjectRoot
	if projectRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		projectRoot = cwd
	}

	// Make project root absolute
	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project root: %w", err)
	}

	analyzer := &ModuleAnalyzer{
		projectRoot:       absRoot,
		pythonPath:        append([]string{absRoot}, options.PythonPath...),
		resolvedModules:   make(map[string]string),
		packageCache:      make(map[string]bool),
		reExportResolver:  NewReExportResolver(absRoot),
		includeStdLib:     options.IncludeStdLib,
		includeThirdParty: options.IncludeThirdParty,
		followRelative:    options.FollowRelative,
	}

	analyzer.excludePatterns = append(analyzer.excludePatterns, options.ExcludePatterns...)

	analyzer.includePatterns = append(analyzer.includePatterns, options.IncludePatterns...)

	return analyzer, nil
}

// AnalyzeProject analyzes all Python modules in the project and builds a dependency graph
func (ma *ModuleAnalyzer) AnalyzeProject() (*DependencyGraph, error) {
	// Collect all Python files in the project
	files, err := ma.collectPythonFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to collect Python files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no Python files found in project")
	}

	// Create dependency graph
	graph := NewDependencyGraph(ma.projectRoot)

	// First pass: Add all modules to graph
	for _, filePath := range files {
		moduleName := ma.filePathToModuleName(filePath)
		if moduleName != "" {
			graph.AddModule(moduleName, filePath)
		}
	}

	// Second pass: Analyze dependencies for each module
	for _, filePath := range files {
		if err := ma.analyzeModuleDependencies(graph, filePath); err != nil {
			// Log warning but continue with other files
			continue
		}
	}

	return graph, nil
}

// AnalyzeFiles analyzes specific Python files and builds a dependency graph
func (ma *ModuleAnalyzer) AnalyzeFiles(filePaths []string) (*DependencyGraph, error) {
	graph := NewDependencyGraph(ma.projectRoot)

	// Filter and validate files
	var validFiles []string
	for _, filePath := range filePaths {
		if ma.isValidPythonFile(filePath) {
			absPath, err := filepath.Abs(filePath)
			if err == nil {
				validFiles = append(validFiles, absPath)
			}
		}
	}

	if len(validFiles) == 0 {
		return nil, fmt.Errorf("no valid Python files provided")
	}

	// Add modules to graph
	for _, filePath := range validFiles {
		moduleName := ma.filePathToModuleName(filePath)
		if moduleName != "" {
			graph.AddModule(moduleName, filePath)
		}
	}

	// Analyze dependencies
	for _, filePath := range validFiles {
		if err := ma.analyzeModuleDependencies(graph, filePath); err != nil {
			continue
		}
	}

	return graph, nil
}

// analyzeModuleDependencies analyzes imports in a single module and adds dependencies to graph
func (ma *ModuleAnalyzer) analyzeModuleDependencies(graph *DependencyGraph, filePath string) error {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Parse the Python file
	p := parser.New()
	ctx := context.Background()
	result, err := p.Parse(ctx, content)
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	moduleName := ma.filePathToModuleName(filePath)
	if moduleName == "" {
		return fmt.Errorf("could not determine module name for %s", filePath)
	}

	// Get module node
	module := graph.GetModule(moduleName)
	if module == nil {
		return fmt.Errorf("module not found in graph: %s", moduleName)
	}

	// Extract module information
	ma.extractModuleInfo(result.AST, module)

	// Collect import statements
	imports := ma.collectModuleImports(result.AST, filePath)

	// Process each import
	for _, imp := range imports {
		// Skip TYPE_CHECKING imports for circular dependency detection
		if imp.IsTypeChecking {
			continue
		}

		targetModule := ma.resolveImport(imp, filePath)
		if targetModule != "" && ma.shouldIncludeDependency(targetModule) {
			// Skip dependencies from __init__.py to its own submodules
			// This is a common Python pattern for re-exporting (internal structure)
			if strings.HasSuffix(filePath, "__init__.py") {
				// Check if target is a submodule of the current package
				if strings.HasPrefix(targetModule, moduleName+".") {
					continue // Skip this dependency
				}
			}

			// Determine edge type
			edgeType := DependencyEdgeImport
			if imp.IsRelative {
				edgeType = DependencyEdgeRelative
			} else if len(imp.ImportedNames) > 0 {
				edgeType = DependencyEdgeFromImport
			}

			// For "from package import name" style imports, resolve through re-exports
			// to find the actual source module. Each imported name may come from a
			// different source module, so we need to add edges for each.
			if len(imp.ImportedNames) > 0 && !imp.IsRelative {
				resolvedModules := make(map[string]bool)
				for _, importedName := range imp.ImportedNames {
					if resolvedModule, found := ma.reExportResolver.ResolveReExport(targetModule, importedName); found {
						resolvedModules[resolvedModule] = true
					} else {
						// Not a re-export, use the original target
						resolvedModules[targetModule] = true
					}
				}
				// Add dependency for each unique resolved module
				for resolvedModule := range resolvedModules {
					graph.AddDependency(moduleName, resolvedModule, edgeType, imp)
				}
			} else {
				// Add dependency to graph
				graph.AddDependency(moduleName, targetModule, edgeType, imp)
			}
		}
	}

	return nil
}

// collectModuleImports collects all import statements from AST
func (ma *ModuleAnalyzer) collectModuleImports(ast *parser.Node, filePath string) []*ImportInfo {
	var imports []*ImportInfo

	ma.walkNode(ast, func(node *parser.Node) bool {
		switch node.Type {
		case parser.NodeImport:
			// Handle "import module" statements
			isTypeChecking := ma.isInTypeCheckingBlock(node)

			if len(node.Children) > 0 {
				for _, child := range node.Children {
					if child.Type == parser.NodeAlias {
						imp := &ImportInfo{
							Statement:      fmt.Sprintf("import %s", child.Name),
							ImportedNames:  []string{child.Name},
							IsRelative:     false,
							Line:           node.Location.StartLine,
							IsTypeChecking: isTypeChecking,
						}
						if child.Value != nil {
							if alias, ok := child.Value.(string); ok {
								imp.Alias = alias
							}
						}
						imports = append(imports, imp)
					}
				}
			} else if len(node.Names) > 0 {
				for _, name := range node.Names {
					imp := &ImportInfo{
						Statement:      fmt.Sprintf("import %s", name),
						ImportedNames:  []string{name},
						IsRelative:     false,
						Line:           node.Location.StartLine,
						IsTypeChecking: isTypeChecking,
					}
					imports = append(imports, imp)
				}
			}

		case parser.NodeImportFrom:
			// Handle "from module import name" statements
			isTypeChecking := ma.isInTypeCheckingBlock(node)
			module := node.Module
			level := ma.calculateRelativeLevel(node.Module)

			// Get imported names - use map to deduplicate since names may appear
			// in both node.Names and child Alias nodes depending on parser version
			nameSet := make(map[string]bool)
			for _, name := range node.Names {
				nameSet[name] = true
			}
			for _, child := range node.Children {
				if child.Type == parser.NodeAlias {
					nameSet[child.Name] = true
				}
			}
			importedNames := make([]string, 0, len(nameSet))
			for name := range nameSet {
				importedNames = append(importedNames, name)
			}

			imp := &ImportInfo{
				Statement:      ma.buildImportStatement(node),
				ImportedNames:  importedNames,
				IsRelative:     level > 0,
				Level:          level,
				Line:           node.Location.StartLine,
				IsTypeChecking: isTypeChecking,
			}

			// Clean module name for relative imports
			if imp.IsRelative {
				imp.Statement = strings.TrimLeft(module, ".")
			} else {
				imp.Statement = module
			}

			imports = append(imports, imp)
		}
		return true
	})

	return imports
}