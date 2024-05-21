package forza

import (
	"fmt"
	"sync"
	"time"
)

type Forza struct {
	tasks []TaskFn
}

type TaskFn func() string

func NewPipeline() *Forza {
	return &Forza{}
}

func (f *Forza) AddTasks(fn ...TaskFn) {
	f.tasks = append(f.tasks, fn...)
}

func (f *Forza) RunConcurrently() []string {
	var wg sync.WaitGroup
	results := make([]string, len(f.tasks))
	resultsChan := make(chan struct {
		index  int
		result string
	}, len(f.tasks))

	// Launch each task in a separate goroutine
	for i, task := range f.tasks {
		wg.Add(1)
		go func(index int, task TaskFn) {
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
