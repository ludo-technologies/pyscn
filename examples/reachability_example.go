package main

import (
	"context"
	"fmt"
	"log"

	"github.com/pyqol/pyqol/internal/analyzer"
	"github.com/pyqol/pyqol/internal/parser"
)

func main() {
	// Example Python code with unreachable code
	pythonCode := `
def process_value(x):
    if x > 0:
        return "positive"
    elif x < 0:
        return "negative"
    else:
        return "zero"
    
    # This code is unreachable
    print("This will never execute")
    cleanup()
    
def infinite_loop():
    while True:
        print("Running...")
        if False:
            break  # This break is unreachable
    print("After loop")  # May be unreachable
`

	// Parse the Python code
	p := parser.New()
	ctx := context.Background()
	result, err := p.Parse(ctx, []byte(pythonCode))
	if err != nil {
		log.Fatalf("Failed to parse: %v", err)
	}

	fmt.Println("Analyzing reachability for Python functions...")
	fmt.Println()

	// Analyze each function
	for _, node := range result.AST.Body {
		if node.Type == parser.NodeFunctionDef {
			// Build CFG for the function
			builder := analyzer.NewCFGBuilder()
			cfg, err := builder.Build(node)
			if err != nil {
				log.Printf("Failed to build CFG for function: %v", err)
				continue
			}

			// Perform reachability analysis
			ra := analyzer.NewReachabilityAnalyzer(cfg)
			report := ra.AnalyzeReachability()

			// Display results
			fmt.Printf("Function: %s\n", node.Name)
			fmt.Printf("  Total blocks: %d\n", report.TotalBlocks)
			fmt.Printf("  Reachable blocks: %d\n", report.ReachableBlocks)
			fmt.Printf("  Unreachable blocks: %d\n", report.UnreachableBlocks)

			if report.HasUnreachableCode() {
				fmt.Println("  Unreachable code detected:")
				for _, block := range report.UnreachableList {
					if len(block.Statements) > 0 {
						fmt.Printf("    - Block %s with %d statements\n", 
							block.ID, len(block.Statements))
					}
				}

				// Mark unreachable blocks for visualization
				ra.MarkUnreachableCode()
			}

			fmt.Printf("  Analysis time: %v\n", report.AnalysisTime)
			fmt.Println()
		}
	}
}