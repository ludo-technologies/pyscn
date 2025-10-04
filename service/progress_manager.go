package service

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/term"
)

// ProgressManagerImpl implements the ProgressManager interface
type ProgressManagerImpl struct {
	mu          sync.Mutex
	writer      io.Writer
	progressBar *progressbar.ProgressBar
	interactive bool
	maxValue    int // Maximum value for progress (set by Initialize)
}

// NewProgressManager creates a new progress manager
func NewProgressManager() domain.ProgressManager {
	return &ProgressManagerImpl{
		writer:      os.Stderr,
		interactive: isInteractiveEnvironment(),
	}
}

// Initialize sets up progress tracking with the maximum value
func (pm *ProgressManagerImpl) Initialize(maxValue int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.maxValue = maxValue
}

// StartTask marks a task as started (creates the progress bar)
func (pm *ProgressManagerImpl) StartTask(taskName string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Create progress bar if interactive and not already created
	if pm.interactive && pm.progressBar == nil {
		pm.progressBar = pm.createProgressBar(taskName, pm.maxValue)
	}
}

// CompleteTask marks a task as completed (finishes the progress bar)
func (pm *ProgressManagerImpl) CompleteTask(taskName string, success bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Finish progress bar if it exists
	if pm.progressBar != nil {
		_ = pm.progressBar.Finish()
	}
}

// UpdateProgress updates the progress
func (pm *ProgressManagerImpl) UpdateProgress(taskName string, processed, total int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Create progress bar on first update if not created by StartTask
	if pm.progressBar == nil && pm.interactive {
		pm.progressBar = pm.createProgressBar(taskName, total)
	}

	// Update progress bar if it exists
	if pm.progressBar != nil {
		_ = pm.progressBar.Set(processed)
	}
}

// SetWriter sets the output writer for progress bars
func (pm *ProgressManagerImpl) SetWriter(writer io.Writer) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.writer = writer

	// Update interactivity check based on new writer
	if file, ok := writer.(*os.File); ok {
		pm.interactive = term.IsTerminal(int(file.Fd()))
	} else {
		pm.interactive = false
	}
}

// IsInteractive returns true if progress bars should be shown
func (pm *ProgressManagerImpl) IsInteractive() bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	return pm.interactive
}

// Close cleans up any resources
func (pm *ProgressManagerImpl) Close() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Finish progress bar if it exists
	if pm.progressBar != nil {
		_ = pm.progressBar.Finish()
	}
}

// createProgressBar creates a new progress bar with consistent styling
func (pm *ProgressManagerImpl) createProgressBar(description string, max int) *progressbar.ProgressBar {
	writer := pm.writer
	if writer == nil {
		writer = io.Discard
	}

	return progressbar.NewOptions(max,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetWriter(writer),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprintln(writer)
		}),
	)
}

// isInteractiveEnvironment returns true if the environment appears to be
// an interactive TTY session (and not CI)
func isInteractiveEnvironment() bool {
	if os.Getenv("CI") != "" {
		return false
	}
	// Best-effort TTY detection
	if fi, err := os.Stderr.Stat(); err == nil {
		return (fi.Mode() & os.ModeCharDevice) != 0
	}
	return false
}
