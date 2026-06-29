package app

import "strings"

// ApplyAnalyzeSelection applies explicit analyzer selection to the use-case
// config. It is shared by CLI and MCP entrypoints so selection semantics stay
// consistent across transports.
func ApplyAnalyzeSelection(config AnalyzeUseCaseConfig, analyses []string) AnalyzeUseCaseConfig {
	selected := normalizeAnalyzeSelection(analyses)
	config.SelectAnalysesUsed = len(selected) > 0
	if !config.SelectAnalysesUsed {
		return config
	}

	config.SkipComplexity = !selected["complexity"]
	config.SkipDeadCode = !selected["deadcode"]
	config.SkipClones = !selected["clones"]
	config.SkipCBO = !selected["cbo"]
	config.SkipLCOM = !selected["lcom"]
	config.SkipSystem = !selected["deps"]
	config.SkipCommunities = !selected["communities"]
	return config
}

func normalizeAnalyzeSelection(analyses []string) map[string]bool {
	selected := make(map[string]bool, len(analyses))
	for _, analysis := range analyses {
		switch strings.ToLower(analysis) {
		case "dead_code":
			selected["deadcode"] = true
		case "clone":
			selected["clones"] = true
		default:
			selected[strings.ToLower(analysis)] = true
		}
	}
	return selected
}
