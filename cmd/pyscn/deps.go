package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ludo-technologies/pyscn/app"
	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"
	"github.com/ludo-technologies/pyscn/service"
	"github.com/spf13/cobra"
)

// DepsCommand represents the dependency analysis command
type DepsCommand struct {
	// Output format flags (only one should be true)
	json       bool
	yaml       bool
	csv        bool
	dot        bool
	html       bool
	noOpen     bool
	configFile string
}

func NewDepsCommand() *DepsCommand { return &DepsCommand{} }

func NewDepsCmd() *cobra.Command {
	c := NewDepsCommand()

	cmd := &cobra.Command{
		Use:   "deps [paths...]",
		Short: "Analyze Python module dependencies and detect cycles",
		Long: `Build module dependency graph from Python imports and detect circular dependencies.

Examples:
  pyscn deps src/
  pyscn deps --html src/
  pyscn deps --dot src/ > deps.dot
  pyscn deps --json src/ | jq .`,
		Args: cobra.MinimumNArgs(1),
		RunE: c.run,
	}

	cmd.Flags().BoolVar(&c.json, "json", false, "Generate JSON report file")
	cmd.Flags().BoolVar(&c.yaml, "yaml", false, "Generate YAML report file")
	cmd.Flags().BoolVar(&c.csv, "csv", false, "Generate CSV report file (edges)")
	cmd.Flags().BoolVar(&c.dot, "dot", false, "Generate DOT graph file")
	cmd.Flags().BoolVar(&c.html, "html", false, "Generate HTML report file")
	cmd.Flags().BoolVar(&c.noOpen, "no-open", false, "Don't auto-open HTML in browser")
	cmd.Flags().StringVarP(&c.configFile, "config", "c", "", "Configuration file path (.pyscn.toml or pyproject)")
	return cmd
}

func (c *DepsCommand) run(cmd *cobra.Command, args []string) error {
	// Expand and validate paths
	paths, err := c.expandAndValidatePaths(args)
	if err != nil {
		return err
	}

	req := domain.DependencyRequest{
		Paths:           paths,
		Recursive:       true,
		IncludePatterns: []string{"*.py", "*.pyi"},
		ExcludePatterns: []string{"**/tests/**", "test_*.py", "*_test.py", "**/__pycache__/**"},
		OutputWriter:    cmd.OutOrStdout(),
		OutputFormat:    domain.OutputFormatText,
	}

	// Load architecture rules from config if available
	arch := c.loadArchitectureConfig(paths)
	if arch != nil {
		req.Architecture = arch
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Determine output format
	formatCount := 0
	if c.json {
		formatCount++
	}
	if c.yaml {
		formatCount++
	}
	if c.csv {
		formatCount++
	}
	if c.dot {
		formatCount++
	}
	if c.html {
		formatCount++
	}
	if formatCount > 1 {
		return fmt.Errorf("only one of --json, --yaml, --csv, --dot, --html can be specified")
	}

	// Default: text to stdout via use case
	if formatCount == 0 {
		useCase, err := c.createUseCase(cmd)
		if err != nil {
			return err
		}
		return useCase.Execute(ctx, req)
	}

	// Non-text outputs via use case (write to .pyscn/reports)
	targetPath := getTargetPathFromArgs(args)
	useCase, err := c.createUseCase(cmd)
	if err != nil {
		return err
	}
	switch {
	case c.json:
		req.OutputFormat = domain.OutputFormatJSON
		req.OutputPath, err = generateOutputFilePath("deps", "json", targetPath)
		if err != nil {
			return err
		}
		return useCase.Execute(ctx, req)
	case c.yaml:
		req.OutputFormat = domain.OutputFormatYAML
		req.OutputPath, err = generateOutputFilePath("deps", "yaml", targetPath)
		if err != nil {
			return err
		}
		return useCase.Execute(ctx, req)
	case c.csv:
		req.OutputFormat = domain.OutputFormatCSV
		req.OutputPath, err = generateOutputFilePath("deps", "csv", targetPath)
		if err != nil {
			return err
		}
		return useCase.Execute(ctx, req)
	case c.html:
		req.OutputFormat = domain.OutputFormatHTML
		req.OutputPath, err = generateOutputFilePath("deps", "html", targetPath)
		if err != nil {
			return err
		}
		req.NoOpen = c.noOpen
		return useCase.Execute(ctx, req)
	case c.dot:
		// DOT special case: compute deps and write DOT directly
		depSvc := service.NewDependencyService()
		resp, err := depSvc.Analyze(ctx, req)
		if err != nil {
			return err
		}
		return c.writeDOT(cmd, targetPath, resp.DOT)
	}
	return nil
}

// writeReportToFile centralizes file path generation and writing via FileOutputWriter
// writeDOT writes DOT content to a timestamped file and prints a status message
func (c *DepsCommand) writeDOT(cmd *cobra.Command, targetPath string, dot string) error {
	outputPath, err := generateOutputFilePath("deps", "dot", targetPath)
	if err != nil {
		return err
	}
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(dot); err != nil {
		return err
	}
	abs := outputPath
	if ap, err := filepath.Abs(outputPath); err == nil {
		abs = ap
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "DOT report generated: %s\n", abs)
	return nil
}

// note: text output is handled by DepsFormatter via DepsUseCase

func (c *DepsCommand) expandAndValidatePaths(args []string) ([]string, error) {
	var paths []string
	for _, arg := range args {
		expanded, err := filepath.Abs(arg)
		if err != nil {
			return nil, fmt.Errorf("invalid path %s: %w", arg, err)
		}
		if _, err := os.Stat(expanded); err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("path does not exist: %s", arg)
			}
			return nil, fmt.Errorf("cannot access path %s: %w", arg, err)
		}
		paths = append(paths, expanded)
	}
	return paths, nil
}

func (c *DepsCommand) loadArchitectureConfig(paths []string) *domain.ArchitectureConfigSpec {
	// Use internal/config to load project config
	// Pick first path as target to aid discovery
	var target string
	if len(paths) > 0 {
		target = paths[0]
	}
	cfg, err := config.LoadConfigWithTarget(c.configFile, target)
	if err != nil || cfg == nil {
		return nil
	}
	if cfg.Architecture == nil || !cfg.Architecture.Enabled {
		return nil
	}
	// Map to domain spec
	spec := &domain.ArchitectureConfigSpec{}
	for _, l := range cfg.Architecture.Layers {
		spec.Layers = append(spec.Layers, domain.ArchitectureLayer{Name: l.Name, Packages: l.Packages})
	}
	for _, r := range cfg.Architecture.Rules {
		spec.Rules = append(spec.Rules, domain.ArchitectureRule{From: r.From, Allow: append([]string{}, r.Allow...)})
	}
	return spec
}

func (c *DepsCommand) createUseCase(cmd *cobra.Command) (*app.DepsUseCase, error) {
	fileReader := service.NewFileReader()
	formatter := service.NewDepsFormatter()
	depSvc := service.NewDependencyService()
	return app.NewDepsUseCaseBuilder().
		WithService(depSvc).
		WithFileReader(fileReader).
		WithFormatter(formatter).
		WithOutputWriter(service.NewFileOutputWriter(cmd.ErrOrStderr())).
		Build()
}
