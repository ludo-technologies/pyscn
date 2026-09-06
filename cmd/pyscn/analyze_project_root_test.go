package main

import (
	"bytes"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestAnalyzeCommandDoesNotExposeProjectRootFlag(t *testing.T) {
	command := NewAnalyzeCommand()
	flag := command.CreateCobraCommand().Flags().Lookup("project-root")
	assert.Nil(t, flag)
}

func TestAnalyzeSummaryPrintsSystemWarnings(t *testing.T) {
	command := NewAnalyzeCommand()
	root := &cobra.Command{}
	var output bytes.Buffer
	root.SetErr(&output)
	command.printSummary(root, &domain.AnalyzeResponse{
		System: &domain.SystemAnalysisResponse{Warnings: []string{"inferred project root differs from analyzed directory"}},
	})
	assert.Contains(t, output.String(), "inferred project root differs from analyzed directory")
}
