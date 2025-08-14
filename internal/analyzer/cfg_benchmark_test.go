package analyzer

import (
	"context"
	"strings"
	"testing"

	"github.com/pyqol/pyqol/internal/parser"
)

// BenchmarkCFGConstruction benchmarks CFG construction speed
func BenchmarkCFGConstruction(b *testing.B) {
	testCases := []struct {
		name string
		code string
	}{
		{
			name: "SimpleFunction",
			code: `
def simple():
    x = 1
    y = 2
    return x + y
`,
		},
		{
			name: "ControlFlow",
			code: `
def control_flow(x):
    if x > 0:
        result = "positive"
    elif x < 0:
        result = "negative"
    else:
        result = "zero"
    return result
`,
		},
		{
			name: "Loop",
			code: `
def loop_function():
    total = 0
    for i in range(100):
        if i % 2 == 0:
            total += i
        else:
            total -= i
    return total
`,
		},
		{
			name: "NestedLoop",
			code: `
def nested_loop():
    result = []
    for i in range(10):
        for j in range(10):
            if i * j > 50:
                break
            result.append(i * j)
    return result
`,
		},
		{
			name: "ExceptionHandling",
			code: `
def exception_handling():
    try:
        risky_operation()
        return "success"
    except ValueError:
        return "value_error"
    except TypeError:
        return "type_error"
    except Exception as e:
        return f"other_error: {e}"
    finally:
        cleanup()
`,
		},
		{
			name: "ComplexFunction",
			code: generateComplexFunction(50),
		},
		{
			name: "LargeLinearFunction",
			code: generateLargeLinearFunction(200),
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// Parse once outside the benchmark
			p := parser.New()
			ctx := context.Background()
			result, err := p.Parse(ctx, []byte(tc.code))
			if err != nil {
				b.Fatalf("Failed to parse: %v", err)
			}
			ast := result.AST

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				builder := NewCFGBuilder()
				cfg, err := builder.Build(ast)
				if err != nil {
					b.Fatalf("Failed to build CFG: %v", err)
				}
				if cfg == nil {
					b.Fatal("CFG is nil")
				}
			}
		})
	}
}

// BenchmarkReachabilityAnalysisSpeed benchmarks reachability analysis speed
func BenchmarkReachabilityAnalysisSpeed(b *testing.B) {
	testCases := []struct {
		name     string
		code     string
		setupCFG func() *CFG
	}{
		{
			name: "SimpleReachability",
			setupCFG: func() *CFG {
				code := `
def simple():
    x = 1
    if x > 0:
        return x
    else:
        return -x
`
				return buildCFGFromCode(b, code)
			},
		},
		{
			name: "ComplexReachability",
			setupCFG: func() *CFG {
				code := generateComplexFunction(100)
				return buildCFGFromCode(b, code)
			},
		},
		{
			name: "ManyBlocksReachability",
			setupCFG: func() *CFG {
				// Create a CFG with many blocks
				cfg := NewCFG("many_blocks")
				
				// Create 1000 blocks in a linear chain
				prev := cfg.Entry
				for i := 0; i < 1000; i++ {
					block := cfg.CreateBlock("block_" + string(rune('0'+i%10)))
					cfg.ConnectBlocks(prev, block, EdgeNormal)
					prev = block
				}
				cfg.ConnectBlocks(prev, cfg.Exit, EdgeNormal)
				
				return cfg
			},
		},
		{
			name: "WideReachability",
			setupCFG: func() *CFG {
				// Create a CFG with wide branching
				cfg := NewCFG("wide_branches")
				
				// Create many branches from entry
				for i := 0; i < 100; i++ {
					block := cfg.CreateBlock("branch_" + string(rune('0'+i%10)))
					cfg.ConnectBlocks(cfg.Entry, block, EdgeNormal)
					cfg.ConnectBlocks(block, cfg.Exit, EdgeNormal)
				}
				
				return cfg
			},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			cfg := tc.setupCFG()
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				analyzer := NewReachabilityAnalyzer(cfg)
				result := analyzer.AnalyzeReachability()
				if result == nil {
					b.Fatal("Reachability result is nil")
				}
			}
		})
	}
}

