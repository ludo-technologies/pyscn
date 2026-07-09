package service

import (
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
)

func TestConfigurationLoaderWithFlags_MergeConfig_CognitiveAndNestingThresholds(t *testing.T) {
	tests := []struct {
		name          string
		explicitFlags map[string]bool
		base          *domain.ComplexityRequest
		override      *domain.ComplexityRequest
		wantCognitive int
		wantNesting   int
	}{
		{
			name:          "explicit cognitive flag wins even when equal to default",
			explicitFlags: map[string]bool{"cognitive-complexity-threshold": true, "nesting-depth-threshold": true},
			base: &domain.ComplexityRequest{
				CognitiveComplexityThreshold: 30,
				NestingDepthThreshold:        11,
			},
			override: &domain.ComplexityRequest{
				CognitiveComplexityThreshold: domain.DefaultCognitiveComplexityThreshold, // 25
				NestingDepthThreshold:        domain.DefaultNestingDepthThreshold,        // 7
			},
			wantCognitive: domain.DefaultCognitiveComplexityThreshold,
			wantNesting:   domain.DefaultNestingDepthThreshold,
		},
		{
			name:          "no explicit flags preserves base",
			explicitFlags: map[string]bool{},
			base: &domain.ComplexityRequest{
				CognitiveComplexityThreshold: 30,
				NestingDepthThreshold:        11,
			},
			override: &domain.ComplexityRequest{
				CognitiveComplexityThreshold: 25,
				NestingDepthThreshold:        7,
			},
			wantCognitive: 30,
			wantNesting:   11,
		},
		{
			name:          "only cognitive flag set, nesting preserves base",
			explicitFlags: map[string]bool{"cognitive-complexity-threshold": true},
			base: &domain.ComplexityRequest{
				CognitiveComplexityThreshold: 30,
				NestingDepthThreshold:        11,
			},
			override: &domain.ComplexityRequest{
				CognitiveComplexityThreshold: 40,
				NestingDepthThreshold:        7,
			},
			wantCognitive: 40,
			wantNesting:   11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewConfigurationLoaderWithFlags(tt.explicitFlags)
			merged := loader.MergeConfig(tt.base, tt.override)
			assert.Equal(t, tt.wantCognitive, merged.CognitiveComplexityThreshold)
			assert.Equal(t, tt.wantNesting, merged.NestingDepthThreshold)
		})
	}
}
