package analyzer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/ludo-technologies/pyscn/domain"
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

	// Re-export resolution
	reExportResolver *ReExportResolver // Resolves re-exports in __init__.py files

	// Analysis options
	includeStdLib     bool
	includeThirdParty bool
	followRelative    bool
}

var pythonModuleExtensions = [...]string{".py", ".pyi"}
var pythonPackageInitFiles = [...]string{"__init__.py", "__init__.pyi"}

// ModuleAnalysisOptions configures module analysis behavior
type ModuleAnalysisOptions struct {
	ProjectRoot       string   // Project root directory
	PythonPath        []string // Additional Python path entries
	ExcludePatterns   []string // Module patterns to exclude; nil uses defaults, empty disables excludes
	IncludePatterns   []string // Module patterns to include; nil uses defaults, empty includes all files
	IncludeStdLib     *bool    // Include standard library dependencies
	IncludeThirdParty *bool    // Include third-party dependencies
	FollowRelative    *bool    // Follow relative imports
}

// DefaultModuleAnalysisOptions returns default analysis options
func DefaultModuleAnalysisOptions() *ModuleAnalysisOptions {
	return &ModuleAnalysisOptions{
		IncludePatterns:   domain.DefaultPythonModuleIncludePatterns(),
		ExcludePatterns:   domain.DefaultAnalysisExcludePatterns(),
		IncludeStdLib:     domain.BoolPtr(false),
		IncludeThirdParty: domain.BoolPtr(true),
		FollowRelative:    domain.BoolPtr(true),
	}
}

// NewModuleAnalyzer creates a new module analyzer
func NewModuleAnalyzer(options *ModuleAnalysisOptions) (*ModuleAnalyzer, error) {
	defaults := DefaultModuleAnalysisOptions()
	if options == nil {
		options = defaults
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
		reExportResolver:  NewReExportResolver(absRoot),
		includeStdLib:     domain.BoolValue(options.IncludeStdLib, domain.BoolValue(defaults.IncludeStdLib, false)),
		includeThirdParty: domain.BoolValue(options.IncludeThirdParty, domain.BoolValue(defaults.IncludeThirdParty, true)),
		followRelative:    domain.BoolValue(options.FollowRelative, domain.BoolValue(defaults.FollowRelative, true)),
	}

	if options.ExcludePatterns != nil {
		analyzer.excludePatterns = append(analyzer.excludePatterns, options.ExcludePatterns...)
	} else {
		analyzer.excludePatterns = append(analyzer.excludePatterns, defaults.ExcludePatterns...)
	}

	if options.IncludePatterns != nil {
		analyzer.includePatterns = append(analyzer.includePatterns, options.IncludePatterns...)
	} else {
		analyzer.includePatterns = append(analyzer.includePatterns, defaults.IncludePatterns...)
	}

	return analyzer, nil
}

// AnalyzeProject analyzes all Python modules in the project and builds a dependency graph
func (ma *ModuleAnalyzer) AnalyzeProject() (*DependencyGraph, error) {
	// Collect all Python files in the project
	files, err := ma.collectPythonFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to collect Python files: %w", err)
	}
	files = ma.canonicalModuleFiles(files)

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
	validFiles = ma.canonicalModuleFiles(validFiles)

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

	facts := ma.collectModuleFacts(result.AST)
	module.FunctionCount = facts.functionCount
	module.ClassCount = facts.classCount
	module.AbstractClassCount = facts.abstractClassCount
	module.PublicNames = facts.publicNames
	module.LineCount = countSourceLines(content)

	// Process each import
	for _, imp := range facts.imports {
		// Skip TYPE_CHECKING imports entirely: they never execute at runtime,
		// so they are not real dependencies for any analysis.
		if imp.IsTypeChecking {
			continue
		}

		// NOTE: lazy (function/method-body) imports are still recorded as edges
		// because they are real runtime dependencies that matter for coupling
		// metrics, dependency matrices, and architecture layer checks. They are
		// flagged via ImportInfo.IsLazy and excluded only from load-time
		// circular-dependency detection. See issue #460.

		targetModule := ma.resolveImport(imp, filePath)
		if targetModule == "" {
			continue
		}

		edgeType := ma.dependencyEdgeType(imp)
		for _, resolvedModule := range ma.importDependencyTargets(graph, imp, targetModule) {
			if ma.shouldSkipPackageInitDependency(filePath, moduleName, resolvedModule) {
				continue
			}
			if ma.shouldIncludeDependency(resolvedModule) {
				graph.AddDependency(moduleName, resolvedModule, edgeType, imp)
			}
		}
	}

	return nil
}

