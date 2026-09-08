package service

import (
	"reflect"
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

func TestOutputFormatResolver_RejectsMultipleLegacyFormats(t *testing.T) {
	if _, _, err := NewOutputFormatResolver().Determine(false, true, false, true); err == nil {
		t.Fatal("expected conflicting legacy formats to fail")
	}
}

func TestOutputFormatResolver_DeterminesAnalyzeReports(t *testing.T) {
	resolver := NewOutputFormatResolver()

	reports := resolver.DetermineAnalyzeReports(false, false, false, false, false)
	if !reflect.DeepEqual(reports, []ReportFormat{{domain.OutputFormatHTML, "html"}}) {
		t.Fatalf("expected default HTML report, got %v", reports)
	}

	reports = resolver.DetermineAnalyzeReports(false, false, false, false, true)
	if !reflect.DeepEqual(reports, []ReportFormat{{domain.OutputFormatText, "txt"}}) {
		t.Fatalf("expected text/txt report, got %v", reports)
	}

	// Every requested format is returned, in flag order (issue #739).
	reports = resolver.DetermineAnalyzeReports(true, true, false, false, true)
	want := []ReportFormat{
		{domain.OutputFormatHTML, "html"},
		{domain.OutputFormatJSON, "json"},
		{domain.OutputFormatText, "txt"},
	}
	if !reflect.DeepEqual(reports, want) {
		t.Fatalf("expected %v, got %v", want, reports)
	}
}
