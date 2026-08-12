package service

import (
	"fmt"
	"strings"

	"github.com/ludo-technologies/pyscn/domain"
)

func formatDirectoryComplexityText(directories []domain.DirectoryComplexityMetrics, limit int) string {
	if len(directories) == 0 {
		return ""
	}

	utils := NewFormatUtils()
	var section strings.Builder
	section.WriteString(utils.FormatSectionHeader("DIRECTORY COMPLEXITY"))
	for index, directory := range directories {
		if limit > 0 && index >= limit {
			break
		}
		fmt.Fprintf(&section, "  %s\n", directory.DirectoryPath)
		fmt.Fprintf(&section, "    Functions: %d\n", directory.FunctionCount)
		fmt.Fprintf(&section, "    Complexity: avg %.2f, max %d, high-risk %d\n",
			directory.AverageComplexity, directory.MaxComplexity, directory.HighRiskFunctionCount)
		fmt.Fprintf(&section, "    Nesting: avg %.2f, max %d\n",
			directory.AverageNestingDepth, directory.MaxNestingDepth)
	}
	if limit > 0 && len(directories) > limit {
		fmt.Fprintf(&section, "  Showing top %d of %d directories\n", limit, len(directories))
	}
	section.WriteString(utils.FormatSectionSeparator())
	return section.String()
}
