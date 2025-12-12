package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// PyprojectToml represents the structure of pyproject.toml
type PyprojectToml struct {
	Tool ToolConfig `toml:"tool"`
}

// ToolConfig represents the [tool] section
type ToolConfig struct {
	Pyscn PyprojectPyscnSection `toml:"pyscn"`
}

// PyprojectPyscnSection represents the [tool.pyscn] section in pyproject.toml
type PyprojectPyscnSection struct {
	Complexity     ComplexityTomlConfig     `toml:"complexity"`
	DeadCode       DeadCodeTomlConfig       `toml:"dead_code"`
	Output         OutputTomlConfig         `toml:"output"`
	Analysis       AnalysisTomlConfig       `toml:"analysis"`
	Cbo            CboTomlConfig            `toml:"cbo"`
	Architecture   ArchitectureTomlConfig   `toml:"architecture"`
	SystemAnalysis SystemAnalysisTomlConfig `toml:"system_analysis"`
	Dependencies   DependenciesTomlConfig   `toml:"dependencies"`
	Clones         ClonesConfig             `toml:"clones"`
}

// LoadPyprojectConfig loads pyscn configuration from pyproject.toml
func LoadPyprojectConfig(startDir string) (*PyscnConfig, error) {
	// Find pyproject.toml file (walk up directory tree)
	configPath, err := findPyprojectToml(startDir)
	if err != nil {
		// Return default config if no pyproject.toml found
		return DefaultPyscnConfig(), nil
	}

	// Read and parse pyproject.toml
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var pyproject PyprojectToml
	if err := toml.Unmarshal(data, &pyproject); err != nil {
		return nil, err
	}

	// Merge with defaults using shared merge logic
	config := DefaultPyscnConfig()
	mergeComplexitySection(config, &pyproject.Tool.Pyscn.Complexity)
	mergeDeadCodeSection(config, &pyproject.Tool.Pyscn.DeadCode)
	mergeOutputSection(config, &pyproject.Tool.Pyscn.Output)
	mergeAnalysisSection(config, &pyproject.Tool.Pyscn.Analysis)
	mergeCboSection(config, &pyproject.Tool.Pyscn.Cbo)
	mergeArchitectureSection(config, &pyproject.Tool.Pyscn.Architecture)
	mergeSystemAnalysisSection(config, &pyproject.Tool.Pyscn.SystemAnalysis)
	mergeDependenciesSection(config, &pyproject.Tool.Pyscn.Dependencies)
	mergeClonesSection(config, &pyproject.Tool.Pyscn.Clones)

	return config, nil
}

// mergeComplexitySection merges settings from the [complexity] section
// This function is shared between .pyscn.toml and pyproject.toml loaders
func mergeComplexitySection(defaults *PyscnConfig, complexity *ComplexityTomlConfig) {
	if complexity.LowThreshold != nil {
		defaults.ComplexityLowThreshold = *complexity.LowThreshold
	}
	if complexity.MediumThreshold != nil {
		defaults.ComplexityMediumThreshold = *complexity.MediumThreshold
	}
	if complexity.MaxComplexity != nil {
		defaults.ComplexityMaxComplexity = *complexity.MaxComplexity
	}
	if complexity.MinComplexity != nil {
		defaults.ComplexityMinComplexity = *complexity.MinComplexity
	}
}