// BenchmarkCFGOperations benchmarks various CFG operations
func BenchmarkCFGOperations(b *testing.B) {
	// Setup a medium-sized CFG
	cfg := NewCFG("benchmark_cfg")
	
	// Create 100 blocks
	blocks := make([]*BasicBlock, 100)
	for i := 0; i < 100; i++ {
		blocks[i] = cfg.CreateBlock("block_" + string(rune('A'+i%26)))
	}
	
	// Connect them in a complex pattern
	for i := 0; i < 99; i++ {
		cfg.ConnectBlocks(blocks[i], blocks[i+1], EdgeNormal)
		if i%10 == 0 && i+10 < 100 {
			cfg.ConnectBlocks(blocks[i], blocks[i+10], EdgeCondTrue)
		}
	}
	
	b.Run("CFGSize", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			size := cfg.Size()
			if size == 0 {
				b.Fatal("Size is zero")
			}
		}
	})
	
	b.Run("CFGWalk", func(b *testing.B) {
		visitor := &testVisitor{
			onBlock: func(b *BasicBlock) bool { return true },
			onEdge:  func(e *Edge) bool { return true },
		}
		
		for i := 0; i < b.N; i++ {
			cfg.Walk(visitor)
		}
	})
	
	b.Run("CFGString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			str := cfg.String()
			if str == "" {
				b.Fatal("String is empty")
			}
		}
	})
}

// BenchmarkCFGMemoryUsage benchmarks memory usage patterns
func BenchmarkCFGMemoryUsage(b *testing.B) {
	b.Run("SmallCFG", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cfg := NewCFG("small")
			for j := 0; j < 10; j++ {
				cfg.CreateBlock("block")
			}
		}
	})
	
	b.Run("MediumCFG", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cfg := NewCFG("medium")
			for j := 0; j < 100; j++ {
				cfg.CreateBlock("block")
			}
		}
	})
	
	b.Run("LargeCFG", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cfg := NewCFG("large")
			for j := 0; j < 1000; j++ {
				cfg.CreateBlock("block")
			}
		}
	})
}

// BenchmarkCFGBuilderScalability tests scalability with different code sizes
func BenchmarkCFGBuilderScalability(b *testing.B) {
	sizes := []int{10, 50, 100, 200, 500}
	
	for _, size := range sizes {
		b.Run("LinearCode_"+string(rune('0'+size/100)), func(b *testing.B) {
			code := generateLargeLinearFunction(size)
			
			// Parse once
			p := parser.New()
			ctx := context.Background()
			result, err := p.Parse(ctx, []byte(code))
			if err != nil {
				b.Fatalf("Failed to parse: %v", err)
			}
			ast := result.AST
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				builder := NewCFGBuilder()
				cfg, err := builder.Build(ast)
				if err != nil {
					b.Fatalf("Failed to build CFG: %v", err)
				}
				if cfg == nil {
					b.Fatal("CFG is nil")
				}
			}
		})
		
		b.Run("ComplexCode_"+string(rune('0'+size/100)), func(b *testing.B) {
			code := generateComplexFunction(size)
			
			// Parse once
			p := parser.New()
			ctx := context.Background()
			result, err := p.Parse(ctx, []byte(code))
			if err != nil {
				b.Fatalf("Failed to parse: %v", err)
			}
			ast := result.AST
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				builder := NewCFGBuilder()
				cfg, err := builder.Build(ast)
				if err != nil {
					b.Fatalf("Failed to build CFG: %v", err)
				}
				if cfg == nil {
					b.Fatal("CFG is nil")
				}
			}
		})
	}
}

// Helper functions

func buildCFGFromCode(b *testing.B, code string) *CFG {
	p := parser.New()
	ctx := context.Background()
	result, err := p.Parse(ctx, []byte(code))
	if err != nil {
		b.Fatalf("Failed to parse: %v", err)
	}
	ast := result.AST
	
	builder := NewCFGBuilder()
	cfg, err := builder.Build(ast)
	if err != nil {
		b.Fatalf("Failed to build CFG: %v", err)
	}
	
	return cfg
}

func generateComplexFunction(statements int) string {
	var builder strings.Builder
	builder.WriteString("def complex_function(x):\n")
	
	// Add some initialization
	builder.WriteString("    result = 0\n")
	builder.WriteString("    temp = x\n")
	
	// Add nested control structures
	for i := 0; i < statements/10; i++ {
		builder.WriteString("    if temp > ")
		builder.WriteString(string(rune('0' + i%10)))
		builder.WriteString(":\n")
		builder.WriteString("        for j in range(")
		builder.WriteString(string(rune('5' + i%3)))
		builder.WriteString("):\n")
		builder.WriteString("            if j % 2 == 0:\n")
		builder.WriteString("                result += j\n")
		builder.WriteString("            else:\n")
		builder.WriteString("                result -= j\n")
		builder.WriteString("        temp -= 1\n")
		builder.WriteString("    else:\n")
		builder.WriteString("        result += temp\n")
	}
	
	// Add some linear statements
	for i := 0; i < statements%10; i++ {
		builder.WriteString("    var")
		builder.WriteString(string(rune('a' + i)))
		builder.WriteString(" = result + ")
		builder.WriteString(string(rune('0' + i)))
		builder.WriteString("\n")
	}
	
	builder.WriteString("    return result\n")
	return builder.String()
}

