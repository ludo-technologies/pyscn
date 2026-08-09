package service

import (
	"fmt"

	"github.com/ludo-technologies/pyscn/domain"
)

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
