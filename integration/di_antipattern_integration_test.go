package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ludo-technologies/pyscn/app"
	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/service"
)

// TestDIAntipatternDetection_ConstructorOverInjection tests detection of constructor over-injection
func TestDIAntipatternDetection_ConstructorOverInjection(t *testing.T) {
	tempDir := t.TempDir()

	// Create test file with constructor over-injection
	testFile := filepath.Join(tempDir, "over_injection.py")
	content := `
class BadService:
    def __init__(self, repo1, repo2, repo3, service1, service2, service3, logger, config):
        self.repo1 = repo1
        self.repo2 = repo2
        self.repo3 = repo3
        self.service1 = service1
        self.service2 = service2
        self.service3 = service3
        self.logger = logger
        self.config = config

class GoodService:
    def __init__(self, repo, logger, config):
        self.repo = repo
        self.logger = logger
        self.config = config
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Setup services
	fileReader := service.NewFileReader()
	formatter := service.NewDIAntipatternFormatter()
	diService := service.NewDIAntipatternService()

	// Create use case
	useCase, err := app.NewDIAntipatternUseCaseBuilder().
		WithService(diService).
		WithFileReader(fileReader).
		WithFormatter(formatter).
		Build()
	if err != nil {
		t.Fatalf("Failed to build use case: %v", err)
	}

	var outputBuffer bytes.Buffer
	recursive := true
	request := domain.DIAntipatternRequest{
		Paths:                     []string{tempDir},
		OutputFormat:              domain.OutputFormatJSON,
		OutputWriter:              &outputBuffer,
		Recursive:                 &recursive,
		ConstructorParamThreshold: 5,
		MinSeverity:               domain.DIAntipatternSeverityWarning,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = useCase.Execute(ctx, request)
	if err != nil {
		t.Fatalf("Use case execution failed: %v", err)
	}

	// Parse output
	var response domain.DIAntipatternResponse
	if err := json.Unmarshal(outputBuffer.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify findings
	if response.Summary.TotalFindings == 0 {
		t.Error("Expected at least one finding")
	}

	// Verify constructor over-injection was detected
	found := false
	for _, f := range response.Findings {
		if f.Type == domain.DIAntipatternConstructorOverInjection && f.ClassName == "BadService" {
			found = true
			if f.Details["parameter_count"].(float64) != 8 {
				t.Errorf("Expected 8 parameters, got %v", f.Details["parameter_count"])
			}
			break
		}
	}
	if !found {
		t.Error("Expected to find constructor over-injection in BadService")
	}
}

// TestDIAntipatternDetection_HiddenDependency tests detection of hidden dependencies
func TestDIAntipatternDetection_HiddenDependency(t *testing.T) {
	tempDir := t.TempDir()

	// Create test file with hidden dependencies
	testFile := filepath.Join(tempDir, "hidden_deps.py")
	content := `
_global_config = {}

class GlobalUser:
    def do_something(self):
        global _global_config
        return _global_config["key"]

class SingletonService:
    _instance = None

    def __new__(cls):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
        return cls._instance
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Setup services
	fileReader := service.NewFileReader()
	formatter := service.NewDIAntipatternFormatter()
	diService := service.NewDIAntipatternService()

	useCase, err := app.NewDIAntipatternUseCaseBuilder().
		WithService(diService).
		WithFileReader(fileReader).
		WithFormatter(formatter).
		Build()
	if err != nil {
		t.Fatalf("Failed to build use case: %v", err)
	}

	var outputBuffer bytes.Buffer
	recursive := true
	request := domain.DIAntipatternRequest{
		Paths:                     []string{tempDir},
		OutputFormat:              domain.OutputFormatJSON,
		OutputWriter:              &outputBuffer,
		Recursive:                 &recursive,
		ConstructorParamThreshold: 5,
		MinSeverity:               domain.DIAntipatternSeverityInfo,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = useCase.Execute(ctx, request)
	if err != nil {
		t.Fatalf("Use case execution failed: %v", err)
	}

	var response domain.DIAntipatternResponse
	if err := json.Unmarshal(outputBuffer.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify hidden dependency findings
	globalFound := false
	singletonFound := false
	for _, f := range response.Findings {
		if f.Type == domain.DIAntipatternHiddenDependency {
			if f.Subtype == string(domain.HiddenDepGlobal) {
				globalFound = true
			}
			if f.Subtype == string(domain.HiddenDepSingleton) {
				singletonFound = true
			}
		}
	}

	if !globalFound {
		t.Error("Expected to find global statement usage")
	}
	if !singletonFound {
		t.Error("Expected to find singleton pattern")
	}
}

// TestDIAntipatternDetection_ConcreteDependency tests detection of concrete dependencies
func TestDIAntipatternDetection_ConcreteDependency(t *testing.T) {
	tempDir := t.TempDir()

	// Create test file with concrete dependencies
	testFile := filepath.Join(tempDir, "concrete_deps.py")
	content := `
class MySQLRepository:
    pass

class FileLogger:
    pass

class ConcreteTypeHintService:
    def __init__(self, repo: MySQLRepository):
        self.repo = repo

class DirectInstantiationService:
    def __init__(self):
        self.logger = FileLogger()
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fileReader := service.NewFileReader()
	formatter := service.NewDIAntipatternFormatter()
	diService := service.NewDIAntipatternService()

	useCase, err := app.NewDIAntipatternUseCaseBuilder().
		WithService(diService).
		WithFileReader(fileReader).
		WithFormatter(formatter).
		Build()
	if err != nil {
		t.Fatalf("Failed to build use case: %v", err)
	}

	var outputBuffer bytes.Buffer
	recursive := true
	request := domain.DIAntipatternRequest{
		Paths:                     []string{tempDir},
		OutputFormat:              domain.OutputFormatJSON,
		OutputWriter:              &outputBuffer,
		Recursive:                 &recursive,
		ConstructorParamThreshold: 5,
		MinSeverity:               domain.DIAntipatternSeverityInfo,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = useCase.Execute(ctx, request)
	if err != nil {
		t.Fatalf("Use case execution failed: %v", err)
	}

	var response domain.DIAntipatternResponse
	if err := json.Unmarshal(outputBuffer.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	typeHintFound := false
	instantiationFound := false
	for _, f := range response.Findings {
		if f.Type == domain.DIAntipatternConcreteDependency {
			if f.Subtype == string(domain.ConcreteDepTypeHint) {
				typeHintFound = true
			}
			if f.Subtype == string(domain.ConcreteDepInstantiation) {
				instantiationFound = true
			}
		}
	}

	if !typeHintFound {
		t.Error("Expected to find concrete type hint")
	}
	if !instantiationFound {
		t.Error("Expected to find direct instantiation")
	}
}

// TestDIAntipatternDetection_ServiceLocator tests detection of service locator pattern
func TestDIAntipatternDetection_ServiceLocator(t *testing.T) {
	tempDir := t.TempDir()

	// Create test file with service locator patterns
	testFile := filepath.Join(tempDir, "service_locator.py")
	content := `
class Container:
    @classmethod
    def get(cls, name):
        return None

class ServiceLocatorUser:
    def __init__(self):
        self.logger = Container.get("logger")

    def process(self):
        return resolve("executor")

class GlobalLocatorUser:
    def __init__(self):
        self.config = get_service("config")
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fileReader := service.NewFileReader()
	formatter := service.NewDIAntipatternFormatter()
	diService := service.NewDIAntipatternService()

	useCase, err := app.NewDIAntipatternUseCaseBuilder().
		WithService(diService).
		WithFileReader(fileReader).
		WithFormatter(formatter).
		Build()
	if err != nil {
		t.Fatalf("Failed to build use case: %v", err)
	}

	var outputBuffer bytes.Buffer
	recursive := true
	request := domain.DIAntipatternRequest{
		Paths:                     []string{tempDir},
		OutputFormat:              domain.OutputFormatJSON,
		OutputWriter:              &outputBuffer,
		Recursive:                 &recursive,
		ConstructorParamThreshold: 5,
		MinSeverity:               domain.DIAntipatternSeverityWarning,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = useCase.Execute(ctx, request)
	if err != nil {
		t.Fatalf("Use case execution failed: %v", err)
	}

	var response domain.DIAntipatternResponse
	if err := json.Unmarshal(outputBuffer.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	serviceLocatorCount := 0
	for _, f := range response.Findings {
		if f.Type == domain.DIAntipatternServiceLocator {
			serviceLocatorCount++
		}
	}

	if serviceLocatorCount < 3 {
		t.Errorf("Expected at least 3 service locator findings, got %d", serviceLocatorCount)
	}
}

// TestDIAntipatternDetection_OutputFormats tests different output formats
func TestDIAntipatternDetection_OutputFormats(t *testing.T) {
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "test.py")
	content := `
class BadService:
    def __init__(self, a, b, c, d, e, f):
        pass
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	formats := []domain.OutputFormat{
		domain.OutputFormatJSON,
		domain.OutputFormatYAML,
		domain.OutputFormatCSV,
		domain.OutputFormatText,
	}

	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			fileReader := service.NewFileReader()
			formatter := service.NewDIAntipatternFormatter()
			diService := service.NewDIAntipatternService()

			useCase, err := app.NewDIAntipatternUseCaseBuilder().
				WithService(diService).
				WithFileReader(fileReader).
				WithFormatter(formatter).
				Build()
			if err != nil {
				t.Fatalf("Failed to build use case: %v", err)
			}

			var outputBuffer bytes.Buffer
			recursive := true
			request := domain.DIAntipatternRequest{
				Paths:                     []string{tempDir},
				OutputFormat:              format,
				OutputWriter:              &outputBuffer,
				Recursive:                 &recursive,
				ConstructorParamThreshold: 5,
				MinSeverity:               domain.DIAntipatternSeverityWarning,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			err = useCase.Execute(ctx, request)
			if err != nil {
				t.Fatalf("Use case execution failed for format %s: %v", format, err)
			}

			if outputBuffer.Len() == 0 {
				t.Errorf("Expected output for format %s, got empty buffer", format)
			}
		})
	}
}

// TestDIAntipatternDetection_SeverityFiltering tests severity filtering
func TestDIAntipatternDetection_SeverityFiltering(t *testing.T) {
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "mixed.py")
	content := `
_config = {}

class Service:
    def __init__(self, repo: ConcreteRepo):
        global _config
        self.repo = repo
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name           string
		minSeverity    domain.DIAntipatternSeverity
		expectFindings bool
	}{
		{"info", domain.DIAntipatternSeverityInfo, true},
		{"warning", domain.DIAntipatternSeverityWarning, true},
		{"error", domain.DIAntipatternSeverityError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileReader := service.NewFileReader()
			formatter := service.NewDIAntipatternFormatter()
			diService := service.NewDIAntipatternService()

			useCase, err := app.NewDIAntipatternUseCaseBuilder().
				WithService(diService).
				WithFileReader(fileReader).
				WithFormatter(formatter).
				Build()
			if err != nil {
				t.Fatalf("Failed to build use case: %v", err)
			}

			var outputBuffer bytes.Buffer
			recursive := true
			request := domain.DIAntipatternRequest{
				Paths:                     []string{tempDir},
				OutputFormat:              domain.OutputFormatJSON,
				OutputWriter:              &outputBuffer,
				Recursive:                 &recursive,
				ConstructorParamThreshold: 5,
				MinSeverity:               tt.minSeverity,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			err = useCase.Execute(ctx, request)
			if err != nil {
				t.Fatalf("Use case execution failed: %v", err)
			}

			var response domain.DIAntipatternResponse
			if err := json.Unmarshal(outputBuffer.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse JSON output: %v", err)
			}

			if tt.expectFindings && response.Summary.TotalFindings == 0 {
				t.Error("Expected findings but got none")
			}
		})
	}
}

// TestDIAntipatternDetection_MultipleFiles tests detection across multiple files
func TestDIAntipatternDetection_MultipleFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create multiple test files
	files := map[string]string{
		"service1.py": `
class Service1:
    def __init__(self, a, b, c, d, e, f):
        pass
`,
		"service2.py": `
class Service2:
    def __init__(self):
        self.logger = get_service("logger")
`,
		"nested/service3.py": `
class Service3:
    _instance = None
    def __new__(cls):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
        return cls._instance
`,
	}

	for path, content := range files {
		fullPath := filepath.Join(tempDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	fileReader := service.NewFileReader()
	formatter := service.NewDIAntipatternFormatter()
	diService := service.NewDIAntipatternService()

	useCase, err := app.NewDIAntipatternUseCaseBuilder().
		WithService(diService).
		WithFileReader(fileReader).
		WithFormatter(formatter).
		Build()
	if err != nil {
		t.Fatalf("Failed to build use case: %v", err)
	}

	var outputBuffer bytes.Buffer
	recursive := true
	request := domain.DIAntipatternRequest{
		Paths:                     []string{tempDir},
		OutputFormat:              domain.OutputFormatJSON,
		OutputWriter:              &outputBuffer,
		Recursive:                 &recursive,
		ConstructorParamThreshold: 5,
		MinSeverity:               domain.DIAntipatternSeverityWarning,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = useCase.Execute(ctx, request)
	if err != nil {
		t.Fatalf("Use case execution failed: %v", err)
	}

	var response domain.DIAntipatternResponse
	if err := json.Unmarshal(outputBuffer.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify findings from multiple files
	if response.Summary.FilesAnalyzed < 3 {
		t.Errorf("Expected at least 3 files analyzed, got %d", response.Summary.FilesAnalyzed)
	}

	if response.Summary.TotalFindings < 3 {
		t.Errorf("Expected at least 3 findings, got %d", response.Summary.TotalFindings)
	}
}

// TestDIAntipatternDetection_CleanCode tests that clean code produces no findings
func TestDIAntipatternDetection_CleanCode(t *testing.T) {
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "clean.py")
	content := `
from abc import ABC, abstractmethod

class IRepository(ABC):
    @abstractmethod
    def save(self, entity):
        pass

class ILogger(ABC):
    @abstractmethod
    def log(self, message):
        pass

class CleanService:
    def __init__(self, repo: IRepository, logger: ILogger):
        self.repo = repo
        self.logger = logger

    def process(self, data):
        self.logger.log("Processing...")
        self.repo.save(data)
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fileReader := service.NewFileReader()
	formatter := service.NewDIAntipatternFormatter()
	diService := service.NewDIAntipatternService()

	useCase, err := app.NewDIAntipatternUseCaseBuilder().
		WithService(diService).
		WithFileReader(fileReader).
		WithFormatter(formatter).
		Build()
	if err != nil {
		t.Fatalf("Failed to build use case: %v", err)
	}

	var outputBuffer bytes.Buffer
	recursive := true
	request := domain.DIAntipatternRequest{
		Paths:                     []string{tempDir},
		OutputFormat:              domain.OutputFormatJSON,
		OutputWriter:              &outputBuffer,
		Recursive:                 &recursive,
		ConstructorParamThreshold: 5,
		MinSeverity:               domain.DIAntipatternSeverityInfo,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = useCase.Execute(ctx, request)
	if err != nil {
		t.Fatalf("Use case execution failed: %v", err)
	}

	var response domain.DIAntipatternResponse
	if err := json.Unmarshal(outputBuffer.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Clean code should have no findings
	if response.Summary.TotalFindings != 0 {
		t.Errorf("Expected no findings for clean code, got %d", response.Summary.TotalFindings)
		for _, f := range response.Findings {
			t.Logf("Unexpected finding: %s - %s", f.Type, f.Description)
		}
	}
}

// TestDIAntipatternDetection_TextOutputFormat tests text output format content
func TestDIAntipatternDetection_TextOutputFormat(t *testing.T) {
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "test.py")
	content := `
class BadService:
    def __init__(self, a, b, c, d, e, f):
        pass
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fileReader := service.NewFileReader()
	formatter := service.NewDIAntipatternFormatter()
	diService := service.NewDIAntipatternService()

	useCase, err := app.NewDIAntipatternUseCaseBuilder().
		WithService(diService).
		WithFileReader(fileReader).
		WithFormatter(formatter).
		Build()
	if err != nil {
		t.Fatalf("Failed to build use case: %v", err)
	}

	var outputBuffer bytes.Buffer
	recursive := true
	request := domain.DIAntipatternRequest{
		Paths:                     []string{tempDir},
		OutputFormat:              domain.OutputFormatText,
		OutputWriter:              &outputBuffer,
		Recursive:                 &recursive,
		ConstructorParamThreshold: 5,
		MinSeverity:               domain.DIAntipatternSeverityWarning,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = useCase.Execute(ctx, request)
	if err != nil {
		t.Fatalf("Use case execution failed: %v", err)
	}

	output := outputBuffer.String()

	// Verify text output contains expected elements
	expectedElements := []string{
		"BadService",
		"constructor_over_injection",
		"warning",
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected text output to contain %q", element)
		}
	}
}
