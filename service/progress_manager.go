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
	mu           sync.Mutex
	writer       io.Writer
	tasks        map[string]*TaskProgress
	totalFiles   int
	interactive  bool
	initialized  bool
}

// TaskProgress tracks the progress of a single task
type TaskProgress struct {
	Name        string
	ProgressBar *progressbar.ProgressBar
	Started     bool
	Completed   bool
	Success     bool
	Processed   int
	Total       int
}

// NewProgressManager creates a new progress manager
func NewProgressManager() domain.ProgressManager {
	return &ProgressManagerImpl{
		tasks:       make(map[string]*TaskProgress),
		writer:      os.Stderr,
		interactive: isInteractiveEnvironment(),
	}
}

// Initialize sets up progress tracking for the given number of files
func (pm *ProgressManagerImpl) Initialize(totalFiles int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.totalFiles = totalFiles
	pm.initialized = true
	pm.tasks = make(map[string]*TaskProgress)
}

// StartTask marks a task as started
func (pm *ProgressManagerImpl) StartTask(taskName string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.initialized {
		return
	}

	task, exists := pm.tasks[taskName]
	if !exists {
		task = &TaskProgress{
			Name:  taskName,
			Total: pm.totalFiles,
		}
		pm.tasks[taskName] = task
	}

	task.Started = true

	// Create progress bar if interactive
	if pm.interactive && task.ProgressBar == nil {
		task.ProgressBar = pm.createProgressBar(taskName, pm.totalFiles)
	}
}

// CompleteTask marks a task as completed
func (pm *ProgressManagerImpl) CompleteTask(taskName string, success bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	task, exists := pm.tasks[taskName]
	if !exists {
		return
	}

	task.Completed = true
	task.Success = success

	// Finish progress bar if it exists
	if task.ProgressBar != nil {
		_ = task.ProgressBar.Finish()
	}
}

// UpdateProgress updates the progress for a specific task
func (pm *ProgressManagerImpl) UpdateProgress(taskName string, processed, total int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	task, exists := pm.tasks[taskName]
	if !exists {
		task = &TaskProgress{
			Name:  taskName,
			Total: total,
		}
		pm.tasks[taskName] = task
	}

	task.Processed = processed
	task.Total = total

	// Update progress bar if it exists
	if task.ProgressBar != nil {
		_ = task.ProgressBar.Set(processed)
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

	// Finish any incomplete progress bars
	for _, task := range pm.tasks {
		if task.ProgressBar != nil && !task.Completed {
			_ = task.ProgressBar.Finish()
		}
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

// GetTaskStatus returns the status of all tasks for reporting
func (pm *ProgressManagerImpl) GetTaskStatus() map[string]*TaskProgress {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Return a copy to avoid race conditions
	status := make(map[string]*TaskProgress)
	for name, task := range pm.tasks {
		taskCopy := *task
		status[name] = &taskCopy
	}
	return status
}