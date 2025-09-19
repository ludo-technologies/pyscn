package service

import "testing"

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
				"project.service",
				"project.service.api",
			},
			rejects: []string{
				"microservice",
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
