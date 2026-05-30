package service

import (
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockDataFormatter_FormatHTML_EscapesFindingStrings(t *testing.T) {
	formatter := NewMockDataFormatter()
	response := &domain.MockDataResponse{
		Files: []domain.FileMockData{
			{
				FilePath: `<script>file</script>.py`,
				Findings: []domain.MockDataFinding{
					{
						Location:  domain.MockDataLocation{StartLine: 4},
						Type:      domain.MockDataTypeKeyword,
						Severity:  domain.MockDataSeverityWarning,
						Value:     `<script>alert("x")</script>`,
						Rationale: `bad <value>`,
					},
				},
			},
		},
		Summary:     domain.MockDataSummary{TotalFiles: 1, TotalFindings: 1, WarningFindings: 1},
		GeneratedAt: `<now>`,
		Version:     `<version>`,
	}

	output, err := formatter.Format(response, domain.OutputFormatHTML)
	require.NoError(t, err)

	assert.NotContains(t, output, `<script>alert("x")</script>`)
	assert.NotContains(t, output, `<script>file</script>.py`)
	assert.Contains(t, output, `&lt;script&gt;alert(&#34;x&#34;)&lt;/script&gt;`)
	assert.Contains(t, output, `bad &lt;value&gt;`)
}
