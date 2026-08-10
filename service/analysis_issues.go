package service

import (
	"fmt"

	"github.com/ludo-technologies/pyscn/domain"
)

type analysisIssue struct {
	filePath string
	message  string
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
		failures = append(failures, domain.AnalysisFailure{
			Analysis: kind,
			Code:     domain.AnalysisFailureCodeExecution,
			Message:  issue.message,
			FilePath: issue.filePath,
		})
	}
	return failures
}
