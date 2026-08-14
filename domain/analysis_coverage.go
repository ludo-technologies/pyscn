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
	// AnalysisKindComplexity identifies complexity analysis.
	AnalysisKindComplexity AnalysisKind = "complexity"
	// AnalysisKindDeadCode identifies dead-code analysis.
	AnalysisKindDeadCode AnalysisKind = "deadcode"
	// AnalysisKindClones identifies clone analysis.
	AnalysisKindClones AnalysisKind = "clones"
	// AnalysisKindCBO identifies class-coupling analysis.
	AnalysisKindCBO AnalysisKind = "cbo"
	// AnalysisKindLCOM identifies class-cohesion analysis.
	AnalysisKindLCOM AnalysisKind = "lcom"
	// AnalysisKindSystem identifies dependency and architecture analysis.
	AnalysisKindSystem AnalysisKind = "system"
	// AnalysisKindCommunities identifies module-community analysis.
	AnalysisKindCommunities AnalysisKind = "communities"
	// AnalysisKindMockData identifies mock-data analysis.
	AnalysisKindMockData AnalysisKind = "mockdata"
	// AnalysisKindDI identifies dependency-injection anti-pattern analysis.
	AnalysisKindDI AnalysisKind = "di"
)

// AnalysisFailureCode identifies a failure produced while an analyzer runs.
type AnalysisFailureCode string

// AnalysisFailureCodeExecution identifies an analyzer execution failure.
const AnalysisFailureCodeExecution AnalysisFailureCode = "execution_error"

// AnalysisFailure records an analyzer failure independently of project
// discovery and parsing diagnostics.
type AnalysisFailure struct {
	Analysis AnalysisKind        `json:"analysis" yaml:"analysis"`
	Code     AnalysisFailureCode `json:"code" yaml:"code"`
	Message  string              `json:"message" yaml:"message"`
	FilePath string              `json:"file_path,omitempty" yaml:"file_path,omitempty"`
	cause    error
}

// NewAnalysisFailure creates a public typed failure while retaining its
// underlying cause for in-process error inspection.
func NewAnalysisFailure(analysis AnalysisKind, code AnalysisFailureCode, filePath, message string, cause error) AnalysisFailure {
	return AnalysisFailure{
		Analysis: analysis,
		Code:     code,
		Message:  message,
		FilePath: filePath,
		cause:    cause,
	}
}

// Error implements error without changing the serialized failure contract.
func (f AnalysisFailure) Error() string {
	return f.Message
}

// Unwrap exposes the retained analyzer cause to errors.Is and errors.As.
func (f AnalysisFailure) Unwrap() error {
	return f.cause
}

// AnalysisFailureReporter is implemented by analyzer responses that can carry
// partial results alongside typed execution failures.
type AnalysisFailureReporter interface {
	AnalysisFailures() []AnalysisFailure
}

// AnalysisFailures returns complexity failures for aggregate analysis.
func (r *ComplexityResponse) AnalysisFailures() []AnalysisFailure {
	if r == nil {
		return nil
	}
	return r.Failures
}

// AnalysisFailures returns dead-code failures for aggregate analysis.
func (r *DeadCodeResponse) AnalysisFailures() []AnalysisFailure {
	if r == nil {
		return nil
	}
	return r.Failures
}

// AnalysisFailures returns clone failures for aggregate analysis.
func (r *CloneResponse) AnalysisFailures() []AnalysisFailure {
	if r == nil {
		return nil
	}
	return r.Failures
}

// AnalysisFailures returns CBO failures for aggregate analysis.
func (r *CBOResponse) AnalysisFailures() []AnalysisFailure {
	if r == nil {
		return nil
	}
	return r.Failures
}

// AnalysisFailures returns LCOM failures for aggregate analysis.
func (r *LCOMResponse) AnalysisFailures() []AnalysisFailure {
	if r == nil {
		return nil
	}
	return r.Failures
}

// AnalysisFailures returns system-analysis failures for aggregate analysis.
func (r *SystemAnalysisResponse) AnalysisFailures() []AnalysisFailure {
	if r == nil {
		return nil
	}
	return r.Failures
}

// AnalysisFailures returns community-analysis failures for aggregate analysis.
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