// mergeClonesSection merges settings from the [clones] section
// This function is shared between .pyscn.toml and pyproject.toml loaders
func mergeClonesSection(defaults *PyscnConfig, clones *ClonesConfig) {
	// Analysis settings
	if clones.MinLines > 0 {
		defaults.Analysis.MinLines = clones.MinLines
	}
	if clones.MinNodes > 0 {
		defaults.Analysis.MinNodes = clones.MinNodes
	}
	if clones.MaxEditDistance > 0 {
		defaults.Analysis.MaxEditDistance = clones.MaxEditDistance
	}
	if clones.CostModelType != "" {
		defaults.Analysis.CostModelType = clones.CostModelType
	}
	if clones.IgnoreLiterals != nil {
		defaults.Analysis.IgnoreLiterals = clones.IgnoreLiterals
	}
	if clones.IgnoreIdentifiers != nil {
		defaults.Analysis.IgnoreIdentifiers = clones.IgnoreIdentifiers
	}
	if clones.SkipDocstrings != nil {
		defaults.Analysis.SkipDocstrings = clones.SkipDocstrings
	}
	if clones.EnableDFA != nil {
		defaults.Analysis.EnableDFA = clones.EnableDFA
	}

	// Thresholds
	if clones.Type1Threshold > 0 {
		defaults.Thresholds.Type1Threshold = clones.Type1Threshold
	}
	if clones.Type2Threshold > 0 {
		defaults.Thresholds.Type2Threshold = clones.Type2Threshold
	}
	if clones.Type3Threshold > 0 {
		defaults.Thresholds.Type3Threshold = clones.Type3Threshold
	}
	if clones.Type4Threshold > 0 {
		defaults.Thresholds.Type4Threshold = clones.Type4Threshold
	}
	if clones.SimilarityThreshold > 0 {
		defaults.Thresholds.SimilarityThreshold = clones.SimilarityThreshold
	}

	// Filtering
	if clones.MinSimilarity >= 0 {
		defaults.Filtering.MinSimilarity = clones.MinSimilarity
	}
	if clones.MaxSimilarity > 0 {
		defaults.Filtering.MaxSimilarity = clones.MaxSimilarity
	}
	if len(clones.EnabledCloneTypes) > 0 {
		defaults.Filtering.EnabledCloneTypes = clones.EnabledCloneTypes
	}
	if clones.MaxResults > 0 {
		defaults.Filtering.MaxResults = clones.MaxResults
	}

	// Grouping
	if clones.GroupingMode != "" {
		defaults.Grouping.Mode = clones.GroupingMode
	}
	if clones.GroupingThreshold > 0 {
		defaults.Grouping.Threshold = clones.GroupingThreshold
	}
	if clones.KCoreK > 0 {
		defaults.Grouping.KCoreK = clones.KCoreK
	}

	// LSH settings
	if clones.LSHEnabled != "" {
		defaults.LSH.Enabled = clones.LSHEnabled
	}
	if clones.LSHAutoThreshold > 0 {
		defaults.LSH.AutoThreshold = clones.LSHAutoThreshold
	}
	if clones.LSHSimilarityThreshold > 0 {
		defaults.LSH.SimilarityThreshold = clones.LSHSimilarityThreshold
	}
	if clones.LSHBands > 0 {
		defaults.LSH.Bands = clones.LSHBands
	}
	if clones.LSHRows > 0 {
		defaults.LSH.Rows = clones.LSHRows
	}
	if clones.LSHHashes > 0 {
		defaults.LSH.Hashes = clones.LSHHashes
	}

	// Performance
	if clones.MaxMemoryMB > 0 {
		defaults.Performance.MaxMemoryMB = clones.MaxMemoryMB
	}
	if clones.BatchSize > 0 {
		defaults.Performance.BatchSize = clones.BatchSize
	}
	if clones.EnableBatching != nil {
		defaults.Performance.EnableBatching = clones.EnableBatching
	}
	if clones.MaxGoroutines > 0 {
		defaults.Performance.MaxGoroutines = clones.MaxGoroutines
	}
	if clones.TimeoutSeconds > 0 {
		defaults.Performance.TimeoutSeconds = clones.TimeoutSeconds
	}

	// Input
	if len(clones.Paths) > 0 {
		defaults.Input.Paths = clones.Paths
	}
	if clones.Recursive != nil {
		defaults.Input.Recursive = clones.Recursive
	}
	if len(clones.IncludePatterns) > 0 {
		defaults.Input.IncludePatterns = clones.IncludePatterns
	}
	if len(clones.ExcludePatterns) > 0 {
		defaults.Input.ExcludePatterns = clones.ExcludePatterns
	}

	// Output
	if clones.ShowDetails != nil {
		defaults.Output.ShowDetails = clones.ShowDetails
	}
	if clones.ShowContent != nil {
		defaults.Output.ShowContent = clones.ShowContent
	}
	if clones.SortBy != "" {
		defaults.Output.SortBy = clones.SortBy
	}
	if clones.GroupClones != nil {
		defaults.Output.GroupClones = clones.GroupClones
	}
	if clones.Format != "" {
		defaults.Output.Format = clones.Format
	}
}

// mergeDeadCodeSection merges settings from the [dead_code] section
func mergeDeadCodeSection(defaults *PyscnConfig, deadCode *DeadCodeTomlConfig) {
	if deadCode.Enabled != nil {
		defaults.DeadCodeEnabled = deadCode.Enabled
	}
	if deadCode.MinSeverity != "" {
		defaults.DeadCodeMinSeverity = deadCode.MinSeverity
	}
	if deadCode.ShowContext != nil {
		defaults.DeadCodeShowContext = deadCode.ShowContext
	}
	if deadCode.ContextLines != nil {
		defaults.DeadCodeContextLines = *deadCode.ContextLines
	}
	if deadCode.SortBy != "" {
		defaults.DeadCodeSortBy = deadCode.SortBy
	}
	if deadCode.DetectAfterReturn != nil {
		defaults.DeadCodeDetectAfterReturn = deadCode.DetectAfterReturn
	}
	if deadCode.DetectAfterBreak != nil {
		defaults.DeadCodeDetectAfterBreak = deadCode.DetectAfterBreak
	}
	if deadCode.DetectAfterContinue != nil {
		defaults.DeadCodeDetectAfterContinue = deadCode.DetectAfterContinue
	}
	if deadCode.DetectAfterRaise != nil {
		defaults.DeadCodeDetectAfterRaise = deadCode.DetectAfterRaise
	}
	if deadCode.DetectUnreachableBranches != nil {
		defaults.DeadCodeDetectUnreachableBranches = deadCode.DetectUnreachableBranches
	}
	if len(deadCode.IgnorePatterns) > 0 {
		defaults.DeadCodeIgnorePatterns = deadCode.IgnorePatterns
	}
}

