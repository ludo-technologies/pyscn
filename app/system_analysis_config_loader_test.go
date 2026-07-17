package app

import (
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

func TestNoOpSystemAnalysisConfigLoader_PointerBoolsUseSparseMerge(t *testing.T) {
	loader := &noOpSystemAnalysisConfigLoader{}
	base := domain.DefaultSystemAnalysisRequest()
	base.Recursive = domain.BoolPtr(false)
	base.AnalyzeDependencies = domain.BoolPtr(false)

	merged := loader.MergeConfig(base, &domain.SystemAnalysisRequest{})
	if domain.BoolValue(merged.Recursive, true) {
		t.Errorf("expected nil recursive override to preserve false base, got %v", merged.Recursive)
	}
	if domain.BoolValue(merged.AnalyzeDependencies, true) {
		t.Errorf("expected nil analyze_dependencies override to preserve false base, got %v", merged.AnalyzeDependencies)
	}

	base.Recursive = domain.BoolPtr(true)
	merged = loader.MergeConfig(base, &domain.SystemAnalysisRequest{Recursive: domain.BoolPtr(false)})
	if domain.BoolValue(merged.Recursive, true) {
		t.Errorf("expected explicit recursive=false override, got %v", merged.Recursive)
	}
}
