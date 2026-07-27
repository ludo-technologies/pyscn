package service

import (
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

func TestOutputFormatResolver_PreservesLegacyDefault(t *testing.T) {
	format, extension, err := NewOutputFormatResolver().Determine(false, false, false, false)
	if err != nil {
		t.Fatalf("determine legacy default: %v", err)
	}
	if format != domain.OutputFormatText || extension != "" {
		t.Fatalf("expected legacy text output without extension, got %q/%q", format, extension)
	}
}

func TestOutputFormatResolver_DeterminesAnalyzeReports(t *testing.T) {
	resolver := NewOutputFormatResolver()

	format, extension, err := resolver.DetermineAnalyzeReport(false, false, false, false, false)
	if err != nil {
		t.Fatalf("determine analyze default: %v", err)
	}
	if format != domain.OutputFormatHTML || extension != "html" {
		t.Fatalf("expected default HTML report, got %q/%q", format, extension)
	}

	format, extension, err = resolver.DetermineAnalyzeReport(false, false, false, false, true)
	if err != nil {
		t.Fatalf("determine analyze text report: %v", err)
	}
	if format != domain.OutputFormatText || extension != "txt" {
		t.Fatalf("expected text/txt report, got %q/%q", format, extension)
	}

	if _, _, err := resolver.DetermineAnalyzeReport(false, true, false, false, true); err == nil {
		t.Fatal("expected conflicting report formats to fail")
	}
}
