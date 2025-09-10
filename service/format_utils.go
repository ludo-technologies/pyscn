package service

import (
    "encoding/json"
    "fmt"
    "io"
    "strings"

    "github.com/ludo-technologies/pyscn/domain"
    "gopkg.in/yaml.v3"
)

// EncodeJSON returns an indented JSON string for the given value.
func EncodeJSON(v interface{}) (string, error) {
    data, err := json.MarshalIndent(v, "", "  ")
    if err != nil {
        return "", domain.NewOutputError("failed to marshal JSON", err)
    }
    return string(data), nil
}

// WriteJSON writes indented JSON for the given value to the writer.
func WriteJSON(w io.Writer, v interface{}) error {
    enc := json.NewEncoder(w)
    enc.SetIndent("", "  ")
    if err := enc.Encode(v); err != nil {
        return domain.NewOutputError("failed to encode JSON", err)
    }
    return nil
}

// EncodeYAML returns a YAML string for the given value.
func EncodeYAML(v interface{}) (string, error) {
    data, err := yaml.Marshal(v)
    if err != nil {
        return "", domain.NewOutputError("failed to marshal YAML", err)
    }
    return string(data), nil
}

// WriteYAML writes YAML for the given value to the writer.
func WriteYAML(w io.Writer, v interface{}) error {
    enc := yaml.NewEncoder(w)
    defer enc.Close()
    enc.SetIndent(2)
    if err := enc.Encode(v); err != nil {
        return domain.NewOutputError("failed to encode YAML", err)
    }
    return nil
}

// Standard formatting constants
const (
    HeaderWidth    = 40
    LabelWidth     = 25
    SectionPadding = 2
    ItemPadding    = 4
)

// ANSI color codes for consistent color usage
const (
    ColorReset  = "\x1b[0m"
    ColorRed    = "\x1b[31m"
    ColorYellow = "\x1b[33m"
    ColorGreen  = "\x1b[32m"
    ColorCyan   = "\x1b[36m"
    ColorBold   = "\x1b[1m"
)

// RiskLevel represents the standard risk levels across all tools
type RiskLevel string

const (
    RiskHigh   RiskLevel = "High"
    RiskMedium RiskLevel = "Medium"
    RiskLow    RiskLevel = "Low"
)

// FormatUtils provides shared formatting utilities
type FormatUtils struct{}

// NewFormatUtils creates a new format utilities instance
func NewFormatUtils() *FormatUtils {
    return &FormatUtils{}
}

// FormatMainHeader creates a standardized main header
func (f *FormatUtils) FormatMainHeader(title string) string {
    var builder strings.Builder
    builder.WriteString(title + "\n")
    builder.WriteString(strings.Repeat("=", HeaderWidth) + "\n\n")
    return builder.String()
}

// FormatSectionHeader creates a standardized section header
func (f *FormatUtils) FormatSectionHeader(title string) string {
    var builder strings.Builder
    builder.WriteString(strings.ToUpper(title) + "\n")
    builder.WriteString(strings.Repeat("-", len(title)) + "\n")
    return builder.String()
}

// FormatSectionSeparator creates a section separator
func (f *FormatUtils) FormatSectionSeparator() string {
    return "\n"
}

// FormatLabel creates a consistently formatted label with right alignment
func (f *FormatUtils) FormatLabel(label string, value interface{}) string {
    padding := LabelWidth - len(label)
    if padding < 0 {
        padding = 0
    }
    return fmt.Sprintf("%s%s: %v\n", strings.Repeat(" ", padding), label, value)
}

// FormatLabelWithIndent creates a formatted label with specific indentation
func (f *FormatUtils) FormatLabelWithIndent(indent int, label string, value interface{}) string {
    return fmt.Sprintf("%s%s: %v\n", strings.Repeat(" ", indent), label, value)
}

// FormatPercentage formats a percentage value consistently
func (f *FormatUtils) FormatPercentage(value float64) string {
    return fmt.Sprintf("%.1f%%", value)
}

// FormatDuration formats duration in milliseconds consistently
func (f *FormatUtils) FormatDuration(durationMs int64) string {
    return fmt.Sprintf("%dms", durationMs)
}

// GetRiskColor returns the appropriate color for a risk level
func (f *FormatUtils) GetRiskColor(risk RiskLevel) string {
    switch risk {
    case RiskHigh:
        return ColorRed
    case RiskMedium:
        return ColorYellow
    case RiskLow:
        return ColorGreen
    default:
        return ColorReset
    }
}

// FormatRiskWithColor formats a risk level with appropriate color
func (f *FormatUtils) FormatRiskWithColor(risk RiskLevel) string {
    color := f.GetRiskColor(risk)
    return fmt.Sprintf("%s%s%s", color, string(risk), ColorReset)
}

// FormatTableHeader creates a table header with consistent formatting
func (f *FormatUtils) FormatTableHeader(columns ...string) string {
    header := strings.Join(columns, "  ")
    separator := strings.Repeat("-", len(header))
    return header + "\n" + separator + "\n"
}

// FormatSummaryStats creates a standardized summary statistics section
func (f *FormatUtils) FormatSummaryStats(stats map[string]interface{}) string {
    var builder strings.Builder
    builder.WriteString(f.FormatSectionHeader("SUMMARY"))
    
    for label, value := range stats {
        builder.WriteString(f.FormatLabelWithIndent(SectionPadding, label, value))
    }
    
    builder.WriteString(f.FormatSectionSeparator())
    return builder.String()
}

// FormatRiskDistribution creates a standardized risk distribution section
func (f *FormatUtils) FormatRiskDistribution(high, medium, low int) string {
    var builder strings.Builder
    builder.WriteString(f.FormatSectionHeader("RISK DISTRIBUTION"))
    builder.WriteString(f.FormatLabelWithIndent(SectionPadding, "High", high))
    builder.WriteString(f.FormatLabelWithIndent(SectionPadding, "Medium", medium))
    builder.WriteString(f.FormatLabelWithIndent(SectionPadding, "Low", low))
    builder.WriteString(f.FormatSectionSeparator())
    return builder.String()
}

// FormatWarningsSection creates a standardized warnings section
func (f *FormatUtils) FormatWarningsSection(warnings []string) string {
    if len(warnings) == 0 {
        return ""
    }
    
    var builder strings.Builder
    builder.WriteString(f.FormatSectionHeader("WARNINGS"))
    
    for _, warning := range warnings {
        builder.WriteString(f.FormatLabelWithIndent(SectionPadding, "⚠", warning))
    }
    
    builder.WriteString(f.FormatSectionSeparator())
    return builder.String()
}

// ConvertToStandardRisk converts various risk representations to standard format
func (f *FormatUtils) ConvertToStandardRisk(risk string) RiskLevel {
    switch strings.ToLower(risk) {
    case "high", "critical", "error":
        return RiskHigh
    case "medium", "warning", "warn":
        return RiskMedium
    case "low", "info", "information":
        return RiskLow
    default:
        return RiskLow
    }
}

// FormatFileStats creates standardized file statistics
func (f *FormatUtils) FormatFileStats(analyzed, total, withIssues int) string {
    var builder strings.Builder
    builder.WriteString(f.FormatSectionHeader("FILE STATISTICS"))
    builder.WriteString(f.FormatLabelWithIndent(SectionPadding, "Total Files", total))
    builder.WriteString(f.FormatLabelWithIndent(SectionPadding, "Analyzed", analyzed))
    builder.WriteString(f.FormatLabelWithIndent(SectionPadding, "With Issues", withIssues))
    builder.WriteString(f.FormatSectionSeparator())
    return builder.String()
}

