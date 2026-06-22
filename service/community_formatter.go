package service

import (
	"fmt"
	"io"

	"github.com/ludo-technologies/pyscn/domain"
)

// CommunityFormatter is a placeholder formatter until dedicated output lands in #584.
type CommunityFormatter struct{}

// NewCommunityFormatter creates a community analysis formatter.
func NewCommunityFormatter() *CommunityFormatter {
	return &CommunityFormatter{}
}

// Format returns a minimal text summary of community analysis results.
func (f *CommunityFormatter) Format(response *domain.CommunityAnalysisResult, format domain.OutputFormat) (string, error) {
	if response == nil {
		return "", fmt.Errorf("community analysis result is nil")
	}

	switch format {
	case domain.OutputFormatText:
		return fmt.Sprintf(
			"Community analysis: %d communities (modularity %.3f, algorithm %s)\n",
			response.TotalCommunities,
			response.Modularity,
			response.Algorithm,
		), nil
	default:
		return "", fmt.Errorf("community formatter does not yet support format %s", format)
	}
}

// Write formats and writes community analysis output.
func (f *CommunityFormatter) Write(response *domain.CommunityAnalysisResult, format domain.OutputFormat, writer io.Writer) error {
	content, err := f.Format(response, format)
	if err != nil {
		return err
	}
	_, err = io.WriteString(writer, content)
	return err
}
