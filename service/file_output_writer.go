package service

import (
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"

    "github.com/ludo-technologies/pyscn/domain"
)

// FileOutputWriter writes reports to files or provided writers and optionally opens HTML in a browser.
type FileOutputWriter struct {
    status io.Writer // where to print status messages (typically stderr)
}

// NewFileOutputWriter creates a new FileOutputWriter.
func NewFileOutputWriter(status io.Writer) *FileOutputWriter {
    if status == nil {
        status = os.Stderr
    }
    return &FileOutputWriter{status: status}
}

// Write implements domain.ReportWriter.
func (w *FileOutputWriter) Write(writer io.Writer, outputPath string, format domain.OutputFormat, noOpen bool, writeFunc func(io.Writer) error) error {
    var out io.Writer

    // If outputPath is provided, write to file; otherwise use writer.
    if outputPath != "" {
        file, err := os.Create(outputPath)
        if err != nil {
            return domain.NewOutputError(fmt.Sprintf("failed to create output file: %s", outputPath), err)
        }
        defer file.Close()
        out = file
    } else {
        out = writer
    }

    if err := writeFunc(out); err != nil {
        return domain.NewOutputError("failed to write output", err)
    }

    // Only emit status/open browser when writing to file
    if outputPath != "" {
        absPath, err := filepath.Abs(outputPath)
        if err != nil {
            absPath = outputPath
        }

        if format == domain.OutputFormatHTML {
            if !noOpen {
                fileURL := "file://" + absPath
                if err := OpenBrowser(fileURL); err != nil {
                    fmt.Fprintf(w.status, "Warning: Could not open browser: %v\n", err)
                } else {
                    fmt.Fprintf(w.status, "HTML report generated and opened: %s\n", absPath)
                    return nil
                }
            } else {
                fmt.Fprintf(w.status, "HTML report generated: %s\n", absPath)
            }
        } else {
            formatName := strings.ToUpper(string(format))
            fmt.Fprintf(w.status, "%s report generated: %s\n", formatName, absPath)
        }
    }

    return nil
}

