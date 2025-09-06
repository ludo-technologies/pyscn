package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pyqol/pyqol/domain"
)

// ProgressReporterImpl implements the ProgressReporter interface
type ProgressReporterImpl struct {
	writer     io.Writer
	totalFiles int
	processed  int
	startTime  time.Time
	enabled    bool
	verbose    bool
}

// NewProgressReporter creates a new progress reporter
func NewProgressReporter(writer io.Writer, enabled, verbose bool) *ProgressReporterImpl {
	if writer == nil {
		writer = os.Stderr // Progress output typically goes to stderr
	}

	return &ProgressReporterImpl{
		writer:  writer,
		enabled: enabled,
		verbose: verbose,
	}
}

// StartProgress starts progress reporting for the given number of files
func (p *ProgressReporterImpl) StartProgress(totalFiles int) {
	if !p.enabled {
		return
	}

	p.totalFiles = totalFiles
	p.processed = 0
	p.startTime = time.Now()

	if p.verbose {
		fmt.Fprintf(p.writer, "🔍 Starting analysis of %d Python files...\n", totalFiles)
	} else if totalFiles > 1 {
		fmt.Fprintf(p.writer, "Analyzing %d files...\n", totalFiles)
	}
}

// UpdateProgress updates the progress with the current file being processed
func (p *ProgressReporterImpl) UpdateProgress(currentFile string, processed, total int) {
	if !p.enabled {
		return
	}

	p.processed = processed

	if p.verbose {
		// Show detailed progress with file names
		fmt.Fprintf(p.writer, "[%d/%d] %s\n", processed+1, total, filepath.Base(currentFile))
	} else if total > 10 {
		// Show simple progress for many files
		if (processed+1)%max(1, total/10) == 0 || processed+1 == total {
			percentage := int((float64(processed+1) / float64(total)) * 100)
			fmt.Fprintf(p.writer, "\rProgress: %d%% (%d/%d)", percentage, processed+1, total)
			if processed+1 == total {
				fmt.Fprintf(p.writer, "\n")
			}
		}
	}
}

// FinishProgress finishes progress reporting
func (p *ProgressReporterImpl) FinishProgress() {
	if !p.enabled {
		return
	}

	elapsed := time.Since(p.startTime)

	if p.verbose {
		rate := float64(p.totalFiles) / elapsed.Seconds()
		fmt.Fprintf(p.writer, "✅ Analysis completed in %v (%.1f files/sec)\n", elapsed.Truncate(time.Millisecond), rate)
	} else if p.totalFiles > 1 {
		fmt.Fprintf(p.writer, "Analysis completed in %v\n", elapsed.Truncate(time.Millisecond))
	}
}

// NoOpProgressReporter is a progress reporter that does nothing
type NoOpProgressReporter struct{}

// NewNoOpProgressReporter creates a no-op progress reporter
func NewNoOpProgressReporter() *NoOpProgressReporter {
	return &NoOpProgressReporter{}
}

func (n *NoOpProgressReporter) StartProgress(totalFiles int)                            {}
func (n *NoOpProgressReporter) UpdateProgress(currentFile string, processed, total int) {}
func (n *NoOpProgressReporter) FinishProgress()                                         {}

// SpinnerProgressReporter shows a simple spinner for single file analysis
type SpinnerProgressReporter struct {
	writer  io.Writer
	enabled bool
	spinner []string
	current int
	done    chan bool
}

// NewSpinnerProgressReporter creates a spinner-based progress reporter
func NewSpinnerProgressReporter(writer io.Writer, enabled bool) *SpinnerProgressReporter {
	if writer == nil {
		writer = os.Stderr
	}

	return &SpinnerProgressReporter{
		writer:  writer,
		enabled: enabled,
		spinner: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		done:    make(chan bool),
	}
}

func (s *SpinnerProgressReporter) StartProgress(totalFiles int) {
	if !s.enabled || totalFiles > 1 {
		return
	}

	fmt.Fprintf(s.writer, "Analyzing... ")
	go s.animate()
}

