package integration

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ludo-technologies/pyscn/app"
	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/service"
)

// TestComplexityCleanCancellation tests context cancellation with working pattern
func TestComplexityCleanCancellation(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test file - use exact same pattern as working test
	testFile := filepath.Join(tempDir, "simple.py")
	content := `
def simple():
    return 1
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Setup services (exactly like the working test)
	fileReader := service.NewFileReader()
	outputFormatter := service.NewOutputFormatter()
	configLoader := service.NewConfigurationLoader()
	complexityService := service.NewComplexityService()

	// Create use case
	useCase := app.NewComplexityUseCase(
		complexityService,
		fileReader,
		outputFormatter,
		configLoader,
	)

	// Use exact same configuration as working test
	var outputBuffer bytes.Buffer
	request := domain.ComplexityRequest{
		Paths:           []string{tempDir},
		OutputFormat:    domain.OutputFormatText,
		OutputWriter:    &outputBuffer,
		MinComplexity:   1,
		MaxComplexity:   0,
		LowThreshold:    3,
		MediumThreshold: 7,
		SortBy:          domain.SortByName,
		ShowDetails:     true,
		Recursive:       true,
		IncludePatterns: []string{"*.py", "*.pyi"},
		ExcludePatterns: []string{},
	}

	// Create a context that will be cancelled immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Execute the use case
	err = useCase.Execute(ctx, request)

	// Should handle cancellation gracefully - context cancellation can be wrapped in analysis error
	if err != nil {
		// Check if the error chain contains context.Canceled
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Expected context.Canceled in error chain, got: %v", err)
		}
	}
}

// TestComplexityCleanFiltering tests complexity filtering with working pattern
func TestComplexityCleanFiltering(t *testing.T) {
	tempDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tempDir, "simple.py")
	content := `
def very_simple():
    return 1

def simple():
    x = 5
    return x

def moderate(x):
    if x > 0:
        return x * 2
    return 0

def complex_function(n):
    if n <= 0:
        return 0
    elif n == 1:
        return 1
    else:
        for i in range(n):
            if i % 2 == 0:
                continue
            else:
                if i > 5:
                    break
        return n
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Setup services (exactly like the working test)
	fileReader := service.NewFileReader()
	outputFormatter := service.NewOutputFormatter()
	configLoader := service.NewConfigurationLoader()
	complexityService := service.NewComplexityService()

	// Create use case
	useCase := app.NewComplexityUseCase(
		complexityService,
		fileReader,
		outputFormatter,
		configLoader,
	)

	// Use exact same configuration as working test
	var outputBuffer bytes.Buffer
	request := domain.ComplexityRequest{
		Paths:           []string{tempDir},
		OutputFormat:    domain.OutputFormatText,
		OutputWriter:    &outputBuffer,
		MinComplexity:   3, // Should filter out very simple functions
		MaxComplexity:   0,
		LowThreshold:    3,
		MediumThreshold: 7,
		SortBy:          domain.SortByName,
		ShowDetails:     true,
		Recursive:       true,
		IncludePatterns: []string{"*.py", "*.pyi"},
		ExcludePatterns: []string{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Execute the use case
	err = useCase.Execute(ctx, request)
	if err != nil {
		t.Fatalf("Use case execution failed: %v", err)
	}

	// Verify output was generated
	if outputBuffer.Len() == 0 {
		t.Error("Expected some output, got empty buffer")
	}
}

// TestComplexityCleanOutputFormats tests different output formats with working pattern
func TestComplexityCleanOutputFormats(t *testing.T) {
	tempDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tempDir, "simple.py")
	content := `
def test_function(x):
    if x > 0:
        return x * 2
    return 0
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test different output formats
	formats := []domain.OutputFormat{
		domain.OutputFormatJSON,
		domain.OutputFormatYAML,
		domain.OutputFormatCSV,
		domain.OutputFormatText,
	}

	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			// Setup services (exactly like the working test)
			fileReader := service.NewFileReader()
			outputFormatter := service.NewOutputFormatter()
			configLoader := service.NewConfigurationLoader()
			complexityService := service.NewComplexityService()

			// Create use case
			useCase := app.NewComplexityUseCase(
				complexityService,
				fileReader,
				outputFormatter,
				configLoader,
			)

			// Use exact same configuration as working test, but vary format
			var outputBuffer bytes.Buffer
			request := domain.ComplexityRequest{
				Paths:           []string{tempDir},
				OutputFormat:    format, // This is the only variation
				OutputWriter:    &outputBuffer,
				MinComplexity:   1,
				MaxComplexity:   0,
				LowThreshold:    3,
				MediumThreshold: 7,
				SortBy:          domain.SortByName,
				ShowDetails:     true,
				Recursive:       true,
				IncludePatterns: []string{"*.py", "*.pyi"},
				ExcludePatterns: []string{},
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Execute the use case
			err := useCase.Execute(ctx, request)
			if err != nil {
				t.Fatalf("Use case execution failed for format %s: %v", format, err)
			}

			// Verify output was generated
			if outputBuffer.Len() == 0 {
				t.Errorf("Expected output for format %s, got empty buffer", format)
			}
		})
	}
}