// mergeOutputSection merges settings from the [output] section
func mergeOutputSection(defaults *PyscnConfig, output *OutputTomlConfig) {
	if output.Format != "" {
		defaults.OutputFormat = output.Format
	}
	if output.ShowDetails != nil {
		defaults.OutputShowDetails = output.ShowDetails
	}
	if output.SortBy != "" {
		defaults.OutputSortBy = output.SortBy
	}
	if output.MinComplexity != nil {
		defaults.OutputMinComplexity = *output.MinComplexity
	}
	if output.Directory != "" {
		defaults.OutputDirectory = output.Directory
	}
}

// mergeAnalysisSection merges settings from the [analysis] section
func mergeAnalysisSection(defaults *PyscnConfig, analysis *AnalysisTomlConfig) {
	if len(analysis.IncludePatterns) > 0 {
		defaults.AnalysisIncludePatterns = analysis.IncludePatterns
	}
	if len(analysis.ExcludePatterns) > 0 {
		defaults.AnalysisExcludePatterns = analysis.ExcludePatterns
	}
	if analysis.Recursive != nil {
		defaults.AnalysisRecursive = analysis.Recursive
	}
	if analysis.FollowSymlinks != nil {
		defaults.AnalysisFollowSymlinks = analysis.FollowSymlinks
	}
}

// mergeCboSection merges settings from the [cbo] section
func mergeCboSection(defaults *PyscnConfig, cbo *CboTomlConfig) {
	if cbo.LowThreshold != nil {
		defaults.CboLowThreshold = *cbo.LowThreshold
	}
	if cbo.MediumThreshold != nil {
		defaults.CboMediumThreshold = *cbo.MediumThreshold
	}
	if cbo.MinCbo != nil {
		defaults.CboMinCbo = *cbo.MinCbo
	}
	if cbo.MaxCbo != nil {
		defaults.CboMaxCbo = *cbo.MaxCbo
	}
	if cbo.ShowZeros != nil {
		defaults.CboShowZeros = cbo.ShowZeros
	}
	if cbo.IncludeBuiltins != nil {
		defaults.CboIncludeBuiltins = cbo.IncludeBuiltins
	}
	if cbo.IncludeImports != nil {
		defaults.CboIncludeImports = cbo.IncludeImports
	}
}

// mergeArchitectureSection merges settings from the [architecture] section
func mergeArchitectureSection(defaults *PyscnConfig, arch *ArchitectureTomlConfig) {
	if arch.Enabled != nil {
		defaults.ArchitectureEnabled = arch.Enabled
	}
	if arch.ValidateLayers != nil {
		defaults.ArchitectureValidateLayers = arch.ValidateLayers
	}
	if arch.ValidateCohesion != nil {
		defaults.ArchitectureValidateCohesion = arch.ValidateCohesion
	}
	if arch.ValidateResponsibility != nil {
		defaults.ArchitectureValidateResponsibility = arch.ValidateResponsibility
	}
	if arch.MinCohesion != nil {
		defaults.ArchitectureMinCohesion = *arch.MinCohesion
	}
	if arch.MaxCoupling != nil {
		defaults.ArchitectureMaxCoupling = *arch.MaxCoupling
	}
	if arch.MaxResponsibilities != nil {
		defaults.ArchitectureMaxResponsibilities = *arch.MaxResponsibilities
	}
	if arch.LayerViolationSeverity != "" {
		defaults.ArchitectureLayerViolationSeverity = arch.LayerViolationSeverity
	}
	if arch.CohesionViolationSeverity != "" {
		defaults.ArchitectureCohesionViolationSeverity = arch.CohesionViolationSeverity
	}
	if arch.ResponsibilityViolationSeverity != "" {
		defaults.ArchitectureResponsibilityViolationSeverity = arch.ResponsibilityViolationSeverity
	}
	if arch.ShowAllViolations != nil {
		defaults.ArchitectureShowAllViolations = arch.ShowAllViolations
	}
	if arch.GroupByType != nil {
		defaults.ArchitectureGroupByType = arch.GroupByType
	}
	if arch.IncludeSuggestions != nil {
		defaults.ArchitectureIncludeSuggestions = arch.IncludeSuggestions
	}
	if arch.MaxViolationsToShow != nil {
		defaults.ArchitectureMaxViolationsToShow = *arch.MaxViolationsToShow
	}
	if len(arch.CustomPatterns) > 0 {
		defaults.ArchitectureCustomPatterns = arch.CustomPatterns
	}
	if len(arch.AllowedPatterns) > 0 {
		defaults.ArchitectureAllowedPatterns = arch.AllowedPatterns
	}
	if len(arch.ForbiddenPatterns) > 0 {
		defaults.ArchitectureForbiddenPatterns = arch.ForbiddenPatterns
	}
	if arch.StrictMode != nil {
		defaults.ArchitectureStrictMode = arch.StrictMode
	}
	if arch.FailOnViolations != nil {
		defaults.ArchitectureFailOnViolations = arch.FailOnViolations
	}
}