func (s *SpinnerProgressReporter) UpdateProgress(currentFile string, processed, total int) {
	// Spinner doesn't need updates
}

func (s *SpinnerProgressReporter) FinishProgress() {
	if !s.enabled {
		return
	}

	select {
	case s.done <- true:
		fmt.Fprintf(s.writer, "\r✅ Analysis completed\n")
	default:
		// Channel already closed or no spinner running
	}
}

func (s *SpinnerProgressReporter) animate() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			fmt.Fprintf(s.writer, "\r%s Analyzing...", s.spinner[s.current])
			s.current = (s.current + 1) % len(s.spinner)
		}
	}
}

// ProgressBarReporter shows a text-based progress bar
type ProgressBarReporter struct {
	writer     io.Writer
	enabled    bool
	totalFiles int
	processed  int
	barWidth   int
}

// NewProgressBarReporter creates a progress bar reporter
func NewProgressBarReporter(writer io.Writer, enabled bool, barWidth int) *ProgressBarReporter {
	if writer == nil {
		writer = os.Stderr
	}
	if barWidth <= 0 {
		barWidth = 40
	}

	return &ProgressBarReporter{
		writer:   writer,
		enabled:  enabled,
		barWidth: barWidth,
	}
}

func (pb *ProgressBarReporter) StartProgress(totalFiles int) {
	if !pb.enabled || totalFiles <= 1 {
		return
	}

	pb.totalFiles = totalFiles
	pb.processed = 0
	fmt.Fprintf(pb.writer, "Analyzing %d files:", totalFiles)
}

func (pb *ProgressBarReporter) UpdateProgress(currentFile string, processed, total int) {
	if !pb.enabled || total <= 1 {
		return
	}

	// Use the actual processed count (1-based for display)
	displayProcessed := processed + 1
	// Use the actual total, not the initially set totalFiles
	displayTotal := total

	percentage := float64(displayProcessed) / float64(displayTotal)
	filled := int(percentage * float64(pb.barWidth))
	
	// Ensure filled doesn't exceed barWidth
	if filled > pb.barWidth {
		filled = pb.barWidth
	}
	
	remaining := pb.barWidth - filled
	if remaining < 0 {
		remaining = 0
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", remaining)

	// Use ANSI escape sequences to clear the line and return to beginning
	fmt.Fprintf(pb.writer, "\r\033[K[%s] %3.0f%% (%d/%d) %s",
		bar, percentage*100, displayProcessed, displayTotal, filepath.Base(currentFile))
}

func (pb *ProgressBarReporter) FinishProgress() {
	if !pb.enabled || pb.totalFiles <= 1 {
		return
	}

	fmt.Fprintf(pb.writer, "\n✅ Analysis completed!\n")
}

// Factory function to create appropriate progress reporter based on context
func CreateProgressReporter(writer io.Writer, totalFiles int, verbose bool) domain.ProgressReporter {
	// Don't show progress for tests or when output is redirected to a file
	if writer == nil || !isTerminal(writer) {
		return NewNoOpProgressReporter()
	}

	// Disable progress reporting if totalFiles is 0 (used to disable progress in concurrent mode)
	if totalFiles == 0 {
		return NewNoOpProgressReporter()
	}

	if totalFiles == 1 {
		return NewSpinnerProgressReporter(writer, true)
	} else if verbose {
		return NewProgressReporter(writer, true, true)
	} else {
		return NewProgressBarReporter(writer, true, 40)
	}
}

// isTerminal checks if the writer is connected to a terminal
func isTerminal(w io.Writer) bool {
	// Simple heuristic: if it's stderr or stdout, assume it's a terminal
	// In a more sophisticated implementation, you could check if it's actually a TTY
	return w == os.Stderr || w == os.Stdout
}

// Helper function for max (since it's not available in older Go versions)
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
