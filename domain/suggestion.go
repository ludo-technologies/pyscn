package domain

import (
	"fmt"
	"sort"
	"strings"
)

// SuggestionCategory represents the analysis category of a suggestion
type SuggestionCategory string

const (
	SuggestionCategoryComplexity  SuggestionCategory = "complexity"
	SuggestionCategoryDeadCode    SuggestionCategory = "dead_code"
	SuggestionCategoryClone       SuggestionCategory = "clone"
	SuggestionCategoryCoupling    SuggestionCategory = "coupling"
	SuggestionCategoryCohesion    SuggestionCategory = "cohesion"
	SuggestionCategoryDependency  SuggestionCategory = "dependency"
	SuggestionCategoryArchitecure SuggestionCategory = "architecture"
)

// SuggestionSeverity represents the importance of a suggestion
type SuggestionSeverity string

const (
	SuggestionSeverityCritical SuggestionSeverity = "critical"
	SuggestionSeverityWarning  SuggestionSeverity = "warning"
	SuggestionSeverityInfo     SuggestionSeverity = "info"
)

// SuggestionEffort represents the estimated effort to address a suggestion
type SuggestionEffort string

const (
	SuggestionEffortEasy     SuggestionEffort = "easy"
	SuggestionEffortModerate SuggestionEffort = "moderate"
	SuggestionEffortHard     SuggestionEffort = "hard"
)

