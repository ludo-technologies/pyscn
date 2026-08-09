package domain

// DiagnosticCode identifies a project-level analysis failure category.
type DiagnosticCode string

const (
	// DiagnosticCodeRead identifies a source file that could not be read.
	DiagnosticCodeRead DiagnosticCode = "read_error"
	// DiagnosticCodeParse identifies source that could not be parsed.
	DiagnosticCodeParse DiagnosticCode = "parse_error"
)

// AnalysisDiagnostic records why a discovered source file was not analyzed.
type AnalysisDiagnostic struct {
	FilePath string         `json:"file_path" yaml:"file_path"`
	Code     DiagnosticCode `json:"code" yaml:"code"`
	Message  string         `json:"message" yaml:"message"`
}

// AnalysisKind identifies an analyzer in cross-analyzer results.
type AnalysisKind string

const (
	AnalysisKindComplexity  AnalysisKind = "complexity"
	AnalysisKindDeadCode    AnalysisKind = "deadcode"
	AnalysisKindClones      AnalysisKind = "clones"
	AnalysisKindCBO         AnalysisKind = "cbo"
	AnalysisKindLCOM        AnalysisKind = "lcom"
	AnalysisKindSystem      AnalysisKind = "system"
	AnalysisKindCommunities AnalysisKind = "communities"
	AnalysisKindMockData    AnalysisKind = "mockdata"
	AnalysisKindDI          AnalysisKind = "di"
)

// AnalysisFailureCode identifies a failure produced while an analyzer runs.
type AnalysisFailureCode string

const AnalysisFailureCodeExecution AnalysisFailureCode = "execution_error"

// AnalysisFailure records an analyzer failure independently of project
// discovery and parsing diagnostics.
type AnalysisFailure struct {
	Analysis AnalysisKind        `json:"analysis" yaml:"analysis"`
	Code     AnalysisFailureCode `json:"code" yaml:"code"`
	Message  string              `json:"message" yaml:"message"`
	FilePath string              `json:"file_path,omitempty" yaml:"file_path,omitempty"`
}

// AnalysisFailureReporter is implemented by analyzer responses that can carry
// partial results alongside typed execution failures.
type AnalysisFailureReporter interface {
	AnalysisFailures() []AnalysisFailure
}

func (r *ComplexityResponse) AnalysisFailures() []AnalysisFailure {
	if r == nil {
		return nil
	}
	return r.Failures
}

func (r *DeadCodeResponse) AnalysisFailures() []AnalysisFailure {
	if r == nil {
		return nil
	}
	return r.Failures
}

func (r *CloneResponse) AnalysisFailures() []AnalysisFailure {
	if r == nil {
		return nil
	}
	return r.Failures
}

func (r *CBOResponse) AnalysisFailures() []AnalysisFailure {
	if r == nil {
		return nil
	}
	return r.Failures
}

func (r *LCOMResponse) AnalysisFailures() []AnalysisFailure {
	if r == nil {
		return nil
	}
	return r.Failures
}

func (r *SystemAnalysisResponse) AnalysisFailures() []AnalysisFailure {
	if r == nil {
		return nil
	}
	return r.Failures
}

func (r *CommunityAnalysisResult) AnalysisFailures() []AnalysisFailure {
	if r == nil {
		return nil
	}
	return r.Failures
}

// AnalysisCoverage records how much of the discovered project was analyzed.
type AnalysisCoverage struct {
	TotalFiles    int
	AnalyzedFiles int
	SkippedFiles  int
	Diagnostics   []AnalysisDiagnostic
}
