package analyzer

import "testing"

func TestDependencyGraph_CloneOwnsAllMutableState(t *testing.T) {
	graph := NewDependencyGraph("/project")
	source := graph.AddModule("pkg.source", "/project/pkg/source.py")
	target := graph.AddModule("pkg.target", "/project/pkg/target.py")
	source.AbstractClassCount = 2
	source.PublicNames = []string{"Source"}
	target.PublicNames = []string{"Target"}
	graph.AddDependency("pkg.source", "pkg.target", DependencyEdgeFromImport, &ImportInfo{
		Statement:     "from pkg.target import Target",
		ImportedNames: []string{"Target"},
		Line:          3,
	})
	graph.RootModules = []string{"pkg.target"}
	graph.LeafModules = []string{"pkg.source"}
	graph.CyclicGroups = [][]string{{"pkg.source", "pkg.target"}}
	graph.ModuleMetrics["pkg.source"] = &ModuleMetrics{
		AbstractClassCount: 2,
		PublicInterface:    1,
	}
	graph.SystemMetrics.RefactoringPriority = []string{"pkg.source"}

	clone := graph.Clone()
	if clone.Nodes["pkg.source"].AbstractClassCount != 2 {
		t.Fatalf("expected abstract class metadata in clone, got %+v", clone.Nodes["pkg.source"])
	}
	if len(clone.CyclicGroups) != 1 || len(clone.ModuleMetrics) != 1 || len(clone.SystemMetrics.RefactoringPriority) != 1 {
		t.Fatalf("expected analysis metadata in clone, got %+v", clone)
	}

	clone.Nodes["pkg.source"].PublicNames[0] = "Changed"
	clone.Edges[0].ImportInfo.ImportedNames[0] = "Changed"
	clone.CyclicGroups[0][0] = "changed"
	clone.ModuleMetrics["pkg.source"].AbstractClassCount = 99
	clone.SystemMetrics.RefactoringPriority[0] = "changed"
	clone.RootModules[0] = "changed"

	if source.PublicNames[0] != "Source" {
		t.Fatal("clone mutated source module names")
	}
	if graph.Edges[0].ImportInfo.ImportedNames[0] != "Target" {
		t.Fatal("clone mutated source import evidence")
	}
	if graph.CyclicGroups[0][0] != "pkg.source" {
		t.Fatal("clone mutated source cycle metadata")
	}
	if graph.ModuleMetrics["pkg.source"].AbstractClassCount != 2 {
		t.Fatal("clone mutated source module metrics")
	}
	if graph.SystemMetrics.RefactoringPriority[0] != "pkg.source" {
		t.Fatal("clone mutated source system metrics")
	}
	if graph.RootModules[0] != "pkg.target" {
		t.Fatal("clone mutated source root modules")
	}
}
