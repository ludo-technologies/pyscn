package main

import (
    "context"
    "encoding/csv"
    "io"
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/ludo-technologies/pyscn/domain"
    "github.com/ludo-technologies/pyscn/service"
    "github.com/spf13/cobra"
    "github.com/ludo-technologies/pyscn/internal/config"
)

// DepsCommand represents the dependency analysis command
type DepsCommand struct {
    // Output format flags (only one should be true)
    json bool
    yaml bool
    csv  bool
    dot  bool
    html bool
    noOpen bool
    configFile string
}

func NewDepsCmd() *cobra.Command {
    c := &DepsCommand{}

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
    cmd.Flags().BoolVar(&c.csv,  "csv",  false, "Generate CSV report file (edges)")
    cmd.Flags().BoolVar(&c.dot,  "dot",  false, "Generate DOT graph file")
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

    svc := service.NewDependencyService()

    req := domain.DependencyRequest{
        Paths:           paths,
        Recursive:       true,
        IncludePatterns: []string{"*.py", "*.pyi"},
        ExcludePatterns: []string{"**/tests/**", "test_*.py", "*_test.py", "**/__pycache__/**"},
        OutputWriter:    cmd.OutOrStdout(),
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
    resp, err := svc.Analyze(ctx, req)
    if err != nil {
        return err
    }

    // Determine output format
    formatCount := 0
    if c.json { formatCount++ }
    if c.yaml { formatCount++ }
    if c.csv  { formatCount++ }
    if c.dot  { formatCount++ }
    if c.html { formatCount++ }
    if formatCount > 1 {
        return fmt.Errorf("only one of --json, --yaml, --csv, --dot can be specified")
    }

    // Default: text to stdout
    if formatCount == 0 {
        return c.writeText(cmd, resp)
    }

    // Non-text outputs write to files under .pyscn/reports
    targetPath := getTargetPathFromArgs(args)
    outputWriter := service.NewFileOutputWriter(cmd.ErrOrStderr())

    if c.json {
        outputPath, err := generateOutputFilePath("deps", "json", targetPath)
        if err != nil { return err }
        return outputWriter.Write(cmd.OutOrStdout(), outputPath, domain.OutputFormatJSON, true, func(w io.Writer) error {
            return service.WriteJSON(w, resp)
        })
    }
    if c.yaml {
        outputPath, err := generateOutputFilePath("deps", "yaml", targetPath)
        if err != nil { return err }
        return outputWriter.Write(cmd.OutOrStdout(), outputPath, domain.OutputFormatYAML, true, func(w io.Writer) error {
            return service.WriteYAML(w, resp)
        })
    }
    if c.csv {
        outputPath, err := generateOutputFilePath("deps", "csv", targetPath)
        if err != nil { return err }
        return outputWriter.Write(cmd.OutOrStdout(), outputPath, domain.OutputFormatCSV, true, func(w io.Writer) error {
            cw := csv.NewWriter(w)
            // header
            if err := cw.Write([]string{"from", "to"}); err != nil { return err }
            for _, e := range resp.Edges {
                if err := cw.Write([]string{e.From, e.To}); err != nil { return err }
            }
            cw.Flush()
            return cw.Error()
        })
    }
    if c.html {
        outputPath, err := generateOutputFilePath("deps", "html", targetPath)
        if err != nil { return err }
        formatter := service.NewHTMLFormatter()
        html, err := formatter.FormatDepsAsHTML(resp, "Python Project")
        if err != nil { return err }
        return outputWriter.Write(cmd.OutOrStdout(), outputPath, domain.OutputFormatHTML, c.noOpen, func(w io.Writer) error {
            _, err := w.Write([]byte(html))
            return err
        })
    }
    if c.dot {
        outputPath, err := generateOutputFilePath("deps", "dot", targetPath)
        if err != nil { return err }
        // Write DOT manually (not part of standard OutputFormat)
        f, err := os.Create(outputPath)
        if err != nil { return err }
        defer f.Close()
        if _, err := f.WriteString(resp.DOT); err != nil { return err }
        abs := outputPath
        if ap, err := filepath.Abs(outputPath); err == nil { abs = ap }
        fmt.Fprintf(cmd.ErrOrStderr(), "DOT report generated: %s\n", abs)
        return nil
    }
    return nil
}

func (c *DepsCommand) writeText(cmd *cobra.Command, resp *domain.DependencyResponse) error {
    var b strings.Builder
    fmt.Fprintf(&b, "Dependency Analysis\n=====================\n\n")
    fmt.Fprintf(&b, "Modules: %d\nEdges:   %d\nCycles:  %d\n", resp.Summary.Modules, resp.Summary.Edges, resp.Summary.Cycles)
    if resp.Summary.LayerViolations > 0 {
        fmt.Fprintf(&b, "Layer Violations: %d\n", resp.Summary.LayerViolations)
    }
    fmt.Fprintf(&b, "\n")

    if len(resp.Cycles) > 0 {
        fmt.Fprintf(&b, "Cycles:\n")
        for i, cyc := range resp.Cycles {
            fmt.Fprintf(&b, "  %d) %s\n", i+1, strings.Join(cyc.Modules, " -> "))
        }
        fmt.Fprintf(&b, "\n")
    }

    if len(resp.Errors) > 0 {
        fmt.Fprintf(&b, "Errors:\n")
        for _, e := range resp.Errors {
            fmt.Fprintf(&b, "  - %s\n", e)
        }
        fmt.Fprintf(&b, "\n")
    }
    if len(resp.Warnings) > 0 {
        fmt.Fprintf(&b, "Warnings:\n")
        for _, w := range resp.Warnings {
            fmt.Fprintf(&b, "  - %s\n", w)
        }
        fmt.Fprintf(&b, "\n")
    }

    if len(resp.LayerViolations) > 0 {
        fmt.Fprintf(&b, "Layer Rule Violations:\n")
        for _, v := range resp.LayerViolations {
            fmt.Fprintf(&b, "  - %s (%s) -> %s (%s)\n", v.FromModule, v.FromLayer, v.ToModule, v.ToLayer)
        }
        fmt.Fprintf(&b, "\n")
    }

    _, err := cmd.OutOrStdout().Write([]byte(b.String()))
    return err
}

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
