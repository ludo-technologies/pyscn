package analyzer

import (
	"context"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

func TestConstructorAnalyzer(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		threshold     int
		wantFindings  int
		wantClassName string
	}{
		{
			name: "under threshold",
			code: `
class Service:
    def __init__(self, repo, logger, config):
        self.repo = repo
        self.logger = logger
        self.config = config
`,
			threshold:    5,
			wantFindings: 0,
		},
		{
			name: "at threshold",
			code: `
class Service:
    def __init__(self, a, b, c, d, e):
        pass
`,
			threshold:    5,
			wantFindings: 0,
		},
		{
			name: "over threshold",
			code: `
class Service:
    def __init__(self, a, b, c, d, e, f):
        pass
`,
			threshold:     5,
			wantFindings:  1,
			wantClassName: "Service",
		},
		{
			name: "way over threshold",
			code: `
class BadService:
    def __init__(self, a, b, c, d, e, f, g, h, i):
        pass
`,
			threshold:     5,
			wantFindings:  1,
			wantClassName: "BadService",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New()
			result, err := p.Parse(context.Background(), []byte(tt.code))
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			analyzer := NewConstructorAnalyzer(tt.threshold)
			findings := analyzer.Analyze(result.AST, "test.py")

			if len(findings) != tt.wantFindings {
				t.Errorf("got %d findings, want %d", len(findings), tt.wantFindings)
			}

			if tt.wantClassName != "" && len(findings) > 0 {
				if findings[0].ClassName != tt.wantClassName {
					t.Errorf("got class name %q, want %q", findings[0].ClassName, tt.wantClassName)
				}
			}
		})
	}
}

func TestConstructorAnalyzer_ArgsKwargs(t *testing.T) {
	tests := []struct {
		name         string
		code         string
		threshold    int
		wantFindings int
	}{
		{
			name: "args only should not count as over-injection",
			code: `
class Service:
    def __init__(self, *args):
        self.args = args
`,
			threshold:    5,
			wantFindings: 0,
		},
		{
			name: "kwargs only should not count as over-injection",
			code: `
class Service:
    def __init__(self, **kwargs):
        self.kwargs = kwargs
`,
			threshold:    5,
			wantFindings: 0,
		},
		{
			name: "mixed params with args/kwargs under threshold",
			code: `
class Service:
    def __init__(self, a, b, *args, **kwargs):
        pass
`,
			threshold:    5,
			wantFindings: 0,
		},
		{
			name: "many params with args/kwargs over threshold",
			code: `
class Service:
    def __init__(self, a, b, c, d, e, f, *args, **kwargs):
        pass
`,
			threshold:    5,
			wantFindings: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New()
			result, err := p.Parse(context.Background(), []byte(tt.code))
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			analyzer := NewConstructorAnalyzer(tt.threshold)
			findings := analyzer.Analyze(result.AST, "test.py")

			if len(findings) != tt.wantFindings {
				t.Errorf("got %d findings, want %d", len(findings), tt.wantFindings)
			}
		})
	}
}

func TestConstructorAnalyzer_NestedClasses(t *testing.T) {
	code := `
class Outer:
    def __init__(self, a, b, c):
        pass

    class Inner:
        def __init__(self, x, y, z, w, v, u):
            pass
`
	p := parser.New()
	result, err := p.Parse(context.Background(), []byte(code))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	analyzer := NewConstructorAnalyzer(5)
	findings := analyzer.Analyze(result.AST, "test.py")

	// Should find over-injection in Inner class only
	if len(findings) != 1 {
		t.Errorf("got %d findings, want 1", len(findings))
	}

	if len(findings) > 0 && findings[0].ClassName != "Inner" {
		t.Errorf("got class name %q, want %q", findings[0].ClassName, "Inner")
	}
}

