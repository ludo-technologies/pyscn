package domain

import (
	"context"
	"io"
)

// DependencyRequest represents input for dependency analysis
type DependencyRequest struct {
	// Input files or directories to analyze
	Paths []string

	// Analysis options
	Recursive       bool
	IncludePatterns []string
	ExcludePatterns []string

	// Output configuration (used by use case formatting)
	OutputFormat OutputFormat
	OutputWriter io.Writer
	OutputPath   string
	NoOpen       bool

	// Optional architecture rules for layer validation
	Architecture *ArchitectureConfigSpec
}

// DependencyEdge represents a directed dependency between modules
type DependencyEdge struct {
	From string `json:"from" yaml:"from"`
	To   string `json:"to" yaml:"to"`
}

// DependencyCycle represents a cycle as a set of modules
type DependencyCycle struct {
	Modules []string `json:"modules" yaml:"modules"`
}

// DependencySummary contains aggregate stats
type DependencySummary struct {
	Modules         int `json:"modules" yaml:"modules"`
	Edges           int `json:"edges" yaml:"edges"`
	Cycles          int `json:"cycles" yaml:"cycles"`
	FilesAnalyzed   int `json:"files_analyzed" yaml:"files_analyzed"`
	LayerViolations int `json:"layer_violations" yaml:"layer_violations"`
}

// DependencyResponse is the result of dependency analysis
type DependencyResponse struct {
	// Graph
	Modules map[string][]string `json:"modules" yaml:"modules"` // module -> files
	Edges   []DependencyEdge    `json:"edges" yaml:"edges"`
	Cycles  []DependencyCycle   `json:"cycles" yaml:"cycles"`

	// Metadata
	Summary     DependencySummary `json:"summary" yaml:"summary"`
	Warnings    []string          `json:"warnings,omitempty" yaml:"warnings,omitempty"`
	Errors      []string          `json:"errors,omitempty" yaml:"errors,omitempty"`
	GeneratedAt string            `json:"generated_at" yaml:"generated_at"`
	Version     string            `json:"version" yaml:"version"`

	// Optional DOT representation (for convenience)
	DOT string `json:"dot,omitempty" yaml:"dot,omitempty"`

	// Optional layer validation
	LayerViolations []LayerViolation `json:"layer_violations_detail,omitempty" yaml:"layer_violations_detail,omitempty"`
}

// DependencyService defines the core business logic for dependency analysis
type DependencyService interface {
	Analyze(ctx context.Context, req DependencyRequest) (*DependencyResponse, error)
}

// DepsOutputFormatter defines the interface for formatting dependency analysis results
type DepsOutputFormatter interface {
	Write(response *DependencyResponse, format OutputFormat, writer io.Writer) error
}

// ArchitectureConfigSpec represents layer-based architecture rules (domain-friendly)
type ArchitectureConfigSpec struct {
	Layers []ArchitectureLayer `json:"layers" yaml:"layers" mapstructure:"layers"`
	Rules  []ArchitectureRule  `json:"rules"  yaml:"rules"  mapstructure:"rules"`
}

// ArchitectureLayer defines a logical layer and module patterns belonging to it
type ArchitectureLayer struct {
	Name     string   `json:"name" yaml:"name" mapstructure:"name"`
	Packages []string `json:"packages" yaml:"packages" mapstructure:"packages"`
}

// ArchitectureRule defines allowed target layers for a given source layer
type ArchitectureRule struct {
	From  string   `json:"from" yaml:"from" mapstructure:"from"`
	Allow []string `json:"allow" yaml:"allow" mapstructure:"allow"`
}

// LayerViolation represents a dependency that violates layer rules
type LayerViolation struct {
	FromModule string `json:"from_module" yaml:"from_module"`
	ToModule   string `json:"to_module" yaml:"to_module"`
	FromLayer  string `json:"from_layer" yaml:"from_layer"`
	ToLayer    string `json:"to_layer" yaml:"to_layer"`
}
