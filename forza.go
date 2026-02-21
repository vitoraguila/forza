package forza

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

// TaskChainFn is a function that takes a context and optional context strings and returns a result or error.
type TaskChainFn func(context.Context, ...string) (string, error)

// TaskFn is a function that takes a context and returns a result or error.
type TaskFn func(context.Context) (string, error)

// Pipeline orchestrates the execution of multiple LLM tasks.
type Pipeline struct {
	tasks  []TaskChainFn
	logger *slog.Logger
}

// NewPipeline creates a new empty Pipeline.
func NewPipeline() *Pipeline {
	return &Pipeline{}
}

// WithLogger sets an optional logger for the pipeline. If nil, no logging occurs.
func (p *Pipeline) WithLogger(l *slog.Logger) *Pipeline {
	p.logger = l
	return p
}

func (p *Pipeline) logDebug(msg string, args ...any) {
	if p.logger != nil {
		p.logger.Debug(msg, args...)
	}
}

// AddTasks appends one or more task functions to the pipeline.
func (p *Pipeline) AddTasks(fn ...TaskChainFn) {
	p.tasks = append(p.tasks, fn...)
}

// CreateChain returns a TaskFn that executes tasks sequentially, passing each
// task's result as context to the next task. If any task returns an error,
// the chain stops and the error is returned.
func (p *Pipeline) CreateChain(tasks ...TaskChainFn) TaskFn {
	return func(ctx context.Context) (string, error) {
		var result string
		for i, task := range tasks {
			if task == nil {
				return "", fmt.Errorf("%w: task at index %d", ErrNilTask, i)
			}
			p.logDebug("chain progress", "task", i+1, "total", len(tasks))

			var err error
			if i == 0 {
				result, err = task(ctx)
			} else {
				result, err = task(ctx, result)
			}
			if err != nil {
				return "", fmt.Errorf("%w: task %d failed: %v", ErrChainInterrupted, i+1, err)
			}
		}
		return result, nil
	}
}

// RunConcurrently executes all added tasks concurrently and returns their results
// in the original order. If any task fails or panics, its error is collected and
// returned as a combined error after all tasks complete.
func (p *Pipeline) RunConcurrently(ctx context.Context) ([]string, error) {
	var wg sync.WaitGroup
	results := make([]string, len(p.tasks))
	errs := make([]error, len(p.tasks))

	type taskResult struct {
		index  int
		result string
		err    error
	}
	resultsChan := make(chan taskResult, len(p.tasks))

	for i, task := range p.tasks {
		if task == nil {
			errs[i] = fmt.Errorf("%w: task at index %d", ErrNilTask, i)
			continue
		}
		wg.Add(1)
		go func(index int, task TaskChainFn) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					resultsChan <- taskResult{
						index: index,
						err:   fmt.Errorf("task %d panicked: %v", index+1, r),
					}
				}
			}()
			p.logDebug("task started", "task", index+1)

			result, err := task(ctx)

			p.logDebug("task finished", "task", index+1)

			resultsChan <- taskResult{index: index, result: result, err: err}
		}(i, task)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for r := range resultsChan {
		results[r.index] = r.result
		errs[r.index] = r.err
	}

	// Collect errors
	var combinedErr error
	for i, err := range errs {
		if err != nil {
			if combinedErr == nil {
				combinedErr = fmt.Errorf("task %d: %w", i+1, err)
			} else {
				combinedErr = fmt.Errorf("%v; task %d: %w", combinedErr, i+1, err)
			}
		}
	}

	return results, combinedErr
}

// RunSequentially executes all added tasks one after another. Each task receives
// no context arguments. If any task fails, execution stops and the error is returned.
func (p *Pipeline) RunSequentially(ctx context.Context) ([]string, error) {
	results := make([]string, 0, len(p.tasks))
	for i, task := range p.tasks {
		if task == nil {
			return results, fmt.Errorf("%w: task at index %d", ErrNilTask, i)
		}
		p.logDebug("sequential progress", "task", i+1, "total", len(p.tasks))

		result, err := task(ctx)
		if err != nil {
			return results, fmt.Errorf("task %d failed: %w", i+1, err)
		}
		results = append(results, result)
	}
	return results, nil
}