func TestHiddenDependencyDetector_GlobalStatement(t *testing.T) {
	code := `
_config = {}

class Service:
    def do_something(self):
        global _config
        return _config
`
	p := parser.New()
	result, err := p.Parse(context.Background(), []byte(code))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	detector := NewHiddenDependencyDetector()
	findings := detector.Analyze(result.AST, "test.py")

	// Should find global statement usage
	globalFindings := filterBySubtype(findings, string(domain.HiddenDepGlobal))
	if len(globalFindings) == 0 {
		t.Error("expected to find global statement usage")
	}
}

func TestHiddenDependencyDetector_Singleton(t *testing.T) {
	code := `
class Singleton:
    _instance = None

    def __new__(cls):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
        return cls._instance
`
	p := parser.New()
	result, err := p.Parse(context.Background(), []byte(code))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	detector := NewHiddenDependencyDetector()
	findings := detector.Analyze(result.AST, "test.py")

	// Should find singleton pattern
	singletonFindings := filterBySubtype(findings, string(domain.HiddenDepSingleton))
	if len(singletonFindings) == 0 {
		t.Error("expected to find singleton pattern")
	}
}

func TestConcreteDependencyDetector_TypeHint(t *testing.T) {
	code := `
class MySQLRepo:
    pass

class Service:
    def __init__(self, repo: MySQLRepo):
        self.repo = repo
`
	p := parser.New()
	result, err := p.Parse(context.Background(), []byte(code))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	detector := NewConcreteDependencyDetector()
	findings := detector.Analyze(result.AST, "test.py")

	// Should find concrete type hint
	typeHintFindings := filterBySubtype(findings, string(domain.ConcreteDepTypeHint))
	if len(typeHintFindings) == 0 {
		t.Error("expected to find concrete type hint")
	}
}

func TestConcreteDependencyDetector_DirectInstantiation(t *testing.T) {
	code := `
class Logger:
    pass

class Service:
    def __init__(self):
        self.logger = Logger()
`
	p := parser.New()
	result, err := p.Parse(context.Background(), []byte(code))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	detector := NewConcreteDependencyDetector()
	findings := detector.Analyze(result.AST, "test.py")

	// Should find direct instantiation
	instantiationFindings := filterBySubtype(findings, string(domain.ConcreteDepInstantiation))
	if len(instantiationFindings) == 0 {
		t.Error("expected to find direct instantiation")
	}
}

func TestConcreteDependencyDetector_AbstractTypeHint(t *testing.T) {
	code := `
class Service:
    def __init__(self, repo: AbstractRepo, logger: ILogger):
        self.repo = repo
        self.logger = logger
`
	p := parser.New()
	result, err := p.Parse(context.Background(), []byte(code))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	detector := NewConcreteDependencyDetector()
	findings := detector.Analyze(result.AST, "test.py")

	// Should NOT find concrete type hints (Abstract prefix, I prefix)
	typeHintFindings := filterBySubtype(findings, string(domain.ConcreteDepTypeHint))
	if len(typeHintFindings) > 0 {
		t.Errorf("unexpected findings for abstract types: %d", len(typeHintFindings))
	}
}

func TestServiceLocatorDetector(t *testing.T) {
	tests := []struct {
		name         string
		code         string
		wantFindings int
	}{
		{
			name: "container.get",
			code: `
class Service:
    def __init__(self):
        self.logger = container.get("logger")
`,
			wantFindings: 1,
		},
		{
			name: "resolve function",
			code: `
class Service:
    def process(self):
        return resolve("executor")
`,
			wantFindings: 1,
		},
		{
			name: "get_service function",
			code: `
class Service:
    def __init__(self):
        self.config = get_service("config")
`,
			wantFindings: 1,
		},
		{
			name: "no service locator",
			code: `
class Service:
    def __init__(self, logger, config):
        self.logger = logger
        self.config = config
`,
			wantFindings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New()
			result, err := p.Parse(context.Background(), []byte(tt.code))
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			detector := NewServiceLocatorDetector()
			findings := detector.Analyze(result.AST, "test.py")

			if len(findings) != tt.wantFindings {
				t.Errorf("got %d findings, want %d", len(findings), tt.wantFindings)
			}
		})
	}
}

