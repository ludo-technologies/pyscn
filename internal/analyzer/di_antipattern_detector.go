package analyzer

import (
	"sort"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// DIAntipatternDetector coordinates all DI anti-pattern detectors
type DIAntipatternDetector struct {
	constructorAnalyzer    *ConstructorAnalyzer
	hiddenDepDetector      *HiddenDependencyDetector
	concreteDepDetector    *ConcreteDependencyDetector
	serviceLocatorDetector *ServiceLocatorDetector
	minSeverity            domain.DIAntipatternSeverity
}

// DIAntipatternOptions configures DI anti-pattern detection
type DIAntipatternOptions struct {
	ConstructorParamThreshold int
	MinSeverity               domain.DIAntipatternSeverity
}

// DefaultDIAntipatternOptions returns default options
func DefaultDIAntipatternOptions() *DIAntipatternOptions {
	return &DIAntipatternOptions{
		ConstructorParamThreshold: domain.DefaultDIConstructorParamThreshold,
		MinSeverity:               domain.DIAntipatternSeverityWarning,
	}
}

// NewDIAntipatternDetector creates a new DI anti-pattern detector
func NewDIAntipatternDetector(options *DIAntipatternOptions) *DIAntipatternDetector {
	if options == nil {
		options = DefaultDIAntipatternOptions()
	}

	return &DIAntipatternDetector{
		constructorAnalyzer:    NewConstructorAnalyzer(options.ConstructorParamThreshold),
		hiddenDepDetector:      NewHiddenDependencyDetector(),
		concreteDepDetector:    NewConcreteDependencyDetector(),
		serviceLocatorDetector: NewServiceLocatorDetector(),
		minSeverity:            options.MinSeverity,
	}
}

// Analyze runs all DI anti-pattern detectors on the given AST
func (d *DIAntipatternDetector) Analyze(ast *parser.Node, filePath string) ([]domain.DIAntipatternFinding, error) {
	if ast == nil {
		return nil, nil
	}

	var allFindings []domain.DIAntipatternFinding

	// Run constructor analyzer
	constructorFindings := d.constructorAnalyzer.Analyze(ast, filePath)
	allFindings = append(allFindings, constructorFindings...)

	// Run hidden dependency detector
	hiddenDepFindings := d.hiddenDepDetector.Analyze(ast, filePath)
	allFindings = append(allFindings, hiddenDepFindings...)

	// Run concrete dependency detector
	concreteDepFindings := d.concreteDepDetector.Analyze(ast, filePath)
	allFindings = append(allFindings, concreteDepFindings...)

	// Run service locator detector
	serviceLocatorFindings := d.serviceLocatorDetector.Analyze(ast, filePath)
	allFindings = append(allFindings, serviceLocatorFindings...)

	// Filter by minimum severity
	filteredFindings := d.filterBySeverity(allFindings)

	return filteredFindings, nil
}

// filterBySeverity filters findings by minimum severity
func (d *DIAntipatternDetector) filterBySeverity(findings []domain.DIAntipatternFinding) []domain.DIAntipatternFinding {
	minOrder := d.minSeverity.SeverityOrder()

	var filtered []domain.DIAntipatternFinding
	for _, finding := range findings {
		if finding.Severity.SeverityOrder() >= minOrder {
			filtered = append(filtered, finding)
		}
	}
	return filtered
}

// SortFindings sorts findings by the specified criteria
func SortFindings(findings []domain.DIAntipatternFinding, sortBy domain.SortCriteria) []domain.DIAntipatternFinding {
	sorted := make([]domain.DIAntipatternFinding, len(findings))
	copy(sorted, findings)

	switch sortBy {
	case domain.SortBySeverity:
		sort.Slice(sorted, func(i, j int) bool {
			// Higher severity first
			if sorted[i].Severity.SeverityOrder() != sorted[j].Severity.SeverityOrder() {
				return sorted[i].Severity.SeverityOrder() > sorted[j].Severity.SeverityOrder()
			}
			// Then by location
			if sorted[i].Location.FilePath != sorted[j].Location.FilePath {
				return sorted[i].Location.FilePath < sorted[j].Location.FilePath
			}
			return sorted[i].Location.StartLine < sorted[j].Location.StartLine
		})
	case domain.SortByName:
		sort.Slice(sorted, func(i, j int) bool {
			// By class name
			if sorted[i].ClassName != sorted[j].ClassName {
				return sorted[i].ClassName < sorted[j].ClassName
			}
			// Then by method name
			if sorted[i].MethodName != sorted[j].MethodName {
				return sorted[i].MethodName < sorted[j].MethodName
			}
			return sorted[i].Location.StartLine < sorted[j].Location.StartLine
		})
	case domain.SortByLocation:
		sort.Slice(sorted, func(i, j int) bool {
			if sorted[i].Location.FilePath != sorted[j].Location.FilePath {
				return sorted[i].Location.FilePath < sorted[j].Location.FilePath
			}
			return sorted[i].Location.StartLine < sorted[j].Location.StartLine
		})
	default:
		// Default to severity
		sort.Slice(sorted, func(i, j int) bool {
			if sorted[i].Severity.SeverityOrder() != sorted[j].Severity.SeverityOrder() {
				return sorted[i].Severity.SeverityOrder() > sorted[j].Severity.SeverityOrder()
			}
			if sorted[i].Location.FilePath != sorted[j].Location.FilePath {
				return sorted[i].Location.FilePath < sorted[j].Location.FilePath
			}
			return sorted[i].Location.StartLine < sorted[j].Location.StartLine
		})
	}

	return sorted
}

// GenerateSummary generates summary statistics from findings
func GenerateSummary(findings []domain.DIAntipatternFinding, filesAnalyzed int) domain.DIAntipatternSummary {
	summary := domain.DIAntipatternSummary{
		TotalFindings: len(findings),
		ByType:        make(map[domain.DIAntipatternType]int),
		BySeverity:    make(map[domain.DIAntipatternSeverity]int),
		FilesAnalyzed: filesAnalyzed,
	}

	// Track affected classes (classes with at least one finding)
	affectedClasses := make(map[string]bool)

	for _, finding := range findings {
		// Count by type
		summary.ByType[finding.Type]++

		// Count by severity
		summary.BySeverity[finding.Severity]++

		// Track affected classes
		if finding.ClassName != "" {
			affectedClasses[finding.ClassName] = true
		}
	}

	summary.AffectedClasses = len(affectedClasses)

	return summary
}

// CalculateDIAntipatterns is a convenience function for detecting DI anti-patterns with default options
func CalculateDIAntipatterns(ast *parser.Node, filePath string) ([]domain.DIAntipatternFinding, error) {
	detector := NewDIAntipatternDetector(DefaultDIAntipatternOptions())
	return detector.Analyze(ast, filePath)
}

// CalculateDIAntipatternsWithConfig detects DI anti-patterns with custom configuration
func CalculateDIAntipatternsWithConfig(ast *parser.Node, filePath string, options *DIAntipatternOptions) ([]domain.DIAntipatternFinding, error) {
	detector := NewDIAntipatternDetector(options)
	return detector.Analyze(ast, filePath)
}