func (ma *ModuleAnalyzer) dependencyEdgeType(imp *ImportInfo) DependencyEdgeType {
	if imp.IsRelative {
		return DependencyEdgeRelative
	}
	if len(imp.ImportedNames) > 0 {
		return DependencyEdgeFromImport
	}
	return DependencyEdgeImport
}

func (ma *ModuleAnalyzer) importDependencyTargets(graph *DependencyGraph, imp *ImportInfo, targetModule string) []string {
	if len(imp.ImportedNames) == 0 {
		return []string{targetModule}
	}

	targets := make(map[string]bool, len(imp.ImportedNames))
	for _, importedName := range imp.ImportedNames {
		if importedName == "*" {
			targets[targetModule] = true
			continue
		}

		if !imp.IsRelative {
			if resolvedModule, found := ma.reExportResolver.ResolveReExport(targetModule, importedName); found {
				targets[resolvedModule] = true
				continue
			}
		}

		if concreteModule := importedNameModule(targetModule, importedName); graph.GetModule(concreteModule) != nil {
			targets[concreteModule] = true
			continue
		}

		targets[targetModule] = true
	}

	return sortedModuleNames(targets)
}

func importedNameModule(moduleName, importedName string) string {
	if moduleName == "" || importedName == "" {
		return ""
	}
	return moduleName + "." + importedName
}

func sortedModuleNames(moduleSet map[string]bool) []string {
	modules := make([]string, 0, len(moduleSet))
	for moduleName := range moduleSet {
		if moduleName != "" {
			modules = append(modules, moduleName)
		}
	}
	sort.Strings(modules)
	return modules
}

func (ma *ModuleAnalyzer) shouldSkipPackageInitDependency(filePath, moduleName, targetModule string) bool {
	return isPythonPackageInit(filePath) && strings.HasPrefix(targetModule, moduleName+".")
}

type moduleFacts struct {
	imports            []*ImportInfo
	functionCount      int
	classCount         int
	abstractClassCount int
	publicNames        []string
}

func (ma *ModuleAnalyzer) collectModuleFacts(ast *parser.Node) moduleFacts {
	var facts moduleFacts
	if ast == nil {
		return facts
	}

	ast.Walk(func(node *parser.Node) bool {
		switch node.Type {
		case parser.NodeImport, parser.NodeImportFrom:
			facts.imports = append(facts.imports, ma.importsFromNode(node)...)
		case parser.NodeFunctionDef, parser.NodeAsyncFunctionDef:
			facts.functionCount++
			if isPublicName(node.Name) {
				facts.publicNames = append(facts.publicNames, node.Name)
			}
		case parser.NodeClassDef:
			facts.classCount++
			if ma.isAbstractClass(node) {
				facts.abstractClassCount++
			}
			if isPublicName(node.Name) {
				facts.publicNames = append(facts.publicNames, node.Name)
			}
		}
		return true
	})

	return facts
}

func isPublicName(name string) bool {
	return name != "" && !strings.HasPrefix(name, "_")
}

func countSourceLines(content []byte) int {
	lines := 1
	for _, b := range content {
		if b == '\n' {
			lines++
		}
	}
	return lines
}

