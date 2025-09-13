package service

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/ludo-technologies/pyscn/domain"
)

// CloneOutputFormatter implements the domain.CloneOutputFormatter interface
type CloneOutputFormatter struct{}

// NewCloneOutputFormatter creates a new clone output formatter
func NewCloneOutputFormatter() *CloneOutputFormatter {
	return &CloneOutputFormatter{}
}

// FormatCloneResponse formats a clone response according to the specified format
func (f *CloneOutputFormatter) FormatCloneResponse(response *domain.CloneResponse, format domain.OutputFormat, writer io.Writer) error {
	switch format {
	case domain.OutputFormatText:
		return f.formatAsText(response, writer)
	case domain.OutputFormatJSON:
		return WriteJSON(writer, response)
	case domain.OutputFormatYAML:
		return WriteYAML(writer, response)
	case domain.OutputFormatCSV:
		return f.formatAsCSV(response, writer)
	case domain.OutputFormatHTML:
		return f.formatAsHTML(response, writer)
	default:
		return domain.NewUnsupportedFormatError(string(format))
	}
}

// FormatCloneStatistics formats clone statistics
func (f *CloneOutputFormatter) FormatCloneStatistics(stats *domain.CloneStatistics, format domain.OutputFormat, writer io.Writer) error {
	switch format {
	case domain.OutputFormatText:
		return f.formatStatsAsText(stats, writer)
	case domain.OutputFormatJSON:
		return WriteJSON(writer, stats)
	case domain.OutputFormatYAML:
		return WriteYAML(writer, stats)
	case domain.OutputFormatCSV:
		return f.formatStatsAsCSV(stats, writer)
	case domain.OutputFormatHTML:
		return f.formatStatsAsHTML(stats, writer)
	default:
		return domain.NewUnsupportedFormatError(string(format))
	}
}

// formatAsText formats the response as human-readable text
func (f *CloneOutputFormatter) formatAsText(response *domain.CloneResponse, writer io.Writer) error {
	if !response.Success {
		fmt.Fprintf(writer, "Clone detection failed: %s\n", response.Error)
		return nil
	}

	utils := NewFormatUtils()

	// Header
	fmt.Fprint(writer, utils.FormatMainHeader("Clone Detection Analysis Report"))

	// Summary
	if response.Statistics != nil {
		stats := map[string]interface{}{
			"Files Analyzed":     response.Statistics.FilesAnalyzed,
			"Lines Analyzed":     response.Statistics.LinesAnalyzed,
			"Clone Pairs":        response.Statistics.TotalClonePairs,
			"Clone Groups":       response.Statistics.TotalCloneGroups,
			"Average Similarity": fmt.Sprintf("%.3f", response.Statistics.AverageSimilarity),
			"Analysis Duration":  utils.FormatDuration(response.Duration),
		}
		fmt.Fprint(writer, utils.FormatSummaryStats(stats))
	}

	// Clone Types breakdown
	if response.Statistics != nil && len(response.Statistics.ClonesByType) > 0 {
		fmt.Fprint(writer, utils.FormatSectionHeader("CLONE TYPES"))
		for cloneType, count := range response.Statistics.ClonesByType {
			fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, cloneType, fmt.Sprintf("%d pairs", count)))
		}
		fmt.Fprint(writer, utils.FormatSectionSeparator())
	}

	if len(response.ClonePairs) == 0 {
		fmt.Fprint(writer, utils.FormatSectionHeader("RESULTS"))
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Status", "No clones detected"))
		return nil
	}

	// Detailed clone information
	if response.Request != nil && response.Request.GroupClones && len(response.CloneGroups) > 0 {
		fmt.Fprint(writer, utils.FormatSectionHeader("CLONE GROUPS"))

		for _, group := range response.CloneGroups {
			if group == nil {
				continue
			}
			fmt.Fprint(writer, utils.FormatLabelWithIndent(0, "Group", fmt.Sprintf("%d (%s, %d clones, similarity: %.3f)",
				group.ID, group.Type.String(), group.Size, group.Similarity)))

			for i, clone := range group.Clones {
				if clone == nil || clone.Location == nil {
					continue
				}
				fmt.Fprint(writer, utils.FormatLabelWithIndent(ItemPadding, fmt.Sprintf("Clone %d", i+1),
					fmt.Sprintf("%s (%d lines, %d nodes)", clone.Location.String(), clone.LineCount, clone.Size)))
			}
			fmt.Fprint(writer, "\n")
		}
	} else {
		fmt.Fprint(writer, utils.FormatSectionHeader("CLONE PAIRS"))

		for i, pair := range response.ClonePairs {
			if pair == nil {
				continue
			}
			fmt.Fprint(writer, utils.FormatLabelWithIndent(0, fmt.Sprintf("Pair %d", i+1),
				fmt.Sprintf("%s (similarity: %.3f, confidence: %.3f)", pair.Type.String(), pair.Similarity, pair.Confidence)))

			if pair.Clone1 != nil && pair.Clone1.Location != nil {
				fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Clone 1",
					fmt.Sprintf("%s (%d lines, %d nodes)", pair.Clone1.Location.String(), pair.Clone1.LineCount, pair.Clone1.Size)))
			}

			if pair.Clone2 != nil && pair.Clone2.Location != nil {
				fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Clone 2",
					fmt.Sprintf("%s (%d lines, %d nodes)", pair.Clone2.Location.String(), pair.Clone2.LineCount, pair.Clone2.Size)))
			}

			if response.Request != nil && response.Request.ShowContent && pair.Clone1 != nil && pair.Clone1.Content != "" {
				fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Preview", ""))
				lines := strings.Split(pair.Clone1.Content, "\n")
				for j, line := range lines {
					if j >= 5 { // Limit preview to 5 lines
						fmt.Fprintf(writer, "%s...\n", strings.Repeat(" ", ItemPadding+2))
						break
					}
					fmt.Fprintf(writer, "%s%s\n", strings.Repeat(" ", ItemPadding+2), line)
				}
			}

			fmt.Fprint(writer, "\n")
		}
	}

	return nil
}

