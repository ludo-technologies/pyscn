package service

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewParallelExecutor(t *testing.T) {
	executor := NewParallelExecutor()

	assert.NotNil(t, executor)

	impl, ok := executor.(*ParallelExecutorImpl)
	require.True(t, ok)
	assert.Equal(t, 0, impl.maxConcurrency)
	assert.Equal(t, 10*time.Minute, impl.timeout)
}

func TestParallelExecutor_Execute_EmptyTasks(t *testing.T) {
	executor := NewParallelExecutor()
	ctx := context.Background()

	err := executor.Execute(ctx, []domain.ExecutableTask{})
	assert.NoError(t, err)
}

func TestParallelExecutor_Execute_SingleTask(t *testing.T) {
	executor := NewParallelExecutor()
	ctx := context.Background()

	executed := false
	task := NewSimpleTask("test-task", true, func(ctx context.Context) (interface{}, error) {
		executed = true
		return "result", nil
	})

	err := executor.Execute(ctx, []domain.ExecutableTask{task})
	assert.NoError(t, err)
	assert.True(t, executed)
}

func TestParallelExecutor_Execute_MultipleTasks(t *testing.T) {
	executor := NewParallelExecutor()
	ctx := context.Background()

	var counter int32
	tasks := make([]domain.ExecutableTask, 5)
	for i := 0; i < 5; i++ {
		tasks[i] = NewSimpleTask("task", true, func(ctx context.Context) (interface{}, error) {
			atomic.AddInt32(&counter, 1)
			return nil, nil
		})
	}

	err := executor.Execute(ctx, tasks)
	assert.NoError(t, err)
	assert.Equal(t, int32(5), counter)
}

func TestParallelExecutor_Execute_DisabledTasks(t *testing.T) {
	executor := NewParallelExecutor()
	ctx := context.Background()

	executed := false
	task := NewSimpleTask("disabled-task", false, func(ctx context.Context) (interface{}, error) {
		executed = true
		return nil, nil
	})

	err := executor.Execute(ctx, []domain.ExecutableTask{task})
	assert.NoError(t, err)
	assert.False(t, executed, "disabled task should not be executed")
}

func TestParallelExecutor_Execute_TaskError(t *testing.T) {
	executor := NewParallelExecutor()
	ctx := context.Background()

	expectedErr := errors.New("task failed")
	task := NewSimpleTask("failing-task", true, func(ctx context.Context) (interface{}, error) {
		return nil, expectedErr
	})

	err := executor.Execute(ctx, []domain.ExecutableTask{task})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failing-task")
	assert.Contains(t, err.Error(), "task failed")
}

func TestParallelExecutor_Execute_MultipleErrors(t *testing.T) {
	executor := NewParallelExecutor()
	ctx := context.Background()

	tasks := []domain.ExecutableTask{
		NewSimpleTask("fail-1", true, func(ctx context.Context) (interface{}, error) {
			return nil, errors.New("error 1")
		}),
		NewSimpleTask("fail-2", true, func(ctx context.Context) (interface{}, error) {
			return nil, errors.New("error 2")
		}),
	}

	err := executor.Execute(ctx, tasks)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "2 errors")
}

func TestParallelExecutor_Execute_ContextCancellation(t *testing.T) {
	executor := NewParallelExecutor()
	ctx, cancel := context.WithCancel(context.Background())

	started := make(chan struct{})
	task := NewSimpleTask("long-task", true, func(ctx context.Context) (interface{}, error) {
		close(started)
		<-ctx.Done()
		return nil, ctx.Err()
	})

	go func() {
		<-started
		cancel()
	}()

	err := executor.Execute(ctx, []domain.ExecutableTask{task})
	assert.Error(t, err)
}

func TestParallelExecutor_SetMaxConcurrency(t *testing.T) {
	executor := NewParallelExecutor()
	impl := executor.(*ParallelExecutorImpl)

	impl.SetMaxConcurrency(4)
	assert.Equal(t, 4, impl.maxConcurrency)
}

func TestParallelExecutor_SetTimeout(t *testing.T) {
	executor := NewParallelExecutor()
	impl := executor.(*ParallelExecutorImpl)

	impl.SetTimeout(5 * time.Second)
	assert.Equal(t, 5*time.Second, impl.timeout)
}

func TestParallelExecutor_Execute_WithConcurrencyLimit(t *testing.T) {
	executor := NewParallelExecutor()
	impl := executor.(*ParallelExecutorImpl)
	impl.SetMaxConcurrency(2)

	ctx := context.Background()
	var maxConcurrent int32
	var currentConcurrent int32

	tasks := make([]domain.ExecutableTask, 5)
	for i := 0; i < 5; i++ {
		tasks[i] = NewSimpleTask("task", true, func(ctx context.Context) (interface{}, error) {
			current := atomic.AddInt32(&currentConcurrent, 1)
			for {
				max := atomic.LoadInt32(&maxConcurrent)
				if current > max {
					if atomic.CompareAndSwapInt32(&maxConcurrent, max, current) {
						break
					}
				} else {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt32(&currentConcurrent, -1)
			return nil, nil
		})
	}

	err := executor.Execute(ctx, tasks)
	assert.NoError(t, err)
	assert.LessOrEqual(t, maxConcurrent, int32(2), "max concurrent should not exceed limit")
}

func TestParallelExecutor_Execute_Timeout(t *testing.T) {
	executor := NewParallelExecutor()
	impl := executor.(*ParallelExecutorImpl)
	impl.SetTimeout(50 * time.Millisecond)

	ctx := context.Background()
	task := NewSimpleTask("slow-task", true, func(ctx context.Context) (interface{}, error) {
		time.Sleep(200 * time.Millisecond)
		return nil, nil
	})

	err := executor.Execute(ctx, []domain.ExecutableTask{task})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
}

func TestSimpleTask_Name(t *testing.T) {
	task := NewSimpleTask("my-task", true, nil)
	assert.Equal(t, "my-task", task.Name())
}

func TestSimpleTask_IsEnabled(t *testing.T) {
	enabledTask := NewSimpleTask("enabled", true, nil)
	assert.True(t, enabledTask.IsEnabled())

	disabledTask := NewSimpleTask("disabled", false, nil)
	assert.False(t, disabledTask.IsEnabled())
}

func TestSimpleTask_Execute_NilFunction(t *testing.T) {
	task := NewSimpleTask("nil-func", true, nil)

	result, err := task.Execute(context.Background())
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no execute function")
}

func TestSimpleTask_Execute_Success(t *testing.T) {
	expected := "test-result"
	task := NewSimpleTask("success", true, func(ctx context.Context) (interface{}, error) {
		return expected, nil
	})

	result, err := task.Execute(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestSimpleTask_Execute_Error(t *testing.T) {
	expectedErr := errors.New("execution failed")
	task := NewSimpleTask("error", true, func(ctx context.Context) (interface{}, error) {
		return nil, expectedErr
	})

	result, err := task.Execute(context.Background())
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, result)
}