// mergeSystemAnalysisSection merges settings from the [system_analysis] section
func mergeSystemAnalysisSection(defaults *PyscnConfig, sa *SystemAnalysisTomlConfig) {
	if sa.Enabled != nil {
		defaults.SystemAnalysisEnabled = sa.Enabled
	}
	if sa.EnableDependencies != nil {
		defaults.SystemAnalysisEnableDependencies = sa.EnableDependencies
	}
	if sa.EnableArchitecture != nil {
		defaults.SystemAnalysisEnableArchitecture = sa.EnableArchitecture
	}
	if sa.UseComplexityData != nil {
		defaults.SystemAnalysisUseComplexityData = sa.UseComplexityData
	}
	if sa.UseClonesData != nil {
		defaults.SystemAnalysisUseClonesData = sa.UseClonesData
	}
	if sa.UseDeadCodeData != nil {
		defaults.SystemAnalysisUseDeadCodeData = sa.UseDeadCodeData
	}
	if sa.GenerateUnifiedReport != nil {
		defaults.SystemAnalysisGenerateUnifiedReport = sa.GenerateUnifiedReport
	}
}

// mergeDependenciesSection merges settings from the [dependencies] section
func mergeDependenciesSection(defaults *PyscnConfig, dep *DependenciesTomlConfig) {
	if dep.Enabled != nil {
		defaults.DependenciesEnabled = dep.Enabled
	}
	if dep.IncludeStdLib != nil {
		defaults.DependenciesIncludeStdLib = dep.IncludeStdLib
	}
	if dep.IncludeThirdParty != nil {
		defaults.DependenciesIncludeThirdParty = dep.IncludeThirdParty
	}
	if dep.FollowRelative != nil {
		defaults.DependenciesFollowRelative = dep.FollowRelative
	}
	if dep.DetectCycles != nil {
		defaults.DependenciesDetectCycles = dep.DetectCycles
	}
	if dep.CalculateMetrics != nil {
		defaults.DependenciesCalculateMetrics = dep.CalculateMetrics
	}
	if dep.FindLongChains != nil {
		defaults.DependenciesFindLongChains = dep.FindLongChains
	}
	if dep.MinCoupling != nil {
		defaults.DependenciesMinCoupling = *dep.MinCoupling
	}
	if dep.MaxCoupling != nil {
		defaults.DependenciesMaxCoupling = *dep.MaxCoupling
	}
	if dep.MinInstability != nil {
		defaults.DependenciesMinInstability = *dep.MinInstability
	}
	if dep.MaxDistance != nil {
		defaults.DependenciesMaxDistance = *dep.MaxDistance
	}
	if dep.SortBy != "" {
		defaults.DependenciesSortBy = dep.SortBy
	}
	if dep.ShowMatrix != nil {
		defaults.DependenciesShowMatrix = dep.ShowMatrix
	}
	if dep.ShowMetrics != nil {
		defaults.DependenciesShowMetrics = dep.ShowMetrics
	}
	if dep.ShowChains != nil {
		defaults.DependenciesShowChains = dep.ShowChains
	}
	if dep.GenerateDotGraph != nil {
		defaults.DependenciesGenerateDotGraph = dep.GenerateDotGraph
	}
	if dep.CycleReporting != "" {
		defaults.DependenciesCycleReporting = dep.CycleReporting
	}
	if dep.MaxCyclesToShow != nil {
		defaults.DependenciesMaxCyclesToShow = *dep.MaxCyclesToShow
	}
	if dep.ShowCyclePaths != nil {
		defaults.DependenciesShowCyclePaths = dep.ShowCyclePaths
	}
}

// findPyprojectToml walks up the directory tree to find pyproject.toml
func findPyprojectToml(startDir string) (string, error) {
	dir := startDir
	for {
		configPath := filepath.Join(dir, "pyproject.toml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	return "", os.ErrNotExist
}
