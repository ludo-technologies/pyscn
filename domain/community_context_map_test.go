package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCommunityContextMap_NilOrEmpty(t *testing.T) {
	assert.Nil(t, BuildCommunityContextMap(nil))
	assert.Nil(t, BuildCommunityContextMap(&CommunityAnalysisResult{}))
}

func TestBuildCommunityContextMap_BridgeFixture(t *testing.T) {
	result := &CommunityAnalysisResult{
		TotalCommunities: 2,
		Communities: []CommunityMetrics{
			{
				ID:                          "community_1",
				Modules:                     []string{"mod.b", "bridge", "mod.a"}, // unsorted on purpose
				Packages:                    []string{"mod"},
				Size:                        3,
				OutgoingCrossCommunityEdges: 1,
				RiskLevel:                   "low",
			},
			{
				ID:                          "community_2",
				Modules:                     []string{"mod.d", "mod.c"},
				Packages:                    []string{"mod"},
				Size:                        2,
				IncomingCrossCommunityEdges: 1,
				RiskLevel:                   "medium",
			},
		},
		BridgeModules: []BridgeModule{
			{Module: "bridge", Community: "community_1", CrossCommunityEdges: 1, TargetCommunities: []string{"community_2"}},
			{Module: "mod.c", Community: "community_2", CrossCommunityEdges: 1, TargetCommunities: []string{"community_1"}},
		},
	}

	cm := BuildCommunityContextMap(result)
	require.NotNil(t, cm)
	assert.Equal(t, CommunityContextMapVersion, cm.Version)
	require.Len(t, cm.Bundles, 2)

	// Bundle ordering is deterministic by community ID, modules are sorted.
	b1 := cm.Bundles[0]
	assert.Equal(t, "community_1", b1.CommunityID)
	assert.Equal(t, []string{"bridge", "mod.a", "mod.b"}, b1.Modules)
	assert.Equal(t, 3, b1.ModuleCount)
	assert.Equal(t, []string{"bridge"}, b1.BridgeModules)
	assert.Equal(t, "low", b1.RiskLevel)
	// "bridge" and "mod.*" share no common package prefix.
	assert.Empty(t, b1.SuggestedReviewScope)
	assert.Equal(t, "3 modules; 1 package; risk low; 1 cross-community edge; 1 bridge module.", b1.Summary)

	b2 := cm.Bundles[1]
	assert.Equal(t, "community_2", b2.CommunityID)
	assert.Equal(t, "mod/", b2.SuggestedReviewScope)

	// Top-level bridge modules are sorted and connect both communities.
	require.Len(t, cm.BridgeModules, 2)
	assert.Equal(t, "bridge", cm.BridgeModules[0].Module)
	assert.Equal(t, []string{"community_1", "community_2"}, cm.BridgeModules[0].Connects)
	assert.Equal(t, "1 cross-community import edge", cm.BridgeModules[0].Reason)
}

func TestBuildCommunityContextMap_ModuleCapAndScope(t *testing.T) {
	modules := make([]string, 0, CommunityContextMapModuleLimit+5)
	for i := range CommunityContextMapModuleLimit + 5 {
		// Zero-padded so lexical sort matches numeric order.
		modules = append(modules, "pkg.sub.m"+string(rune('a'+i)))
	}
	result := &CommunityAnalysisResult{
		TotalCommunities: 1,
		Communities: []CommunityMetrics{
			{ID: "community_1", Modules: modules, Packages: []string{"pkg.sub"}, Size: len(modules), RiskLevel: "high"},
		},
	}

	cm := BuildCommunityContextMap(result)
	require.NotNil(t, cm)
	require.Len(t, cm.Bundles, 1)
	b := cm.Bundles[0]

	assert.Equal(t, len(modules), b.ModuleCount)
	// Listed modules are capped at the limit plus one "... +N more" marker.
	assert.Len(t, b.Modules, CommunityContextMapModuleLimit+1)
	assert.Equal(t, "... +5 more", b.Modules[CommunityContextMapModuleLimit])
	// Common prefix across all members yields the package directory.
	assert.Equal(t, "pkg/sub/", b.SuggestedReviewScope)
}

func TestSuggestedReviewScope_SingleModuleDropsLeaf(t *testing.T) {
	result := &CommunityAnalysisResult{
		TotalCommunities: 1,
		Communities: []CommunityMetrics{
			{ID: "community_1", Modules: []string{"app.orders.service"}, Size: 1, RiskLevel: "low"},
		},
	}
	cm := BuildCommunityContextMap(result)
	require.NotNil(t, cm)
	require.Len(t, cm.Bundles, 1)
	assert.Equal(t, "app/orders/", cm.Bundles[0].SuggestedReviewScope)
}
