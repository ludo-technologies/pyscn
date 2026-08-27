package service

import (
	"fmt"

	"github.com/ludo-technologies/pyscn/domain"
)

type analysisIssue struct {
	filePath       string
	message        string
	cause          error
	diagnosticCode domain.DiagnosticCode
}

func projectFileDiagnostic(file *ProjectFile) (domain.AnalysisDiagnostic, bool) {
	if file == nil {
		return domain.AnalysisDiagnostic{Code: domain.DiagnosticCodeRead, Message: "invalid project file"}, true
	}
	if file.ReadErr != nil {
		return domain.AnalysisDiagnostic{FilePath: file.Path, Code: domain.DiagnosticCodeRead, Message: file.ReadErr.Error()}, true
	}
	if file.ParseErr != nil {
		return domain.AnalysisDiagnostic{FilePath: file.Path, Code: domain.DiagnosticCodeParse, Message: file.ParseErr.Error()}, true
	}
	if file.AST == nil || file.parseResult == nil {
		return domain.AnalysisDiagnostic{FilePath: file.Path, Code: domain.DiagnosticCodeParse, Message: "invalid parse result"}, true
	}
	return domain.AnalysisDiagnostic{}, false
}

func projectFileAnalysisIssue(file *ProjectFile) (analysisIssue, bool) {
	diagnostic, invalid := projectFileDiagnostic(file)
	if !invalid {
		return analysisIssue{}, false
	}
	var cause error
	if file != nil {
		if file.ReadErr != nil {
			cause = file.ReadErr
		} else {
			cause = file.ParseErr
		}
	}
	return analysisIssue{
		filePath:       diagnostic.FilePath,
		message:        fmt.Sprintf("%s: %s", diagnostic.Code, diagnostic.Message),
		cause:          cause,
		diagnosticCode: diagnostic.Code,
	}, true
}

func diagnosticMessages(diagnostics []domain.AnalysisDiagnostic) []string {
	messages := make([]string, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		messages = append(messages, fmt.Sprintf("[%s] %s: %s", diagnostic.FilePath, diagnostic.Code, diagnostic.Message))
	}
	return messages
}

func failureMessages(failures []domain.AnalysisFailure) []string {
	messages := make([]string, 0, len(failures))
	for _, failure := range failures {
		messages = append(messages, fmt.Sprintf("[%s] %s: %s", failure.FilePath, failure.Code, failure.Message))
	}
	return messages
}

func analysisIssueMessages(issues []analysisIssue) []string {
	messages := make([]string, 0, len(issues))
	for _, issue := range issues {
		if issue.filePath == "" {
			messages = append(messages, issue.message)
			continue
		}
		messages = append(messages, fmt.Sprintf("[%s] %s", issue.filePath, issue.message))
	}
	return messages
}

func analyzerFailures(kind domain.AnalysisKind, issues []analysisIssue) []domain.AnalysisFailure {
	failures := make([]domain.AnalysisFailure, 0, len(issues))
	for _, issue := range issues {
		if issue.diagnosticCode != "" {
			continue
		}
		failures = append(failures, domain.NewAnalysisFailure(
			kind,
			domain.AnalysisFailureCodeExecution,
			issue.filePath,
			issue.message,
			issue.cause,
		))
	}
	return failures
}
