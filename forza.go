package forza

import (
	"fmt"
	"sync"
	"time"
)

// TaskChainFn is a function that takes optional context strings and returns a result or error.
type TaskChainFn func(...string) (string, error)

// TaskFn is a function that takes no arguments and returns a result or error.
type TaskFn func() (string, error)

// Pipeline orchestrates the execution of multiple LLM tasks.
type Pipeline struct {
	tasks []TaskChainFn
}

// NewPipeline creates a new empty Pipeline.
func NewPipeline() *Pipeline {
	return &Pipeline{}
}

// AddTasks appends one or more task functions to the pipeline.
func (p *Pipeline) AddTasks(fn ...TaskChainFn) {
	p.tasks = append(p.tasks, fn...)
}

// CreateChain returns a TaskFn that executes tasks sequentially, passing each
// task's result as context to the next task. If any task returns an error,
// the chain stops and the error is returned.
func (p *Pipeline) CreateChain(tasks ...TaskChainFn) TaskFn {
	return func() (string, error) {
		var result string
		for i, task := range tasks {
			if task == nil {
				return "", fmt.Errorf("%w: task at index %d", ErrNilTask, i)
			}
			fmt.Printf("Task %d of %d chains\n", i+1, len(tasks))

			var err error
			if i == 0 {
				result, err = task()
			} else {
				result, err = task(result)
			}
			if err != nil {
				return "", fmt.Errorf("%w: task %d failed: %v", ErrChainInterrupted, i+1, err)
			}
		}
		return result, nil
	}
}

// RunConcurrently executes all added tasks concurrently and returns their results
// in the original order. If any task fails, its error is collected and returned
// as a combined error after all tasks complete.
func (p *Pipeline) RunConcurrently() ([]string, error) {
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
		wg.Add(1)
		go func(index int, task TaskChainFn) {
			defer wg.Done()
			startTime := time.Now()
			fmt.Printf("Task %d started at %s\n", index+1, startTime.Format("15:04:05.000"))

			result, err := task()

			endTime := time.Now()
			fmt.Printf("Task %d finished at %s (Duration: %s)\n", index+1, endTime.Format("15:04:05.000"), endTime.Sub(startTime))

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
func (p *Pipeline) RunSequentially() ([]string, error) {
	results := make([]string, 0, len(p.tasks))
	for i, task := range p.tasks {
		if task == nil {
			return results, fmt.Errorf("%w: task at index %d", ErrNilTask, i)
		}
		fmt.Printf("Task %d of %d\n", i+1, len(p.tasks))

		result, err := task()
		if err != nil {
			return results, fmt.Errorf("task %d failed: %w", i+1, err)
		}
		results = append(results, result)
	}
	return results, nil
}