func TestDIAntipatternDetector_Integration(t *testing.T) {
	code := `
_global_config = {}

class Container:
    @classmethod
    def get(cls, name):
        return None

class BadService:
    _instance = None

    def __init__(self, a, b, c, d, e, f, repo: ConcreteRepo):
        global _global_config
        self.logger = Container.get("logger")
        self.helper = Helper()
`
	p := parser.New()
	result, err := p.Parse(context.Background(), []byte(code))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	options := &DIAntipatternOptions{
		ConstructorParamThreshold: 5,
		MinSeverity:               domain.DIAntipatternSeverityInfo,
	}

	detector := NewDIAntipatternDetector(options)
	findings, err := detector.Analyze(result.AST, "test.py")
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}

	// Should find multiple anti-patterns
	if len(findings) < 3 {
		t.Errorf("expected at least 3 findings, got %d", len(findings))
	}

	// Verify we found different types
	types := make(map[domain.DIAntipatternType]bool)
	for _, f := range findings {
		types[f.Type] = true
	}

	if !types[domain.DIAntipatternConstructorOverInjection] {
		t.Error("expected to find constructor over-injection")
	}
}

func TestSortFindings(t *testing.T) {
	findings := []domain.DIAntipatternFinding{
		{
			Type:     domain.DIAntipatternConstructorOverInjection,
			Severity: domain.DIAntipatternSeverityWarning,
			Location: domain.SourceLocation{FilePath: "b.py", StartLine: 10},
		},
		{
			Type:     domain.DIAntipatternHiddenDependency,
			Severity: domain.DIAntipatternSeverityError,
			Location: domain.SourceLocation{FilePath: "a.py", StartLine: 5},
		},
		{
			Type:     domain.DIAntipatternConcreteDependency,
			Severity: domain.DIAntipatternSeverityInfo,
			Location: domain.SourceLocation{FilePath: "a.py", StartLine: 15},
		},
	}

	// Sort by severity (descending)
	sorted := SortFindings(findings, domain.SortBySeverity)

	if sorted[0].Severity != domain.DIAntipatternSeverityError {
		t.Error("expected error severity first")
	}
	if sorted[1].Severity != domain.DIAntipatternSeverityWarning {
		t.Error("expected warning severity second")
	}
	if sorted[2].Severity != domain.DIAntipatternSeverityInfo {
		t.Error("expected info severity last")
	}
}

func TestGenerateSummary(t *testing.T) {
	findings := []domain.DIAntipatternFinding{
		{
			Type:      domain.DIAntipatternConstructorOverInjection,
			Severity:  domain.DIAntipatternSeverityWarning,
			ClassName: "Service1",
		},
		{
			Type:      domain.DIAntipatternHiddenDependency,
			Severity:  domain.DIAntipatternSeverityError,
			ClassName: "Service1",
		},
		{
			Type:      domain.DIAntipatternServiceLocator,
			Severity:  domain.DIAntipatternSeverityWarning,
			ClassName: "Service2",
		},
	}

	summary := GenerateSummary(findings, 5)

	if summary.TotalFindings != 3 {
		t.Errorf("expected 3 total findings, got %d", summary.TotalFindings)
	}

	if summary.FilesAnalyzed != 5 {
		t.Errorf("expected 5 files analyzed, got %d", summary.FilesAnalyzed)
	}

	if summary.AffectedClasses != 2 {
		t.Errorf("expected 2 affected classes, got %d", summary.AffectedClasses)
	}

	if summary.ByType[domain.DIAntipatternConstructorOverInjection] != 1 {
		t.Error("expected 1 constructor over-injection finding")
	}

	if summary.BySeverity[domain.DIAntipatternSeverityWarning] != 2 {
		t.Error("expected 2 warning severity findings")
	}
}

// Helper function to filter findings by subtype
func filterBySubtype(findings []domain.DIAntipatternFinding, subtype string) []domain.DIAntipatternFinding {
	var filtered []domain.DIAntipatternFinding
	for _, f := range findings {
		if f.Subtype == subtype {
			filtered = append(filtered, f)
		}
	}
	return filtered
}
