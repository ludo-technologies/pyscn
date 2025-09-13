package parser

import (
	"context"
	"testing"
)

func TestImportExtraction(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		expectedImport []string
		expectedFrom   []struct {
			module string
			names  []string
			level  int
		}
	}{
		{
			name: "simple imports",
			source: `import os
import sys
import json`,
			expectedImport: []string{"os", "sys", "json"},
			expectedFrom:   []struct{ module string; names []string; level int }{},
		},
		{
			name: "aliased imports",
			source: `import numpy as np
import pandas as pd`,
			expectedImport: []string{"numpy", "pandas"},
			expectedFrom:   []struct{ module string; names []string; level int }{},
		},
		{
			name: "from imports",
			source: `from typing import List, Dict, Optional
from collections import defaultdict
from os.path import join, exists`,
			expectedImport: []string{},
			expectedFrom: []struct{ module string; names []string; level int }{
				{module: "typing", names: []string{"List", "Dict", "Optional"}, level: 0},
				{module: "collections", names: []string{"defaultdict"}, level: 0},
				{module: "os.path", names: []string{"join", "exists"}, level: 0},
			},
		},
		{
			name: "relative imports",
			source: `from . import utils
from .. import parent_module
from ...package import something`,
			expectedImport: []string{},
			expectedFrom: []struct{ module string; names []string; level int }{
				{module: "", names: []string{"utils"}, level: 1},
				{module: "", names: []string{"parent_module"}, level: 2},
				{module: "package", names: []string{"something"}, level: 3},
			},
		},
		{
			name: "wildcard import",
			source: `from math import *`,
			expectedImport: []string{},
			expectedFrom: []struct{ module string; names []string; level int }{
				{module: "math", names: []string{"*"}, level: 0},
			},
		},
		{
			name: "mixed imports",
			source: `import os
import sys as system
from typing import List, Optional
from collections import defaultdict
from . import local_module
from ..parent import something`,
			expectedImport: []string{"os", "sys"},
			expectedFrom: []struct{ module string; names []string; level int }{
				{module: "typing", names: []string{"List", "Optional"}, level: 0},
				{module: "collections", names: []string{"defaultdict"}, level: 0},
				{module: "", names: []string{"local_module"}, level: 1},
				{module: "parent", names: []string{"something"}, level: 2},
			},
		},
	}

	parser := New()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(ctx, []byte(tt.source))
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			// Check import statements
			imports := result.AST.FindByType(NodeImport)
			if len(imports) != len(tt.expectedImport) {
				t.Errorf("Expected %d import statements, got %d", 
					len(tt.expectedImport), len(imports))
			}

			for i, imp := range imports {
				if i < len(tt.expectedImport) {
					if len(imp.Names) == 0 {
						t.Errorf("Import %d has no names", i)
					} else if imp.Names[0] != tt.expectedImport[i] {
						t.Errorf("Import %d: expected %s, got %s", 
							i, tt.expectedImport[i], imp.Names[0])
					}
				}
			}

			// Check from import statements
			importFroms := result.AST.FindByType(NodeImportFrom)
			if len(importFroms) != len(tt.expectedFrom) {
				t.Errorf("Expected %d from-import statements, got %d", 
					len(tt.expectedFrom), len(importFroms))
			}

			for i, imp := range importFroms {
				if i < len(tt.expectedFrom) {
					expected := tt.expectedFrom[i]
					
					if imp.Module != expected.module {
						t.Errorf("ImportFrom %d: module expected %s, got %s", 
							i, expected.module, imp.Module)
					}
					
					if imp.Level != expected.level {
						t.Errorf("ImportFrom %d: level expected %d, got %d", 
							i, expected.level, imp.Level)
					}
					
					if len(imp.Names) != len(expected.names) {
						t.Errorf("ImportFrom %d: expected %d names, got %d", 
							i, len(expected.names), len(imp.Names))
					} else {
						for j, name := range expected.names {
							if j < len(imp.Names) && imp.Names[j] != name {
								t.Errorf("ImportFrom %d, name %d: expected %s, got %s", 
									i, j, name, imp.Names[j])
							}
						}
					}
				}
			}
		})
	}
}

func TestImportWithAliases(t *testing.T) {
	source := `import numpy as np
import pandas as pd
from sklearn.model_selection import train_test_split as tts`

	parser := New()
	ctx := context.Background()

	result, err := parser.Parse(ctx, []byte(source))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Check aliased imports
	imports := result.AST.FindByType(NodeImport)
	if len(imports) != 2 {
		t.Errorf("Expected 2 import statements, got %d", len(imports))
	}

	// Check that aliases are stored as children
	for _, imp := range imports {
		if len(imp.Children) == 0 {
			t.Errorf("Import %s should have alias child", imp.Names[0])
		}
	}

	// Verify the actual import names
	expectedImports := []string{"numpy", "pandas"}
	for i, imp := range imports {
		if i < len(expectedImports) {
			if len(imp.Names) == 0 || imp.Names[0] != expectedImports[i] {
				t.Errorf("Import %d: expected %s, got %v", 
					i, expectedImports[i], imp.Names)
			}
		}
	}
}