// formatAsJSON formats the response as JSON
// formatAsCSV formats the response as CSV
func (f *CloneOutputFormatter) formatAsCSV(response *domain.CloneResponse, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header
	header := []string{
		"pair_id", "clone_type", "similarity", "confidence", "distance",
		"clone1_file", "clone1_start_line", "clone1_end_line", "clone1_size", "clone1_lines",
		"clone2_file", "clone2_start_line", "clone2_end_line", "clone2_size", "clone2_lines",
	}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write clone pairs
	for _, pair := range response.ClonePairs {
		record := []string{
			fmt.Sprintf("%d", pair.ID),
			pair.Type.String(),
			fmt.Sprintf("%.6f", pair.Similarity),
			fmt.Sprintf("%.6f", pair.Confidence),
			fmt.Sprintf("%.2f", pair.Distance),
			pair.Clone1.Location.FilePath,
			fmt.Sprintf("%d", pair.Clone1.Location.StartLine),
			fmt.Sprintf("%d", pair.Clone1.Location.EndLine),
			fmt.Sprintf("%d", pair.Clone1.Size),
			fmt.Sprintf("%d", pair.Clone1.LineCount),
			pair.Clone2.Location.FilePath,
			fmt.Sprintf("%d", pair.Clone2.Location.StartLine),
			fmt.Sprintf("%d", pair.Clone2.Location.EndLine),
			fmt.Sprintf("%d", pair.Clone2.Size),
			fmt.Sprintf("%d", pair.Clone2.LineCount),
		}
		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}

// formatStatsAsText formats statistics as text
func (f *CloneOutputFormatter) formatStatsAsText(stats *domain.CloneStatistics, writer io.Writer) error {
	fmt.Fprintf(writer, "Clone Detection Statistics\n")
	fmt.Fprintf(writer, "==========================\n\n")
	fmt.Fprintf(writer, "Files analyzed: %d\n", stats.FilesAnalyzed)
	fmt.Fprintf(writer, "Lines analyzed: %d\n", stats.LinesAnalyzed)
	fmt.Fprintf(writer, "Clone pairs: %d\n", stats.TotalClonePairs)
	fmt.Fprintf(writer, "Clone groups: %d\n", stats.TotalCloneGroups)

	if stats.AverageSimilarity > 0 {
		fmt.Fprintf(writer, "Average similarity: %.3f\n", stats.AverageSimilarity)
	}

	if len(stats.ClonesByType) > 0 {
		fmt.Fprintf(writer, "\nClone types:\n")
		for cloneType, count := range stats.ClonesByType {
			fmt.Fprintf(writer, "  %s: %d\n", cloneType, count)
		}
	}

	return nil
}

// formatStatsAsJSON and formatStatsAsYAML were replaced by shared helpers in format_utils.go

// formatStatsAsCSV formats statistics as CSV
func (f *CloneOutputFormatter) formatStatsAsCSV(stats *domain.CloneStatistics, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write basic statistics
	records := [][]string{
		{"metric", "value"},
		{"files_analyzed", fmt.Sprintf("%d", stats.FilesAnalyzed)},
		{"lines_analyzed", fmt.Sprintf("%d", stats.LinesAnalyzed)},
		{"total_clones", fmt.Sprintf("%d", stats.TotalClones)},
		{"total_clone_pairs", fmt.Sprintf("%d", stats.TotalClonePairs)},
		{"total_clone_groups", fmt.Sprintf("%d", stats.TotalCloneGroups)},
		{"average_similarity", fmt.Sprintf("%.6f", stats.AverageSimilarity)},
	}

	for _, record := range records {
		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	// Write clone type breakdown
	for cloneType, count := range stats.ClonesByType {
		record := []string{fmt.Sprintf("clone_type_%s", strings.ToLower(cloneType)), fmt.Sprintf("%d", count)}
		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}

// formatAsHTML formats the response as Lighthouse-style HTML
func (f *CloneOutputFormatter) formatAsHTML(response *domain.CloneResponse, writer io.Writer) error {
	htmlFormatter := NewHTMLFormatter()
	projectName := "Python Project" // Default project name, could be configurable

	htmlContent, err := htmlFormatter.FormatCloneAsHTML(response, projectName)
	if err != nil {
		return fmt.Errorf("failed to format as HTML: %w", err)
	}

	_, err = writer.Write([]byte(htmlContent))
	return err
}

// formatStatsAsHTML formats statistics as HTML
func (f *CloneOutputFormatter) formatStatsAsHTML(stats *domain.CloneStatistics, writer io.Writer) error {
	// Create a minimal clone response with only statistics for HTML formatting
	response := &domain.CloneResponse{
		Success:    true,
		Statistics: stats,
		ClonePairs: []*domain.ClonePair{}, // Empty pairs for stats-only view
	}

	return f.formatAsHTML(response, writer)
}
