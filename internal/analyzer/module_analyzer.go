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

// resolveImport resolves an import to a module name
func (ma *ModuleAnalyzer) resolveImport(imp *ImportInfo, fromFile string) string {
	if imp.IsRelative {
		return ma.resolveRelativeImport(imp, fromFile)
	}
	// For absolute imports, try to resolve within the project first
	return ma.resolveAbsoluteImportWithProject(imp, fromFile)
}

// resolveRelativeImport resolves relative imports like "from .module import name"
func (ma *ModuleAnalyzer) resolveRelativeImport(imp *ImportInfo, fromFile string) string {
	if !ma.followRelative {
		return ""
	}

	// Get the directory of the current file
	currentDir := filepath.Dir(fromFile)

	// Navigate up the directory tree based on the level
	targetDir := currentDir
	for i := 0; i < imp.Level; i++ {
		targetDir = filepath.Dir(targetDir)
	}

	// Build the target module path
	if imp.Statement != "" {
		targetPath := filepath.Join(targetDir, strings.ReplaceAll(imp.Statement, ".", string(filepath.Separator)))
		return ma.pathToModuleName(targetPath)
	}

	return ma.pathToModuleName(targetDir)
}

// resolveAbsoluteImport resolves absolute imports
func (ma *ModuleAnalyzer) resolveAbsoluteImport(imp *ImportInfo) string {
	moduleName := ma.moduleNameFromImport(imp)
	if moduleName == "" {
		return ""
	}

	// Check cache first
	if resolved, exists := ma.resolvedModules[moduleName]; exists {
		return resolved
	}

	// Try to find the module in the Python path
	for _, pathEntry := range ma.pythonPath {
		modulePath := filepath.Join(pathEntry, strings.ReplaceAll(moduleName, ".", string(filepath.Separator)))

		// Check for package (__init__.py)
		if initFile := filepath.Join(modulePath, "__init__.py"); ma.fileExists(initFile) {
			ma.resolvedModules[moduleName] = moduleName
			return moduleName
		}

		// Check for module file
		if moduleFile := modulePath + ".py"; ma.fileExists(moduleFile) {
			ma.resolvedModules[moduleName] = moduleName
			return moduleName
		}
	}

	// Check if it's a standard library or third-party module
	if ma.isStandardLibrary(moduleName) {
		if ma.includeStdLib {
			ma.resolvedModules[moduleName] = moduleName
			return moduleName
		}
		return ""
	}

	if ma.includeThirdParty {
		ma.resolvedModules[moduleName] = moduleName
		return moduleName
	}

	return ""
}

// resolveAbsoluteImportWithProject resolves absolute imports, checking project modules first
func (ma *ModuleAnalyzer) resolveAbsoluteImportWithProject(imp *ImportInfo, fromFile string) string {
	moduleName := ma.moduleNameFromImport(imp)
	if moduleName == "" {
		return ""
	}

	// Check cache first
	if resolved, exists := ma.resolvedModules[moduleName]; exists {
		return resolved
	}

	// First, try to resolve within the current project directory
	// Build possible module path relative to the file's directory
	currentDir := filepath.Dir(fromFile)

	// Try to find the module in the same directory or project root
	searchPaths := []string{
		currentDir,               // Current directory
		ma.projectRoot,           // Project root
		filepath.Dir(currentDir), // Parent directory
	}

	for _, searchPath := range searchPaths {
		// Try to build module path from the import name
		modulePath := filepath.Join(searchPath, strings.ReplaceAll(moduleName, ".", string(filepath.Separator)))

		// Check if it's a Python file
		if moduleFile := modulePath + ".py"; ma.fileExists(moduleFile) {
			// Calculate the module name based on project structure
			resolvedName := ma.filePathToModuleName(moduleFile)
			if resolvedName != "" {
				ma.resolvedModules[moduleName] = resolvedName
				return resolvedName
			}
		}

		// Check if it's a package (directory with __init__.py)
		if initFile := filepath.Join(modulePath, "__init__.py"); ma.fileExists(initFile) {
			resolvedName := ma.filePathToModuleName(initFile)
			if resolvedName != "" {
				// For __init__.py files, use the package name (without __init__)
				resolvedName = strings.TrimSuffix(resolvedName, ".__init__")
				ma.resolvedModules[moduleName] = resolvedName
				return resolvedName
			}
		}
	}

	// Fall back to the original resolveAbsoluteImport logic
	return ma.resolveAbsoluteImport(imp)
}

