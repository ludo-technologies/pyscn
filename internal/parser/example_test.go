package parser_test

import (
	"context"
	"fmt"
	"log"

	"github.com/ludo-technologies/pyscn/internal/parser"
)

func ExampleParser_Parse() {
	p := parser.New()
	ctx := context.Background()

	source := []byte(`def greet(name):
    return f"Hello, {name}!"

print(greet("World"))`)

	result, err := p.Parse(ctx, source)
	if err != nil {
		log.Fatal(err)
	}

	// Find all function definitions
	functions := p.FindNodes(result.RootNode, "function_definition")
	fmt.Printf("Found %d function(s)\n", len(functions))

	// Output: Found 1 function(s)
}

func ExampleParser_FindNodes() {
	p := parser.New()
	ctx := context.Background()

	source := []byte(`
class Calculator:
    def add(self, a, b):
        return a + b
    
    def subtract(self, a, b):
        return a - b
    
    def multiply(self, a, b):
        return a * b
`)

	result, err := p.Parse(ctx, source)
	if err != nil {
		log.Fatal(err)
	}

	// Find all class definitions
	classes := p.FindNodes(result.RootNode, "class_definition")
	fmt.Printf("Classes: %d\n", len(classes))

	// Find all function definitions
	functions := p.FindNodes(result.RootNode, "function_definition")
	fmt.Printf("Methods: %d\n", len(functions))

	// Output:
	// Classes: 1
	// Methods: 3
}
