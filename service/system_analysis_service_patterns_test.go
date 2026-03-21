package service

import (
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/analyzer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompileModulePatternAnchoring(t *testing.T) {
	svc := NewSystemAnalysisService()

	tests := []struct {
		name    string
		pattern string
		matches []string
		rejects []string
	}{
		{
			name:    "basic segment",
			pattern: "service",
			matches: []string{
				"service",
				"service.handlers",
				"service_handler",
				"user_service",
				"project.service",
				"project.service.api",
			},
			rejects: []string{
				"microservice",
				"user_microservice",
				"core.microservice.adapter",
			},
		},
		{
			name:    "plural segment",
			pattern: "services",
			matches: []string{
				"services",
				"project.services",
				"project.services.api",
				"core.services.auth",
			},
			rejects: []string{
				"microservices.auth",
				"serviceslayer",
			},
		},
		{
			name:    "wildcard",
			pattern: "*service",
			matches: []string{
				"service",
				"microservice",
				"project.microservice",
				"layer.inner.microservice",
			},
			rejects: []string{
				"service_layer",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			re := svc.compileModulePattern(tc.pattern)
			if re == nil {
				t.Fatalf("expected regex for pattern %q", tc.pattern)
			}

			for _, module := range tc.matches {
				if !re.MatchString(module) {
					t.Fatalf("expected pattern %q to match %q", tc.pattern, module)
				}
			}

			for _, module := range tc.rejects {
				if re.MatchString(module) {
					t.Fatalf("expected pattern %q to reject %q", tc.pattern, module)
				}
			}
		})
	}
}

func TestFindLayerForModule_AmbiguousModule(t *testing.T) {
	svc := NewSystemAnalysisService()

	// Layers: "domain" owns "domain" pattern, "presentation" owns "routers" pattern.
	// Module "domain.routers" should classify as "domain" because "domain" matches
	// at prefix position (position 0), while "routers" matches at suffix position.
	compiled := make(map[string][]compiledPattern)
	for _, tc := range []struct {
		layer   string
		pattern string
	}{
		{"domain", "domain"},
		{"presentation", "routers"},
	} {
		cp := svc.compileModulePatterns(tc.pattern)
		require.NotNil(t, cp, "pattern %q should compile", tc.pattern)
		compiled[tc.layer] = append(compiled[tc.layer], *cp)
	}

	assert.Equal(t, "domain", svc.findLayerForModule("domain.routers", compiled),
		"domain.routers should be classified as domain (prefix match wins)")
	assert.Equal(t, "presentation", svc.findLayerForModule("routers", compiled),
		"routers alone should be classified as presentation")
	assert.Equal(t, "presentation", svc.findLayerForModule("app.routers", compiled),
		"app.routers should be classified as presentation")
	assert.Equal(t, "domain", svc.findLayerForModule("domain", compiled),
		"domain alone should be classified as domain")
}

func TestFindLayerForModule_PrefixMatchPriority(t *testing.T) {
	svc := NewSystemAnalysisService()

	// Both "service" and "app" could match "service.app.handler".
	// "service" matches at prefix → should win over "app" at suffix.
	compiled := make(map[string][]compiledPattern)
	for _, tc := range []struct {
		layer   string
		pattern string
	}{
		{"application", "service"},
		{"infrastructure", "app"},
	} {
		cp := svc.compileModulePatterns(tc.pattern)
		require.NotNil(t, cp)
		compiled[tc.layer] = append(compiled[tc.layer], *cp)
	}

	assert.Equal(t, "application", svc.findLayerForModule("service.app.handler", compiled),
		"prefix match 'service' should win over suffix match 'app'")
}

func TestFindLayerForModule_UnderscoreSeparatedModules(t *testing.T) {
	svc := NewSystemAnalysisService()

	compiled := make(map[string][]compiledPattern)
	for _, tc := range []struct {
		layer   string
		pattern string
	}{
		{"presentation", "api"},
		{"application", "service"},
		{"infrastructure", "repository"},
	} {
		cp := svc.compileModulePatterns(tc.pattern)
		require.NotNil(t, cp)
		compiled[tc.layer] = append(compiled[tc.layer], *cp)
	}

	assert.Equal(t, "application", svc.findLayerForModule("service_user", compiled),
		"service_ prefix should be treated as a layer boundary")
	assert.Equal(t, "application", svc.findLayerForModule("user_service", compiled),
		"_service suffix should be treated as a layer boundary")
	assert.Equal(t, "infrastructure", svc.findLayerForModule("user_repository", compiled),
		"_repository suffix should be treated as a layer boundary")
	assert.Equal(t, "presentation", svc.findLayerForModule("api_v1", compiled),
		"api_ prefix should be treated as a layer boundary")
	assert.Equal(t, "", svc.findLayerForModule("microservice", compiled),
		"plain substrings without separators should not match")
}

func TestCompileModulePatterns_PrefixSuffix(t *testing.T) {
	svc := NewSystemAnalysisService()

	cp := svc.compileModulePatterns("routers")
	require.NotNil(t, cp)

	// Prefix matches
	match := cp.matchModule("routers")
	assert.True(t, match.matched)
	assert.True(t, match.isPrefix, "routers alone should be a prefix match")

	match = cp.matchModule("routers.users")
	assert.True(t, match.matched)
	assert.True(t, match.isPrefix, "routers.users should be a prefix match")

	// Suffix matches
	match = cp.matchModule("app.routers")
	assert.True(t, match.matched)
	assert.False(t, match.isPrefix, "app.routers should be a suffix match")

	match = cp.matchModule("domain.routers.users")
	assert.True(t, match.matched)
	assert.False(t, match.isPrefix, "domain.routers.users should be a suffix match")

	// No match
	match = cp.matchModule("microrouters")
	assert.False(t, match.matched, "microrouters should not match")
}

func TestFindLayerForModule_SpecificityBeatsGeneric(t *testing.T) {
	svc := NewSystemAnalysisService()

	// "app.services" (specificity=1) should beat "service" (specificity=0)
	// when both match at the same position.
	compiled := make(map[string][]compiledPattern)
	for _, tc := range []struct {
		layer   string
		pattern string
	}{
		{"generic", "service"},
		{"specific", "app.services"},
	} {
		cp := svc.compileModulePatterns(tc.pattern)
		require.NotNil(t, cp)
		compiled[tc.layer] = append(compiled[tc.layer], *cp)
	}

	assert.Equal(t, "specific", svc.findLayerForModule("app.services.billing", compiled),
		"more specific pattern (app.services) should win")
}

func TestFindLayerForModule_DeterministicOnTie(t *testing.T) {
	svc := NewSystemAnalysisService()

	// Two layers with same specificity and pattern length — alphabetical layer name wins.
	compiled := make(map[string][]compiledPattern)
	for _, tc := range []struct {
		layer   string
		pattern string
	}{
		{"beta", "utils"},
		{"alpha", "utils"},
	} {
		cp := svc.compileModulePatterns(tc.pattern)
		require.NotNil(t, cp)
		compiled[tc.layer] = append(compiled[tc.layer], *cp)
	}

	// Run multiple times to verify determinism
	for i := 0; i < 10; i++ {
		result := svc.findLayerForModule("utils.helpers", compiled)
		assert.Equal(t, "alpha", result, "alphabetically first layer should win on tie (iteration %d)", i)
	}
}

func TestFindLayerForModule_SpecificityBeatsPrefixPosition(t *testing.T) {
	svc := NewSystemAnalysisService()

	// "foo" matches "foo.api.v1.controller" at prefix (specificity=0).
	// "api.v1" matches at suffix (specificity=1).
	// The more specific "api.v1" must win despite being a suffix match.
	compiled := make(map[string][]compiledPattern)
	for _, tc := range []struct {
		layer   string
		pattern string
	}{
		{"catch_all", "foo"},
		{"api_layer", "api.v1"},
	} {
		cp := svc.compileModulePatterns(tc.pattern)
		require.NotNil(t, cp)
		compiled[tc.layer] = append(compiled[tc.layer], *cp)
	}

	assert.Equal(t, "api_layer", svc.findLayerForModule("foo.api.v1.controller", compiled),
		"more specific 'api.v1' (suffix) should beat less specific 'foo' (prefix)")
	// When specificity is equal, prefix still wins
	assert.Equal(t, "catch_all", svc.findLayerForModule("foo.bar", compiled),
		"with equal specificity, prefix 'foo' should win")
}

func TestBuildModuleLayerMap_AmbiguousPackagesPrefixWins(t *testing.T) {
	svc := NewSystemAnalysisService()

	graph := analyzer.NewDependencyGraph("/project")
	graph.AddModule("domain.routers.api", "/project/domain/routers/api.py")
	graph.AddModule("routers.main", "/project/routers/main.py")
	graph.AddModule("app.services.billing", "/project/app/services/billing.py")

	rules := &domain.ArchitectureRules{
		Layers: []domain.Layer{
			{Name: "domain", Packages: []string{"domain"}},
			{Name: "presentation", Packages: []string{"routers"}},
			{Name: "application", Packages: []string{"app.services"}},
		},
	}

	moduleToLayer := svc.buildModuleLayerMap(graph, rules)

	assert.Equal(t, "domain", moduleToLayer["domain.routers.api"],
		"domain.routers.api: prefix 'domain' should win over suffix 'routers'")
	assert.Equal(t, "presentation", moduleToLayer["routers.main"],
		"routers.main: prefix 'routers' → presentation")
	assert.Equal(t, "application", moduleToLayer["app.services.billing"],
		"app.services.billing: prefix 'app.services' → application")
}
