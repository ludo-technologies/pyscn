package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/pyqol/pyqol/domain"
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
		return f.formatAsJSON(response, writer)
	case domain.OutputFormatYAML:
		return f.formatAsYAML(response, writer)
	case domain.OutputFormatCSV:
		return f.formatAsCSV(response, writer)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// FormatCloneStatistics formats clone statistics
func (f *CloneOutputFormatter) FormatCloneStatistics(stats *domain.CloneStatistics, format domain.OutputFormat, writer io.Writer) error {
	switch format {
	case domain.OutputFormatText:
		return f.formatStatsAsText(stats, writer)
	case domain.OutputFormatJSON:
		return f.formatStatsAsJSON(stats, writer)
	case domain.OutputFormatYAML:
		return f.formatStatsAsYAML(stats, writer)
	case domain.OutputFormatCSV:
		return f.formatStatsAsCSV(stats, writer)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// formatAsText formats the response as human-readable text
func (f *CloneOutputFormatter) formatAsText(response *domain.CloneResponse, writer io.Writer) error {
	if !response.Success {
		fmt.Fprintf(writer, "Clone detection failed: %s\n", response.Error)
		return nil
	}

	// Print header
	fmt.Fprintf(writer, "Clone Detection Results\n")
	fmt.Fprintf(writer, "======================\n\n")

	// Print statistics
	if response.Statistics != nil {
		fmt.Fprintf(writer, "Summary:\n")
		fmt.Fprintf(writer, "  Files analyzed: %d\n", response.Statistics.FilesAnalyzed)
		fmt.Fprintf(writer, "  Lines analyzed: %d\n", response.Statistics.LinesAnalyzed)
		fmt.Fprintf(writer, "  Clone pairs found: %d\n", response.Statistics.TotalClonePairs)
		fmt.Fprintf(writer, "  Clone groups found: %d\n", response.Statistics.TotalCloneGroups)

		if response.Statistics.AverageSimilarity > 0 {
			fmt.Fprintf(writer, "  Average similarity: %.3f\n", response.Statistics.AverageSimilarity)
		}

		fmt.Fprintf(writer, "  Analysis duration: %dms\n\n", response.Duration)
	}

	// Print clone types breakdown
	if response.Statistics != nil && len(response.Statistics.ClonesByType) > 0 {
		fmt.Fprintf(writer, "Clone Types:\n")
		for cloneType, count := range response.Statistics.ClonesByType {
			fmt.Fprintf(writer, "  %s: %d pairs\n", cloneType, count)
		}
		fmt.Fprintf(writer, "\n")
	}

	if len(response.ClonePairs) == 0 {
		fmt.Fprintf(writer, "No clones detected.\n")
		return nil
	}

	// Print detailed clone pairs
	if response.Request != nil && response.Request.GroupClones && len(response.CloneGroups) > 0 {
		fmt.Fprintf(writer, "Clone Groups:\n")
		fmt.Fprintf(writer, "=============\n\n")

		for _, group := range response.CloneGroups {
			if group == nil {
				continue
			}
			fmt.Fprintf(writer, "Group %d (%s, %d clones, similarity: %.3f):\n",
				group.ID, group.Type.String(), group.Size, group.Similarity)

			for i, clone := range group.Clones {
				if clone == nil || clone.Location == nil {
					continue
				}
				fmt.Fprintf(writer, "  %d. %s (%d lines, %d nodes)\n",
					i+1, clone.Location.String(), clone.LineCount, clone.Size)
			}
			fmt.Fprintf(writer, "\n")
		}
	} else {
		fmt.Fprintf(writer, "Clone Pairs:\n")
		fmt.Fprintf(writer, "============\n\n")

		for i, pair := range response.ClonePairs {
			if pair == nil {
				continue
			}
			fmt.Fprintf(writer, "%d. %s (similarity: %.3f, confidence: %.3f)\n",
				i+1, pair.Type.String(), pair.Similarity, pair.Confidence)

			if pair.Clone1 != nil && pair.Clone1.Location != nil {
				fmt.Fprintf(writer, "   Clone 1: %s (%d lines, %d nodes)\n",
					pair.Clone1.Location.String(), pair.Clone1.LineCount, pair.Clone1.Size)
			}

			if pair.Clone2 != nil && pair.Clone2.Location != nil {
				fmt.Fprintf(writer, "   Clone 2: %s (%d lines, %d nodes)\n",
					pair.Clone2.Location.String(), pair.Clone2.LineCount, pair.Clone2.Size)
			}

			if response.Request != nil && response.Request.ShowContent && pair.Clone1 != nil && pair.Clone1.Content != "" {
				fmt.Fprintf(writer, "   Content preview:\n")
				lines := strings.Split(pair.Clone1.Content, "\n")
				for j, line := range lines {
					if j >= 5 { // Limit preview to 5 lines
						fmt.Fprintf(writer, "     ...\n")
						break
					}
					fmt.Fprintf(writer, "     %s\n", line)
				}
			}

			fmt.Fprintf(writer, "\n")
		}
	}

	return nil
}

// formatAsJSON formats the response as JSON
func (f *CloneOutputFormatter) formatAsJSON(response *domain.CloneResponse, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(response)
}

// formatAsYAML formats the response as YAML
func (f *CloneOutputFormatter) formatAsYAML(response *domain.CloneResponse, writer io.Writer) error {
	encoder := yaml.NewEncoder(writer)
	defer encoder.Close()
	encoder.SetIndent(2)
	return encoder.Encode(response)
}

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

// formatStatsAsJSON formats statistics as JSON
func (f *CloneOutputFormatter) formatStatsAsJSON(stats *domain.CloneStatistics, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(stats)
}

// formatStatsAsYAML formats statistics as YAML
func (f *CloneOutputFormatter) formatStatsAsYAML(stats *domain.CloneStatistics, writer io.Writer) error {
	encoder := yaml.NewEncoder(writer)
	defer encoder.Close()
	encoder.SetIndent(2)
	return encoder.Encode(stats)
}

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