// Suggestion represents an actionable improvement suggestion derived from analysis results
type Suggestion struct {
	Category    SuggestionCategory `json:"category"`
	Severity    SuggestionSeverity `json:"severity"`
	Effort      SuggestionEffort   `json:"effort"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Steps       []string           `json:"steps,omitempty"`
	FilePath    string             `json:"file_path,omitempty"`
	Function    string             `json:"function,omitempty"`
	ClassName   string             `json:"class_name,omitempty"`
	StartLine   int                `json:"start_line,omitempty"`
	MetricValue string             `json:"metric_value,omitempty"`
	Threshold   string             `json:"threshold,omitempty"`
}

// maxSuggestionsPerCategory is the maximum number of suggestions generated per analysis category
const maxSuggestionsPerCategory = 10

// GenerateSuggestions derives actionable suggestions from AnalyzeResponse.
// Suggestions are sorted by priority: severity (critical > warning > info) then effort (easy first).
func GenerateSuggestions(response *AnalyzeResponse) []Suggestion {
	var suggestions []Suggestion

	suggestions = append(suggestions, generateComplexitySuggestions(response.Complexity)...)
	suggestions = append(suggestions, generateDeadCodeSuggestions(response.DeadCode)...)
	suggestions = append(suggestions, generateCloneSuggestions(response.Clone)...)
	suggestions = append(suggestions, generateCBOSuggestions(response.CBO)...)
	suggestions = append(suggestions, generateLCOMSuggestions(response.LCOM)...)
	suggestions = append(suggestions, generateSystemSuggestions(response.System)...)

	sortSuggestions(suggestions)
	return suggestions
}

// generateComplexitySuggestions generates suggestions from complexity analysis results
func generateComplexitySuggestions(resp *ComplexityResponse) []Suggestion {
	if resp == nil {
		return nil
	}

	var suggestions []Suggestion
	for _, f := range resp.Functions {
		if len(suggestions) >= maxSuggestionsPerCategory {
			break
		}

		complexity := f.Metrics.Complexity
		if complexity <= ComplexityThresholdMedium {
			continue
		}

		sev := SuggestionSeverityWarning
		if complexity > ComplexityThresholdHigh {
			sev = SuggestionSeverityCritical
		}

		desc := fmt.Sprintf("Function '%s' has cyclomatic complexity of %d.", f.Name, complexity)
		var steps []string
		if f.Metrics.NestingDepth > 4 {
			desc += " Consider using early returns or guard clauses to reduce nesting depth."
			steps = []string{
				fmt.Sprintf("Identify the deepest nested block at line %d", f.StartLine),
				"Convert nested conditions to early returns or guard clauses",
				fmt.Sprintf("Re-run: pyscn analyze %s", f.FilePath),
			}
		} else if f.Metrics.LoopStatements >= 3 {
			desc += " Consider extracting helper functions to reduce complexity."
			steps = []string{
				"Extract loop bodies into named helper functions",
				"Consider using list comprehensions or map/filter for simple loops",
				fmt.Sprintf("Re-run: pyscn analyze %s", f.FilePath),
			}
		} else if f.Metrics.ExceptionHandlers >= 3 {
			desc += " Consider extracting helper functions to reduce complexity."
			steps = []string{
				"Consolidate exception handlers — merge similar except blocks",
				"Extract try/except blocks into dedicated error-handling functions",
				fmt.Sprintf("Re-run: pyscn analyze %s", f.FilePath),
			}
		} else {
			desc += " Consider extracting helper functions to reduce complexity."
			steps = []string{
				fmt.Sprintf("Split '%s' into smaller functions, each handling one responsibility", f.Name),
				"Extract blocks of 10+ lines into well-named helper functions",
				fmt.Sprintf("Re-run: pyscn analyze %s", f.FilePath),
			}
		}

		suggestions = append(suggestions, Suggestion{
			Category:    SuggestionCategoryComplexity,
			Severity:    sev,
			Effort:      SuggestionEffortModerate,
			Title:       fmt.Sprintf("Refactor high-complexity function '%s'", f.Name),
			Description: desc,
			Steps:       steps,
			FilePath:    f.FilePath,
			Function:    f.Name,
			StartLine:   f.StartLine,
			MetricValue: fmt.Sprintf("%d", complexity),
			Threshold:   fmt.Sprintf("%d", ComplexityThresholdMedium),
		})
	}
	return suggestions
}

// generateDeadCodeSuggestions generates suggestions from dead code detection results
func generateDeadCodeSuggestions(resp *DeadCodeResponse) []Suggestion {
	if resp == nil {
		return nil
	}

	var suggestions []Suggestion
	for _, file := range resp.Files {
		for _, fn := range file.Functions {
			for _, finding := range fn.Findings {
				if len(suggestions) >= maxSuggestionsPerCategory {
					return suggestions
				}

				sev := mapDeadCodeSeverity(finding.Severity)
				effort := deadCodeEffort(finding.Reason)

				title := fmt.Sprintf("Remove dead code after %s in '%s'",
					humanizeReason(finding.Reason), finding.FunctionName)
				desc := finding.Description
				if desc == "" {
					desc = fmt.Sprintf("Dead code detected at lines %d-%d in function '%s'.",
						finding.Location.StartLine, finding.Location.EndLine, finding.FunctionName)
				}
				if effort == SuggestionEffortEasy {
					desc += " This code is safely removable."
				}

				var steps []string
				if finding.Reason == "unreachable_branch" {
					steps = []string{
						"Review the condition guarding this branch — it may always be true/false",
						"Simplify the conditional or remove the unreachable branch",
						"Run tests to confirm no regressions",
					}
				} else {
					steps = []string{
						fmt.Sprintf("Delete lines %d-%d in %s",
							finding.Location.StartLine, finding.Location.EndLine, finding.Location.FilePath),
						"Verify no side effects are lost (logging, cleanup, assignments)",
						"Run tests to confirm no regressions",
					}
				}

				suggestions = append(suggestions, Suggestion{
					Category:    SuggestionCategoryDeadCode,
					Severity:    sev,
					Effort:      effort,
					Title:       title,
					Description: desc,
					Steps:       steps,
					FilePath:    finding.Location.FilePath,
					Function:    finding.FunctionName,
					StartLine:   finding.Location.StartLine,
				})
			}
		}
	}
	return suggestions
}

// generateCloneSuggestions generates suggestions from clone detection results
func generateCloneSuggestions(resp *CloneResponse) []Suggestion {
	if resp == nil {
		return nil
	}

	var suggestions []Suggestion
	for _, group := range resp.CloneGroups {
		effort := cloneEffort(group.Type)
		memberCount := len(group.Clones)
		sev := SuggestionSeverityWarning
		if memberCount >= 4 {
			sev = SuggestionSeverityCritical
		}

		typeStr := cloneTypeLabel(group.Type)
		title := fmt.Sprintf("Extract duplicated code (%s, %d fragments)", typeStr, memberCount)
		desc := cloneDescription(group)
		steps := cloneSteps(group.Type, memberCount)

		s := Suggestion{
			Category:    SuggestionCategoryClone,
			Severity:    sev,
			Effort:      effort,
			Title:       title,
			Description: desc,
			Steps:       steps,
			MetricValue: fmt.Sprintf("%.0f%% similarity", group.Similarity*100),
		}
		if len(group.Clones) > 0 && group.Clones[0] != nil && group.Clones[0].Location != nil {
			s.FilePath = group.Clones[0].Location.FilePath
			s.StartLine = group.Clones[0].Location.StartLine
		}

		suggestions = append(suggestions, s)
	}

	// Sort by priority first, then truncate — so the most important groups survive the cap
	sortSuggestions(suggestions)
	if len(suggestions) > maxSuggestionsPerCategory {
		suggestions = suggestions[:maxSuggestionsPerCategory]
	}
	return suggestions
}

// generateCBOSuggestions generates suggestions from CBO analysis results
func generateCBOSuggestions(resp *CBOResponse) []Suggestion {
	if resp == nil {
		return nil
	}

	var suggestions []Suggestion
	for _, cls := range resp.Classes {
		if len(suggestions) >= maxSuggestionsPerCategory {
			break
		}

		cbo := cls.Metrics.CouplingCount
		if cbo <= 3 {
			continue
		}

		sev := SuggestionSeverityWarning
		effort := SuggestionEffortModerate
		if cbo > 7 {
			sev = SuggestionSeverityCritical
			effort = SuggestionEffortHard
		}

		deps := strings.Join(cls.Metrics.DependentClasses, ", ")
		desc := fmt.Sprintf("Class '%s' depends on %d other classes (%s). Consider applying dependency inversion or splitting responsibilities.",
			cls.Name, cbo, deps)

		var steps []string
		if cbo > 7 {
			steps = []string{
				fmt.Sprintf("List the %d dependencies: %s", cbo, deps),
				"Introduce interfaces for the most-used dependencies (dependency inversion)",
				fmt.Sprintf("Consider splitting '%s' if it has multiple responsibilities", cls.Name),
			}
		} else {
			steps = []string{
				fmt.Sprintf("Review the %d dependencies: %s", cbo, deps),
				"Look for dependencies that can be removed or replaced with interfaces",
			}
		}

		suggestions = append(suggestions, Suggestion{
			Category:    SuggestionCategoryCoupling,
			Severity:    sev,
			Effort:      effort,
			Title:       fmt.Sprintf("Reduce coupling in class '%s'", cls.Name),
			Description: desc,
			Steps:       steps,
			FilePath:    cls.FilePath,
			ClassName:   cls.Name,
			StartLine:   cls.StartLine,
			MetricValue: fmt.Sprintf("%d", cbo),
			Threshold:   "7",
		})
	}
	return suggestions
}

// generateLCOMSuggestions generates suggestions from LCOM analysis results
func generateLCOMSuggestions(resp *LCOMResponse) []Suggestion {
	if resp == nil {
		return nil
	}

	var suggestions []Suggestion
	for _, cls := range resp.Classes {
		if len(suggestions) >= maxSuggestionsPerCategory {
			break
		}

		lcom := cls.Metrics.LCOM4
		if lcom <= 2 {
			continue
		}

		sev := SuggestionSeverityWarning
		effort := SuggestionEffortModerate
		if lcom > 5 {
			sev = SuggestionSeverityCritical
			effort = SuggestionEffortHard
		}

		desc := fmt.Sprintf("Class '%s' has LCOM4=%d, indicating %d disconnected method groups.",
			cls.Name, lcom, lcom)
		var groupSummary string
		if len(cls.Metrics.MethodGroups) > 0 {
			groupStrs := make([]string, 0, len(cls.Metrics.MethodGroups))
			for _, g := range cls.Metrics.MethodGroups {
				groupStrs = append(groupStrs, fmt.Sprintf("[%s]", strings.Join(g, ", ")))
			}
			groupSummary = strings.Join(groupStrs, ", ")
			desc += fmt.Sprintf(" Method groups: %s. Consider splitting into separate classes.", groupSummary)
		}

		var steps []string
		if lcom > 5 {
			steps = []string{
				fmt.Sprintf("This class has %d disconnected method groups: %s", lcom, groupSummary),
				fmt.Sprintf("Split into %d classes, one per method group", lcom),
				"Use composition to connect the new classes if they need to collaborate",
			}
		} else {
			steps = []string{
				fmt.Sprintf("Review the %d method groups for natural class boundaries", lcom),
				"Consider extracting one group into a separate class",
			}
		}

		suggestions = append(suggestions, Suggestion{
			Category:    SuggestionCategoryCohesion,
			Severity:    sev,
			Effort:      effort,
			Title:       fmt.Sprintf("Split low-cohesion class '%s'", cls.Name),
			Description: desc,
			Steps:       steps,
			FilePath:    cls.FilePath,
			ClassName:   cls.Name,
			StartLine:   cls.StartLine,
			MetricValue: fmt.Sprintf("%d", lcom),
			Threshold:   "2",
		})
	}
	return suggestions
}

// generateSystemSuggestions generates suggestions from system-level analysis results
func generateSystemSuggestions(resp *SystemAnalysisResponse) []Suggestion {
	if resp == nil {
		return nil
	}

	var suggestions []Suggestion

	// Cycle breaking suggestions (own limit)
	if resp.DependencyAnalysis != nil && resp.DependencyAnalysis.CircularDependencies != nil {
		depCount := 0
		cd := resp.DependencyAnalysis.CircularDependencies
		for _, s := range cd.CycleBreakingSuggestions {
			if depCount >= maxSuggestionsPerCategory {
				break
			}
			suggestions = append(suggestions, Suggestion{
				Category:    SuggestionCategoryDependency,
				Severity:    SuggestionSeverityCritical,
				Effort:      SuggestionEffortHard,
				Title:       "Break circular dependency",
				Description: s,
				Steps:       []string{s},
			})
			depCount++
		}
	}

	// Architecture violations (own limit, independent of dependency count)
	if resp.ArchitectureAnalysis != nil {
		archCount := 0
		for _, v := range resp.ArchitectureAnalysis.Violations {
			if archCount >= maxSuggestionsPerCategory {
				break
			}
			if v.Suggestion == "" {
				continue
			}

			sev := mapViolationSeverity(v.Severity)
			effort := SuggestionEffortModerate
			if sev == SuggestionSeverityCritical {
				effort = SuggestionEffortHard
			}

			suggestions = append(suggestions, Suggestion{
				Category:    SuggestionCategoryArchitecure,
				Severity:    sev,
				Effort:      effort,
				Title:       fmt.Sprintf("Fix architecture violation in '%s'", v.Module),
				Description: v.Suggestion,
				Steps:       []string{v.Suggestion},
			})
			archCount++
		}
	}

	return suggestions
}

// sortSuggestions sorts suggestions by priority:
// Critical+Easy > Critical+Moderate > Warning+Easy > Warning+Moderate > Critical+Hard > Warning+Hard > Info+*
func sortSuggestions(suggestions []Suggestion) {
	sort.SliceStable(suggestions, func(i, j int) bool {
		return suggestionPriority(suggestions[i]) < suggestionPriority(suggestions[j])
	})
}

// suggestionPriority returns a numeric priority value (lower = higher priority)
func suggestionPriority(s Suggestion) int {
	sevWeight := 0
	switch s.Severity {
	case SuggestionSeverityCritical:
		sevWeight = 0
	case SuggestionSeverityWarning:
		sevWeight = 1
	case SuggestionSeverityInfo:
		sevWeight = 2
	}

	effortWeight := 0
	switch s.Effort {
	case SuggestionEffortEasy:
		effortWeight = 0
	case SuggestionEffortModerate:
		effortWeight = 1
	case SuggestionEffortHard:
		effortWeight = 2
	}

	// Hard effort is deprioritized: critical+hard comes after warning+moderate
	// Priority buckets:
	// 0: critical+easy
	// 1: critical+moderate
	// 2: warning+easy
	// 3: warning+moderate
	// 4: critical+hard
	// 5: warning+hard
	// 6+: info+*
	if sevWeight <= 1 && effortWeight == 2 {
		// Hard tasks are pushed after moderate tasks of all severities
		return 4 + sevWeight
	}
	if sevWeight == 2 {
		return 6 + effortWeight
	}
	return sevWeight*2 + effortWeight
}

// Helper functions

func mapDeadCodeSeverity(sev DeadCodeSeverity) SuggestionSeverity {
	switch sev {
	case DeadCodeSeverityCritical:
		return SuggestionSeverityCritical
	case DeadCodeSeverityWarning:
		return SuggestionSeverityWarning
	default:
		return SuggestionSeverityInfo
	}
}

func deadCodeEffort(reason string) SuggestionEffort {
	switch reason {
	case "after_return", "after_break", "after_continue", "after_raise":
		return SuggestionEffortEasy
	default:
		return SuggestionEffortModerate
	}
}

func humanizeReason(reason string) string {
	switch reason {
	case "after_return":
		return "return"
	case "after_break":
		return "break"
	case "after_continue":
		return "continue"
	case "after_raise":
		return "raise"
	case "unreachable_branch":
		return "unreachable branch"
	default:
		return reason
	}
}

func cloneEffort(t CloneType) SuggestionEffort {
	switch t {
	case Type1Clone:
		return SuggestionEffortEasy
	case Type2Clone:
		return SuggestionEffortEasy
	case Type3Clone:
		return SuggestionEffortModerate
	case Type4Clone:
		return SuggestionEffortHard
	default:
		return SuggestionEffortModerate
	}
}

func cloneTypeLabel(t CloneType) string {
	switch t {
	case Type1Clone:
		return "Type-1 exact"
	case Type2Clone:
		return "Type-2 renamed"
	case Type3Clone:
		return "Type-3 similar"
	case Type4Clone:
		return "Type-4 semantic"
	default:
		return "unknown type"
	}
}

func cloneDescription(group *CloneGroup) string {
	switch group.Type {
	case Type1Clone:
		return "Identical code fragments found. Extract into a shared function."
	case Type2Clone:
		return "Syntactically identical code with different identifiers. Extract into a parameterized function."
	case Type3Clone:
		return "Structurally similar code with modifications. Consider a higher-order function or template method."
	case Type4Clone:
		return "Semantically similar code with different structure. Evaluate whether a common abstraction is appropriate."
	default:
		return "Duplicated code detected. Consider refactoring to reduce duplication."
	}
}

func cloneSteps(t CloneType, memberCount int) []string {
	switch t {
	case Type1Clone:
		return []string{
			"Create a shared function from the duplicated code",
			fmt.Sprintf("Replace all %d occurrences with calls to the new function", memberCount),
			"Run tests to confirm no regressions",
		}
	case Type2Clone:
		return []string{
			"Create a parameterized function, passing differing identifiers/literals as arguments",
			fmt.Sprintf("Replace all %d occurrences with calls using appropriate arguments", memberCount),
			"Run tests to confirm no regressions",
		}
	case Type3Clone:
		return []string{
			"Identify the common structure across fragments",
			"Extract shared logic into a function, using parameters or callbacks for varying parts",
			"Run tests to confirm no regressions",
		}
	case Type4Clone:
		return []string{
			fmt.Sprintf("Analyze whether the %d fragments serve the same purpose", memberCount),
			"If so, design a common abstraction (base class, strategy pattern, or utility)",
			"Refactor incrementally — start with 2 fragments, then extend",
		}
	default:
		return []string{
			"Identify the duplicated logic",
			"Extract into a shared function or module",
			"Run tests to confirm no regressions",
		}
	}
}

func mapViolationSeverity(sev ViolationSeverity) SuggestionSeverity {
	switch sev {
	case ViolationSeverityCritical, ViolationSeverityError:
		return SuggestionSeverityCritical
	case ViolationSeverityWarning:
		return SuggestionSeverityWarning
	default:
		return SuggestionSeverityInfo
	}
}

// SeverityIcon returns an emoji icon for the suggestion severity
func (s Suggestion) SeverityIcon() string {
	switch s.Severity {
	case SuggestionSeverityCritical:
		return "\U0001F534" // red circle
	case SuggestionSeverityWarning:
		return "\U0001F7E1" // yellow circle
	default:
		return "\U0001F535" // blue circle
	}
}
