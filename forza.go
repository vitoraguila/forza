package forza

import (
	"fmt"
	"sync"
	"time"
)

type forza struct {
	tasks []taskChainFn
}
type taskFn func() string
type taskChainFn func(...string) string

func NewPipeline() *forza {
	return &forza{}
}

func (f *forza) AddTasks(fn ...taskChainFn) {
	f.tasks = append(f.tasks, fn...)
}

func (f *forza) CreateChain(tasks ...taskChainFn) taskFn {
	return func() string {
		var result string
		for i, task := range tasks {
			if task == nil {
				panic("There is no action pre set for the task.")
			}
			fmt.Printf("Task %d of %d chains\n", i+1, len(tasks))

			if i == 0 {
				result = task()
				continue
			}

			if (i + 1) < len(tasks) {
				result = tasks[i+1](result)
				continue
			}
		}
		return result
	}
}

func (f *forza) RunConcurrently() []string {
	var wg sync.WaitGroup
	results := make([]string, len(f.tasks))
	resultsChan := make(chan struct {
		index  int
		result string
	}, len(f.tasks))

	// Launch each task in a separate goroutine
	for i, task := range f.tasks {
		wg.Add(1)
		go func(index int, task taskChainFn) {
			defer wg.Done()
			startTime := time.Now()
			fmt.Printf("Task %d started at %s\n", index+1, startTime.Format("15:04:05.000"))

			// Execute the task and send the result to the channel
			result := task()

			endTime := time.Now()
			fmt.Printf("Task %d finished at %s (Duration: %s)\n", index+1, endTime.Format("15:04:05.000"), endTime.Sub(startTime))

			resultsChan <- struct {
				index  int
				result string
			}{index, result}
		}(i, task)
	}

	// Close the results channel once all tasks are done
	go func() {
		wg.Wait()
		close(resultsChan)

	}()

	// Collect the results from the channel
	for result := range resultsChan {
		results[result.index] = result.result
	}

	return results
}