func (ma *ModuleAnalyzer) importsFromNode(node *parser.Node) []*ImportInfo {
	isTypeChecking := ma.isInTypeCheckingBlock(node)
	isLazy := ma.isInFunctionScope(node)

	switch node.Type {
	case parser.NodeImport:
		var imports []*ImportInfo
		if len(node.Children) > 0 {
			for _, child := range node.Children {
				if child.Type != parser.NodeAlias {
					continue
				}
				imp := &ImportInfo{
					Statement:      fmt.Sprintf("import %s", child.Name),
					ImportedNames:  []string{child.Name},
					IsRelative:     false,
					Line:           node.Location.StartLine,
					IsTypeChecking: isTypeChecking,
					IsLazy:         isLazy,
				}
				if alias, ok := child.Value.(string); ok {
					imp.Alias = alias
				}
				imports = append(imports, imp)
			}
			return imports
		}

		imports = make([]*ImportInfo, 0, len(node.Names))
		for _, name := range node.Names {
			imports = append(imports, &ImportInfo{
				Statement:      fmt.Sprintf("import %s", name),
				ImportedNames:  []string{name},
				IsRelative:     false,
				Line:           node.Location.StartLine,
				IsTypeChecking: isTypeChecking,
				IsLazy:         isLazy,
			})
		}
		return imports

	case parser.NodeImportFrom:
		module := node.Module
		level := node.Level

		seenNames := make(map[string]bool, len(node.Names)+len(node.Children))
		importedNames := make([]string, 0, len(node.Names)+len(node.Children))
		addName := func(name string) {
			if name == "" || seenNames[name] {
				return
			}
			seenNames[name] = true
			importedNames = append(importedNames, name)
		}
		for _, name := range node.Names {
			addName(name)
		}
		for _, child := range node.Children {
			if child.Type == parser.NodeAlias {
				addName(child.Name)
			}
		}

		imp := &ImportInfo{
			Statement:      ma.buildImportStatement(node),
			ImportedNames:  importedNames,
			IsRelative:     level > 0,
			Level:          level,
			Line:           node.Location.StartLine,
			IsTypeChecking: isTypeChecking,
			IsLazy:         isLazy,
		}

		if imp.IsRelative {
			imp.Statement = strings.TrimLeft(module, ".")
		} else {
			imp.Statement = module
		}

		return []*ImportInfo{imp}
	default:
		return nil
	}
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
	for i := 1; i < imp.Level; i++ {
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

	cacheKey := pythonPathImportCacheKey(moduleName)

	// Check cache first
	if resolved, exists := ma.resolvedModules[cacheKey]; exists {
		return resolved
	}

	// Try to find the module in the Python path
	for _, pathEntry := range ma.pythonPath {
		modulePath := filepath.Join(pathEntry, strings.ReplaceAll(moduleName, ".", string(filepath.Separator)))

		if initFile := ma.resolvePackageInit(modulePath); initFile != "" {
			ma.resolvedModules[cacheKey] = moduleName
			return moduleName
		}

		if moduleFile := ma.resolveModuleFile(modulePath); moduleFile != "" {
			ma.resolvedModules[cacheKey] = moduleName
			return moduleName
		}
	}

	// Check if it's a standard library or third-party module
	if ma.isStandardLibrary(moduleName) {
		if ma.includeStdLib {
			ma.resolvedModules[cacheKey] = moduleName
			return moduleName
		}
		return ""
	}

	if ma.includeThirdParty {
		ma.resolvedModules[cacheKey] = moduleName
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

	cacheKey := projectImportCacheKey(moduleName, fromFile)

	// Check cache first
	if resolved, exists := ma.resolvedModules[cacheKey]; exists {
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

		if moduleFile := ma.resolveModuleFile(modulePath); moduleFile != "" {
			// Calculate the module name based on project structure
			resolvedName := ma.filePathToModuleName(moduleFile)
			if ma.importMatchesResolvedModule(moduleName, resolvedName) {
				ma.resolvedModules[cacheKey] = resolvedName
				return resolvedName
			}
		}

		if initFile := ma.resolvePackageInit(modulePath); initFile != "" {
			resolvedName := ma.filePathToModuleName(initFile)
			if resolvedName != "" {
				// For __init__ files, use the package name.
				resolvedName = strings.TrimSuffix(resolvedName, ".__init__")
				if ma.importMatchesResolvedModule(moduleName, resolvedName) {
					ma.resolvedModules[cacheKey] = resolvedName
					return resolvedName
				}
			}
		}
	}

	// Fall back to the original resolveAbsoluteImport logic
	return ma.resolveAbsoluteImport(imp)
}

// importMatchesResolvedModule verifies that an absolute import maps by
// qualified module path, while still supporting script-style local imports for
// non-stdlib modules. Bare stdlib imports must not bind to same-basename
// project modules discovered under the current file's directory.
func (ma *ModuleAnalyzer) importMatchesResolvedModule(importName, resolvedName string) bool {
	if importName == "" || resolvedName == "" {
		return false
	}
	if resolvedName == importName {
		return true
	}
	if !strings.Contains(importName, ".") {
		return !ma.isStandardLibrary(importName) && strings.HasSuffix(resolvedName, "."+importName)
	}
	return strings.HasSuffix(resolvedName, "."+importName)
}

func projectImportCacheKey(moduleName, fromFile string) string {
	dir := filepath.Dir(fromFile)
	if absDir, err := filepath.Abs(dir); err == nil {
		dir = absDir
	}
	return "project\x00" + dir + "\x00" + moduleName
}

func pythonPathImportCacheKey(moduleName string) string {
	return "pythonpath\x00" + moduleName
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

		// Check if file is a Python module and matches include patterns
		if ma.isValidPythonFile(path) && ma.matchesIncludePatterns(path) && !ma.matchesExcludePatterns(path) {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func (ma *ModuleAnalyzer) canonicalModuleFiles(filePaths []string) []string {
	selected := make(map[string]string, len(filePaths))
	for _, filePath := range filePaths {
		moduleName := ma.filePathToModuleName(filePath)
		if moduleName == "" {
			continue
		}
		current, exists := selected[moduleName]
		if !exists || preferPythonModuleFile(filePath, current) {
			selected[moduleName] = filePath
		}
	}

	moduleNames := make([]string, 0, len(selected))
	for moduleName := range selected {
		moduleNames = append(moduleNames, moduleName)
	}
	sort.Strings(moduleNames)

	files := make([]string, 0, len(moduleNames))
	for _, moduleName := range moduleNames {
		files = append(files, selected[moduleName])
	}
	return files
}

// filePathToModuleName converts a file path to a Python module name
func (ma *ModuleAnalyzer) filePathToModuleName(filePath string) string {
	// Make path relative to project root
	relPath, err := filepath.Rel(ma.projectRoot, filePath)
	if err != nil {
		return ""
	}

	relPath = stripPythonModuleExtension(relPath)

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

func (ma *ModuleAnalyzer) isAbstractClass(classNode *parser.Node) bool {
	for _, base := range classNode.Bases {
		if ma.isAbstractClassName(ma.nodeQualifiedName(base)) {
			return true
		}
	}

	for _, child := range classNode.Body {
		if ma.isAbstractMethod(child) {
			return true
		}
	}

	return false
}

func (ma *ModuleAnalyzer) isAbstractMethod(node *parser.Node) bool {
	if node == nil || (node.Type != parser.NodeFunctionDef && node.Type != parser.NodeAsyncFunctionDef) {
		return false
	}

	for _, decorator := range node.Decorator {
		if ma.nodeQualifiedName(decorator) == "abstractmethod" || strings.HasSuffix(ma.nodeQualifiedName(decorator), ".abstractmethod") {
			return true
		}
	}

	return false
}

func (ma *ModuleAnalyzer) isAbstractClassName(name string) bool {
	switch name {
	case "ABC", "abc.ABC", "ABCMeta", "abc.ABCMeta":
		return true
	default:
		return false
	}
}

func (ma *ModuleAnalyzer) nodeQualifiedName(node *parser.Node) string {
	if node == nil {
		return ""
	}

	switch node.Type {
	case parser.NodeDecorator:
		if value, ok := node.Value.(*parser.Node); ok {
			return ma.nodeQualifiedName(value)
		}
	case parser.NodeCall:
		if value, ok := node.Value.(*parser.Node); ok {
			return ma.nodeQualifiedName(value)
		}
	case parser.NodeKeyword:
		if node.Name == "metaclass" {
			if value, ok := node.Value.(*parser.Node); ok {
				return ma.nodeQualifiedName(value)
			}
		}
	case parser.NodeKeywordArgument:
		if len(node.Children) >= 3 && ma.nodeQualifiedName(node.Children[0]) == "metaclass" {
			return ma.nodeQualifiedName(node.Children[2])
		}
	case parser.NodeName:
		return node.Name
	case parser.NodeAttribute:
		left := ""
		if value, ok := node.Value.(*parser.Node); ok {
			left = ma.nodeQualifiedName(value)
		}
		if left == "" && node.Left != nil {
			left = ma.nodeQualifiedName(node.Left)
		}
		if left == "" {
			return node.Name
		}
		return left + "." + node.Name
	}

	return ""
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
	return hasPythonModuleExtension(filePath)
}

// fileExists checks if a file exists
func (ma *ModuleAnalyzer) fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func (ma *ModuleAnalyzer) resolveModuleFile(modulePath string) string {
	for _, ext := range pythonModuleExtensions {
		filePath := modulePath + ext
		if ma.fileExists(filePath) {
			return filePath
		}
	}
	return ""
}

func (ma *ModuleAnalyzer) resolvePackageInit(packagePath string) string {
	for _, name := range pythonPackageInitFiles {
		filePath := filepath.Join(packagePath, name)
		if ma.fileExists(filePath) {
			return filePath
		}
	}
	return ""
}

func hasPythonModuleExtension(filePath string) bool {
	for _, ext := range pythonModuleExtensions {
		if strings.HasSuffix(filePath, ext) {
			return true
		}
	}
	return false
}

func stripPythonModuleExtension(filePath string) string {
	for _, ext := range pythonModuleExtensions {
		if strings.HasSuffix(filePath, ext) {
			return strings.TrimSuffix(filePath, ext)
		}
	}
	return filePath
}

func preferPythonModuleFile(candidate, current string) bool {
	candidatePriority := pythonModuleFilePriority(candidate)
	currentPriority := pythonModuleFilePriority(current)
	if candidatePriority != currentPriority {
		return candidatePriority < currentPriority
	}
	return candidate < current
}

func pythonModuleFilePriority(filePath string) int {
	switch {
	case strings.HasSuffix(filePath, ".py"):
		return 0
	case strings.HasSuffix(filePath, ".pyi"):
		return 1
	default:
		return 2
	}
}

func isPythonPackageInit(filePath string) bool {
	base := filepath.Base(filePath)
	for _, name := range pythonPackageInitFiles {
		if base == name {
			return true
		}
	}
	return false
}

// matchesIncludePatterns checks if path matches any include pattern
func (ma *ModuleAnalyzer) matchesIncludePatterns(path string) bool {
	if len(ma.includePatterns) == 0 {
		return true
	}
	for _, pattern := range ma.includePatterns {
		if matchPathPattern(pattern, ma.projectRoot, path) {
			return true
		}
	}
	return false
}

// matchesExcludePatterns checks if path matches any exclude pattern
func (ma *ModuleAnalyzer) matchesExcludePatterns(path string) bool {
	for _, pattern := range ma.excludePatterns {
		if matchPathPattern(pattern, ma.projectRoot, path) {
			return true
		}
	}
	return false
}

func matchPathPattern(pattern, root, path string) bool {
	for _, candidate := range pathPatternCandidates(root, path) {
		if matched, _ := doublestar.Match(pattern, candidate); matched {
			return true
		}
	}
	return false
}

func pathPatternCandidates(root, path string) []string {
	candidates := []string{
		filepath.ToSlash(path),
		filepath.Base(path),
	}
	if rel, err := filepath.Rel(root, path); err == nil {
		candidates = append(candidates, filepath.ToSlash(rel))
	}
	return candidates
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
		"time": true, "socket": true, "subprocess": true, "multiprocessing": true,
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

// isInTypeCheckingBlock checks if a node is inside a TYPE_CHECKING conditional block
func (ma *ModuleAnalyzer) isInTypeCheckingBlock(node *parser.Node) bool {
	child := node
	current := node.Parent
	for current != nil {
		if current.Type == parser.NodeIf || current.Type == parser.NodeElifClause {
			if ma.isTypeCheckingCondition(current.Test) && containsDirectNode(current.Body, child) {
				return true
			}
		}
		child = current
		current = current.Parent
	}
	return false
}

// isInFunctionScope reports whether the node is nested inside a function or
// method body. Such imports are "lazy": they run only when the function is
// called, not at module load time, so they cannot create a load-time circular
// dependency. A class body, by contrast, executes at definition (load) time,
// so imports directly in a class body are not considered lazy.
func (ma *ModuleAnalyzer) isInFunctionScope(node *parser.Node) bool {
	for current := node.Parent; current != nil; current = current.Parent {
		switch current.Type {
		case parser.NodeFunctionDef, parser.NodeAsyncFunctionDef:
			return true
		}
	}
	return false
}

func containsDirectNode(nodes []*parser.Node, target *parser.Node) bool {
	for _, node := range nodes {
		if node == target {
			return true
		}
	}
	return false
}

func (ma *ModuleAnalyzer) isTypeCheckingCondition(expr *parser.Node) bool {
	if expr == nil {
		return false
	}

	if expr.Type == parser.NodeName && expr.Name == "TYPE_CHECKING" {
		return true
	}

	if expr.Type == parser.NodeAttribute && expr.Name == "TYPE_CHECKING" {
		return true
	}

	if expr.Type == parser.NodeBoolOp {
		return ma.isTypeCheckingBoolOp(expr)
	}

	return false
}

func (ma *ModuleAnalyzer) isTypeCheckingBoolOp(expr *parser.Node) bool {
	children := expr.GetChildren()
	if len(children) == 0 {
		return false
	}

	switch expr.Op {
	case "and":
		for _, child := range children {
			if ma.isTypeCheckingCondition(child) {
				return true
			}
		}
	case "or":
		for _, child := range children {
			if !ma.isTypeCheckingCondition(child) {
				return false
			}
		}
		return true
	}

	return false
}