func generateLargeLinearFunction(statements int) string {
	var builder strings.Builder
	builder.WriteString("def large_linear_function():\n")
	
	for i := 0; i < statements; i++ {
		builder.WriteString("    var")
		builder.WriteString(string(rune('a' + i%26)))
		builder.WriteString(" = ")
		builder.WriteString(string(rune('0' + i%10)))
		builder.WriteString("\n")
	}
	
	builder.WriteString("    return var")
	builder.WriteString(string(rune('a' + (statements-1)%26)))
	builder.WriteString("\n")
	return builder.String()
}

// BenchmarkRealWorldCFG benchmarks CFG operations on realistic code patterns
func BenchmarkRealWorldCFG(b *testing.B) {
	realWorldCode := `
def process_data(data_list):
    """Process a list of data items with various transformations."""
    if not data_list:
        return []
    
    result = []
    errors = []
    
    try:
        for item in data_list:
            if item is None:
                continue
                
            try:
                # Validate item
                if not isinstance(item, dict):
                    raise ValueError("Item must be dict")
                
                # Process based on type
                if item.get('type') == 'A':
                    processed = process_type_a(item)
                elif item.get('type') == 'B':
                    processed = process_type_b(item)
                elif item.get('type') == 'C':
                    processed = process_type_c(item)
                else:
                    processed = process_default(item)
                
                # Validate result
                if processed and validate_result(processed):
                    result.append(processed)
                else:
                    errors.append(f"Invalid result for {item}")
                    
            except ValueError as e:
                errors.append(f"Value error for {item}: {e}")
                continue
            except TypeError as e:
                errors.append(f"Type error for {item}: {e}")
                continue
            except Exception as e:
                errors.append(f"Unexpected error for {item}: {e}")
                break
                
    except Exception as e:
        errors.append(f"Fatal error: {e}")
        return None
        
    finally:
        if errors:
            log_errors(errors)
    
    return result if not errors else None

def process_type_a(item):
    if 'value' not in item:
        return None
    return item['value'] * 2

def process_type_b(item):
    if 'value' not in item:
        return None
    return item['value'] + 10

def process_type_c(item):
    if 'value' not in item:
        return None
    return str(item['value'])

def process_default(item):
    return item.get('value', 0)

def validate_result(result):
    return result is not None

def log_errors(errors):
    for error in errors:
        print(f"ERROR: {error}")
`

	b.Run("RealWorldParsing", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			p := parser.New()
			ctx := context.Background()
			result, err := p.Parse(ctx, []byte(realWorldCode))
			if err != nil {
				b.Fatalf("Failed to parse: %v", err)
			}
			if result.AST == nil {
				b.Fatal("AST is nil")
			}
		}
	})
	
	b.Run("RealWorldCFGBuild", func(b *testing.B) {
		// Parse once
		p := parser.New()
		ctx := context.Background()
		result, err := p.Parse(ctx, []byte(realWorldCode))
		if err != nil {
			b.Fatalf("Failed to parse: %v", err)
		}
		ast := result.AST
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			builder := NewCFGBuilder()
			cfg, err := builder.Build(ast)
			if err != nil {
				b.Fatalf("Failed to build CFG: %v", err)
			}
			if cfg == nil {
				b.Fatal("CFG is nil")
			}
		}
	})
	
	b.Run("RealWorldReachability", func(b *testing.B) {
		// Parse and build once
		p := parser.New()
		ctx := context.Background()
		result, err := p.Parse(ctx, []byte(realWorldCode))
		if err != nil {
			b.Fatalf("Failed to parse: %v", err)
		}
		ast := result.AST
		
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)
		if err != nil {
			b.Fatalf("Failed to build CFG: %v", err)
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			analyzer := NewReachabilityAnalyzer(cfg)
			result := analyzer.AnalyzeReachability()
			if result == nil {
				b.Fatal("Reachability result is nil")
			}
		}
	})
}