// moduleNameFromImport normalizes the module name from an import statement
func (ma *ModuleAnalyzer) moduleNameFromImport(imp *ImportInfo) string {
	if imp == nil {
		return ""
	}

	moduleName := strings.TrimSpace(imp.Statement)

	// Handle plain "import foo as bar" statements by stripping the prefix and alias
	if strings.HasPrefix(moduleName, "import ") {
		moduleName = strings.TrimSpace(strings.TrimPrefix(moduleName, "import "))
		if idx := strings.Index(moduleName, " as "); idx != -1 {
			moduleName = moduleName[:idx]
		}
	}

	if moduleName == "" && len(imp.ImportedNames) > 0 {
		moduleName = imp.ImportedNames[0]
	}

	return strings.TrimSpace(moduleName)
}

// Helper methods

// collectPythonFiles collects all Python files in the project
func (ma *ModuleAnalyzer) collectPythonFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(ma.projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip problematic files
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file matches include patterns
		if ma.matchesIncludePatterns(path) && !ma.matchesExcludePatterns(path) {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// filePathToModuleName converts a file path to a Python module name
func (ma *ModuleAnalyzer) filePathToModuleName(filePath string) string {
	// Make path relative to project root
	relPath, err := filepath.Rel(ma.projectRoot, filePath)
	if err != nil {
		return ""
	}

	// Remove .py extension
	relPath = strings.TrimSuffix(relPath, ".py")

	// Handle __init__.py files
	if strings.HasSuffix(relPath, "__init__") {
		relPath = relPath[:len(relPath)-len("__init__")]
		if strings.HasSuffix(relPath, string(filepath.Separator)) {
			relPath = relPath[:len(relPath)-1]
		}
	}

	// Convert path separators to dots
	moduleName := strings.ReplaceAll(relPath, string(filepath.Separator), ".")

	// Clean up leading/trailing dots
	moduleName = strings.Trim(moduleName, ".")

	return moduleName
}

// pathToModuleName converts a directory path to a module name
func (ma *ModuleAnalyzer) pathToModuleName(path string) string {
	relPath, err := filepath.Rel(ma.projectRoot, path)
	if err != nil {
		return ""
	}
	return strings.ReplaceAll(relPath, string(filepath.Separator), ".")
}

// extractModuleInfo extracts information about the module from its AST
func (ma *ModuleAnalyzer) extractModuleInfo(ast *parser.Node, module *ModuleNode) {
	var functionCount, classCount int
	var publicNames []string

	ma.walkNode(ast, func(node *parser.Node) bool {
		switch node.Type {
		case parser.NodeFunctionDef, parser.NodeAsyncFunctionDef:
			functionCount++
			if node.Name != "" && !strings.HasPrefix(node.Name, "_") {
				publicNames = append(publicNames, node.Name)
			}
		case parser.NodeClassDef:
			classCount++
			if node.Name != "" && !strings.HasPrefix(node.Name, "_") {
				publicNames = append(publicNames, node.Name)
			}
		}
		return true
	})

	// Update module information
	module.FunctionCount = functionCount
	module.ClassCount = classCount
	module.PublicNames = publicNames

	// Count lines (approximation)
	if _, err := os.Stat(module.FilePath); err == nil {
		module.LineCount = ma.estimateLineCount(module.FilePath)
	}
}

// shouldIncludeDependency checks if a dependency should be included
func (ma *ModuleAnalyzer) shouldIncludeDependency(moduleName string) bool {
	if moduleName == "" {
		return false
	}

	// Check exclude patterns
	for _, pattern := range ma.excludePatterns {
		if matched, _ := doublestar.Match(pattern, moduleName); matched {
			return false
		}
	}

	return true
}

// Utility methods

// walkNode recursively walks AST nodes
func (ma *ModuleAnalyzer) walkNode(node *parser.Node, visitor func(*parser.Node) bool) {
	if node == nil || !visitor(node) {
		return
	}

	for _, child := range node.Children {
		ma.walkNode(child, visitor)
	}

	for _, child := range node.Body {
		ma.walkNode(child, visitor)
	}
}

// calculateRelativeLevel calculates the level of relative import (number of dots)
func (ma *ModuleAnalyzer) calculateRelativeLevel(module string) int {
	level := 0
	for _, char := range module {
		if char == '.' {
			level++
		} else {
			break
		}
	}
	return level
}

// buildImportStatement builds the original import statement string
func (ma *ModuleAnalyzer) buildImportStatement(node *parser.Node) string {
	if node.Type == parser.NodeImportFrom {
		var names []string
		for _, child := range node.Children {
			if child.Type == parser.NodeAlias {
				names = append(names, child.Name)
			}
		}
		return fmt.Sprintf("from %s import %s", node.Module, strings.Join(names, ", "))
	}
	return ""
}

// isValidPythonFile checks if a file is a valid Python file
func (ma *ModuleAnalyzer) isValidPythonFile(filePath string) bool {
	return strings.HasSuffix(filePath, ".py") || strings.HasSuffix(filePath, ".pyi")
}

// fileExists checks if a file exists
func (ma *ModuleAnalyzer) fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// matchesIncludePatterns checks if path matches any include pattern
func (ma *ModuleAnalyzer) matchesIncludePatterns(path string) bool {
	if len(ma.includePatterns) == 0 {
		return true
	}
	for _, pattern := range ma.includePatterns {
		if matched, _ := doublestar.Match(pattern, filepath.Base(path)); matched {
			return true
		}
	}
	return false
}

// matchesExcludePatterns checks if path matches any exclude pattern
func (ma *ModuleAnalyzer) matchesExcludePatterns(path string) bool {
	for _, pattern := range ma.excludePatterns {
		if matched, _ := doublestar.Match(pattern, path); matched {
			return true
		}
		if matched, _ := doublestar.Match(pattern, filepath.Base(path)); matched {
			return true
		}
	}
	return false
}

// isStandardLibrary checks if a module is part of the Python standard library
func (ma *ModuleAnalyzer) isStandardLibrary(moduleName string) bool {
	// Common standard library modules
	stdLibModules := map[string]bool{
		"os": true, "sys": true, "re": true, "json": true, "datetime": true,
		"collections": true, "itertools": true, "functools": true, "operator": true,
		"math": true, "random": true, "string": true, "io": true, "pathlib": true,
		"unittest": true, "logging": true, "argparse": true, "configparser": true,
		"urllib": true, "http": true, "typing": true, "abc": true, "asyncio": true,
		"contextlib": true, "dataclasses": true, "enum": true, "pickle": true,
		"sqlite3": true, "csv": true, "xml": true, "html": true, "email": true,
	}

	// Check direct match
	if stdLibModules[moduleName] {
		return true
	}

	// Check root module for qualified names
	if strings.Contains(moduleName, ".") {
		rootModule := strings.Split(moduleName, ".")[0]
		return stdLibModules[rootModule]
	}

	return false
}

// estimateLineCount estimates line count for a file
func (ma *ModuleAnalyzer) estimateLineCount(filePath string) int {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0
	}
	return strings.Count(string(content), "\n") + 1
}

// isInTypeCheckingBlock checks if a node is inside a TYPE_CHECKING conditional block
func (ma *ModuleAnalyzer) isInTypeCheckingBlock(node *parser.Node) bool {
	// Walk up the parent chain to find if we're inside an if statement
	current := node.Parent
	for current != nil {
		if current.Type == parser.NodeIf {
			// Check if this is a TYPE_CHECKING condition
			if ma.isTypeCheckingCondition(current.Test) {
				return true
			}
		}
		current = current.Parent
	}
	return false
}

// isTypeCheckingCondition checks if an expression is a TYPE_CHECKING condition
func (ma *ModuleAnalyzer) isTypeCheckingCondition(expr *parser.Node) bool {
	if expr == nil {
		return false
	}

	// Handle simple case: just TYPE_CHECKING
	if expr.Type == parser.NodeName && expr.Name == "TYPE_CHECKING" {
		return true
	}

	// Handle attribute access: typing.TYPE_CHECKING
	if expr.Type == parser.NodeAttribute && expr.Name == "TYPE_CHECKING" {
		return true
	}

	// Handle binary operations that include TYPE_CHECKING
	// e.g., "TYPE_CHECKING and sys.version_info >= (3, 9)"
	if expr.Type == parser.NodeBoolOp {
		return ma.containsTypeChecking(expr)
	}

	// Handle comparisons and other complex expressions
	if expr.Type == parser.NodeCompare {
		return ma.containsTypeChecking(expr)
	}

	return false
}

// containsTypeChecking recursively checks if an expression contains TYPE_CHECKING
func (ma *ModuleAnalyzer) containsTypeChecking(node *parser.Node) bool {
	if node == nil {
		return false
	}

	// Check current node
	if (node.Type == parser.NodeName && node.Name == "TYPE_CHECKING") ||
		(node.Type == parser.NodeAttribute && node.Name == "TYPE_CHECKING") {
		return true
	}

	// Recursively check all children
	for _, child := range node.GetChildren() {
		if ma.containsTypeChecking(child) {
			return true
		}
	}

	return false
}
