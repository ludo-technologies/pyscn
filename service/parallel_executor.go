package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
)

// ParallelExecutorImpl implements the ParallelExecutor interface
type ParallelExecutorImpl struct {
	maxConcurrency int
	timeout        time.Duration
}

// NewParallelExecutor creates a new parallel executor
func NewParallelExecutor() domain.ParallelExecutor {
	return &ParallelExecutorImpl{
		maxConcurrency: 0, // No limit by default
		timeout:        10 * time.Minute,
	}
}

// Execute runs tasks in parallel with the given configuration
func (pe *ParallelExecutorImpl) Execute(ctx context.Context, tasks []domain.ExecutableTask) error {
	if len(tasks) == 0 {
		return nil
	}

	// Apply timeout if configured
	if pe.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, pe.timeout)
		defer cancel()
	}

	// Create a channel for task execution if concurrency is limited
	var semaphore chan struct{}
	if pe.maxConcurrency > 0 {
		semaphore = make(chan struct{}, pe.maxConcurrency)
	}

	// Execute tasks
	var wg sync.WaitGroup
	errChan := make(chan error, len(tasks))

	for _, task := range tasks {
		if !task.IsEnabled() {
			continue
		}

		wg.Add(1)
		go func(t domain.ExecutableTask) {
			defer wg.Done()

			// Acquire semaphore if concurrency is limited
			if semaphore != nil {
				semaphore <- struct{}{}
				defer func() { <-semaphore }()
			}

			// Check context before executing
			select {
			case <-ctx.Done():
				errChan <- fmt.Errorf("task %s cancelled: %w", t.Name(), ctx.Err())
				return
			default:
			}

			// Execute the task
			_, err := t.Execute(ctx)
			if err != nil {
				errChan <- fmt.Errorf("task %s failed: %w", t.Name(), err)
			}
		}(task)
	}

	// Wait for all tasks to complete
	doneChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	// Wait for completion or timeout
	select {
	case <-doneChan:
		// All tasks completed
		close(errChan)

		// Collect errors
		var errors []error
		for err := range errChan {
			errors = append(errors, err)
		}

		if len(errors) > 0 {
			return fmt.Errorf("parallel execution failed with %d errors: %v", len(errors), errors[0])
		}
		return nil

	case <-ctx.Done():
		// Timeout occurred
		return fmt.Errorf("parallel execution timed out after %v", pe.timeout)
	}
}

// SetMaxConcurrency sets the maximum number of concurrent tasks
func (pe *ParallelExecutorImpl) SetMaxConcurrency(max int) {
	pe.maxConcurrency = max
}

// SetTimeout sets the timeout for all tasks
func (pe *ParallelExecutorImpl) SetTimeout(timeout time.Duration) {
	pe.timeout = timeout
}

// SimpleTask is a basic implementation of ExecutableTask
type SimpleTask struct {
	name    string
	enabled bool
	execute func(context.Context) (interface{}, error)
}

// NewSimpleTask creates a new simple task
func NewSimpleTask(name string, enabled bool, execute func(context.Context) (interface{}, error)) domain.ExecutableTask {
	return &SimpleTask{
		name:    name,
		enabled: enabled,
		execute: execute,
	}
}

// Name returns the name of the task
func (t *SimpleTask) Name() string {
	return t.name
}

// Execute runs the task and returns the result
func (t *SimpleTask) Execute(ctx context.Context) (interface{}, error) {
	if t.execute == nil {
		return nil, fmt.Errorf("task %s has no execute function", t.name)
	}
	return t.execute(ctx)
}

// IsEnabled returns whether the task should be executed
func (t *SimpleTask) IsEnabled() bool {
	return t.enabled
}