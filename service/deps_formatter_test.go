package service

import (
    "bytes"
    "strings"
    "testing"

    "github.com/ludo-technologies/pyscn/domain"
)

func sampleDepsResponse() *domain.DependencyResponse {
    return &domain.DependencyResponse{
        Modules: map[string][]string{"pkg.a": {"/path/a.py"}, "pkg.b": {"/path/b.py"}},
        Edges: []domain.DependencyEdge{{From: "pkg.a", To: "pkg.b"}},
        Cycles: []domain.DependencyCycle{},
        Summary: domain.DependencySummary{Modules: 2, Edges: 1, Cycles: 0},
    }
}

func TestDepsFormatter_Text(t *testing.T) {
    f := NewDepsFormatter()
    var buf bytes.Buffer
    if err := f.Write(sampleDepsResponse(), domain.OutputFormatText, &buf); err != nil {
        t.Fatalf("write text: %v", err)
    }
    out := buf.String()
    if !strings.Contains(out, "Dependency Analysis") || !strings.Contains(out, "Edges:") && !strings.Contains(out, "Edges") {
        t.Fatalf("unexpected text output: %s", out)
    }
}

func TestDepsFormatter_JSON(t *testing.T) {
    f := NewDepsFormatter()
    var buf bytes.Buffer
    if err := f.Write(sampleDepsResponse(), domain.OutputFormatJSON, &buf); err != nil {
        t.Fatalf("write json: %v", err)
    }
    if !strings.HasPrefix(buf.String(), "{") {
        t.Fatalf("expected json output, got: %s", buf.String())
    }
}

func TestDepsFormatter_YAML(t *testing.T) {
    f := NewDepsFormatter()
    var buf bytes.Buffer
    if err := f.Write(sampleDepsResponse(), domain.OutputFormatYAML, &buf); err != nil {
        t.Fatalf("write yaml: %v", err)
    }
    if !strings.Contains(buf.String(), "summary:") {
        t.Fatalf("expected yaml output, got: %s", buf.String())
    }
}

func TestDepsFormatter_CSV(t *testing.T) {
    f := NewDepsFormatter()
    var buf bytes.Buffer
    if err := f.Write(sampleDepsResponse(), domain.OutputFormatCSV, &buf); err != nil {
        t.Fatalf("write csv: %v", err)
    }
    s := buf.String()
    if !strings.HasPrefix(s, "from,to") {
        t.Fatalf("unexpected csv header: %s", s)
    }
}

func TestDepsFormatter_HTML(t *testing.T) {
    f := NewDepsFormatter()
    var buf bytes.Buffer
    if err := f.Write(sampleDepsResponse(), domain.OutputFormatHTML, &buf); err != nil {
        t.Fatalf("write html: %v", err)
    }
    if !strings.Contains(buf.String(), "<html") {
        t.Fatalf("expected html output, got: %s", buf.String())
    }
}

