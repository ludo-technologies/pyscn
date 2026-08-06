package main

import "errors"

// Exit codes for the CLI. `pyscn check` documents this contract in its help
// text, and CI pipelines branch on it: a quality issue is a verdict about the
// code, while an analysis error means the verdict itself is untrustworthy.
const (
	// exitCodeQualityIssues means the analysis ran and found problems.
	exitCodeQualityIssues = 1
	// exitCodeAnalysisError means the analysis could not be completed over
	// the requested targets, so the result must not be read as a pass.
	exitCodeAnalysisError = 2
)

// analysisError marks an error as an analysis failure rather than a quality
// verdict, so main can map it to exitCodeAnalysisError.
type analysisError struct {
	err error
}

func newAnalysisError(err error) error {
	return &analysisError{err: err}
}

func (e *analysisError) Error() string {
	return e.err.Error()
}

func (e *analysisError) Unwrap() error {
	return e.err
}

// exitCodeFor maps a command error to the process exit code.
func exitCodeFor(err error) int {
	var analysisErr *analysisError
	if errors.As(err, &analysisErr) {
		return exitCodeAnalysisError
	}
	return exitCodeQualityIssues
}
