package domain

import "io"

// ReportWriter abstracts writing reports to a destination (file or writer)
// and handling side-effects like opening HTML reports in a browser.
//
// Implementations live in the service layer.
type ReportWriter interface {
    // Write writes formatted content using the provided writeFunc.
    // - If outputPath is non-empty, implementations should create/truncate the file
    //   at that path and pass the file as the writer to writeFunc.
    // - If outputPath is empty, implementations should pass the provided writer to writeFunc.
    // Implementations may emit user-facing status messages (e.g., file paths) and
    // optionally open HTML outputs in a browser when format is OutputFormatHTML and noOpen is false.
    Write(writer io.Writer, outputPath string, format OutputFormat, noOpen bool, writeFunc func(io.Writer) error) error
}

