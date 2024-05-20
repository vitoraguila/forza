package forza

import "sync"

type Forza struct {
	tasks []TaskFn
}

type TaskFn func() string

type ForzaService interface {
	RunConcurrently() []string
	AddTask(fn TaskFn)
}

func NewForza() ForzaService {
	return &Forza{}
}

func (f *Forza) AddTask(fn TaskFn) {
	f.tasks = append(f.tasks, fn)
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
			// Execute the task and send the result to the channel
			result := task()
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
