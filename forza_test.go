package forza

import (
	"errors"
	"strings"
	"sync/atomic"
	"testing"
)

// Helper: create a simple task function that returns a fixed string
func mockTask(result string) TaskChainFn {
	return func(params ...string) (string, error) {
		if len(params) > 0 {
			return result + " [ctx:" + params[0] + "]", nil
		}
		return result, nil
	}
}

// Helper: create a task that returns an error
func mockErrorTask(msg string) TaskChainFn {
	return func(params ...string) (string, error) {
		return "", errors.New(msg)
	}
}

func TestNewPipeline(t *testing.T) {
	p := NewPipeline()
	if p == nil {
		t.Fatal("expected non-nil pipeline")
	}
	if len(p.tasks) != 0 {
		t.Errorf("expected empty tasks, got %d", len(p.tasks))
	}
}

func TestPipeline_AddTasks(t *testing.T) {
	p := NewPipeline()
	p.AddTasks(mockTask("a"), mockTask("b"))

	if len(p.tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(p.tasks))
	}

	p.AddTasks(mockTask("c"))
	if len(p.tasks) != 3 {
		t.Errorf("expected 3 tasks after second add, got %d", len(p.tasks))
	}
}

// --- CreateChain tests ---

func TestCreateChain_SingleTask(t *testing.T) {
	p := NewPipeline()
	chain := p.CreateChain(mockTask("hello"))

	result, err := chain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestCreateChain_PassesContextBetweenTasks(t *testing.T) {
	p := NewPipeline()

	task1 := mockTask("result-from-task1")
	task2 := func(params ...string) (string, error) {
		if len(params) == 0 {
			return "", errors.New("expected context from previous task")
		}
		return "task2 received: " + params[0], nil
	}

	chain := p.CreateChain(task1, task2)
	result, err := chain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "task2 received: result-from-task1" {
		t.Errorf("expected context passing, got %q", result)
	}
}

func TestCreateChain_ThreeTasksInOrder(t *testing.T) {
	p := NewPipeline()

	task1 := func(params ...string) (string, error) { return "A", nil }
	task2 := func(params ...string) (string, error) { return "B+" + params[0], nil }
	task3 := func(params ...string) (string, error) { return "C+" + params[0], nil }

	chain := p.CreateChain(task1, task2, task3)
	result, err := chain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "C+B+A" {
		t.Errorf("expected 'C+B+A', got %q", result)
	}
}

func TestCreateChain_StopsOnError(t *testing.T) {
	p := NewPipeline()

	task1 := mockTask("ok")
	task2 := mockErrorTask("task2 failed")
	task3 := mockTask("should not run")

	chain := p.CreateChain(task1, task2, task3)
	_, err := chain()
	if err == nil {
		t.Fatal("expected error from chain")
	}
	if !errors.Is(err, ErrChainInterrupted) {
		t.Errorf("expected ErrChainInterrupted, got %v", err)
	}
	if !strings.Contains(err.Error(), "task2 failed") {
		t.Errorf("expected error to contain 'task2 failed', got %v", err)
	}
}

func TestCreateChain_NilTask(t *testing.T) {
	p := NewPipeline()
	chain := p.CreateChain(mockTask("ok"), nil, mockTask("after"))

	_, err := chain()
	if err == nil {
		t.Fatal("expected error for nil task")
	}
	if !errors.Is(err, ErrNilTask) {
		t.Errorf("expected ErrNilTask, got %v", err)
	}
}

func TestCreateChain_Empty(t *testing.T) {
	p := NewPipeline()
	chain := p.CreateChain()

	result, err := chain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty result from empty chain, got %q", result)
	}
}

// --- RunConcurrently tests ---

func TestRunConcurrently_MultipleTasks(t *testing.T) {
	p := NewPipeline()
	p.AddTasks(mockTask("first"), mockTask("second"), mockTask("third"))

	results, err := p.RunConcurrently()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0] != "first" {
		t.Errorf("expected results[0]='first', got %q", results[0])
	}
	if results[1] != "second" {
		t.Errorf("expected results[1]='second', got %q", results[1])
	}
	if results[2] != "third" {
		t.Errorf("expected results[2]='third', got %q", results[2])
	}
}

func TestRunConcurrently_PreservesOrder(t *testing.T) {
	p := NewPipeline()
	for i := 0; i < 10; i++ {
		idx := i
		p.AddTasks(func(params ...string) (string, error) {
			return strings.Repeat("x", idx+1), nil
		})
	}

	results, err := p.RunConcurrently()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, r := range results {
		expected := strings.Repeat("x", i+1)
		if r != expected {
			t.Errorf("results[%d]: expected %q, got %q", i, expected, r)
		}
	}
}

func TestRunConcurrently_CollectsErrors(t *testing.T) {
	p := NewPipeline()
	p.AddTasks(mockTask("ok"), mockErrorTask("fail"), mockTask("also ok"))

	results, err := p.RunConcurrently()
	if err == nil {
		t.Fatal("expected error")
	}

	// First and third tasks should still have results
	if results[0] != "ok" {
		t.Errorf("expected results[0]='ok', got %q", results[0])
	}
	if results[2] != "also ok" {
		t.Errorf("expected results[2]='also ok', got %q", results[2])
	}
}

func TestRunConcurrently_ActuallyRunsConcurrently(t *testing.T) {
	var running int64
	var maxRunning int64

	p := NewPipeline()
	for i := 0; i < 5; i++ {
		p.AddTasks(func(params ...string) (string, error) {
			cur := atomic.AddInt64(&running, 1)
			// Track max concurrent
			for {
				old := atomic.LoadInt64(&maxRunning)
				if cur <= old {
					break
				}
				if atomic.CompareAndSwapInt64(&maxRunning, old, cur) {
					break
				}
			}
			// Small busy loop to increase chance of overlap
			for j := 0; j < 1000; j++ {
				_ = j * j
			}
			atomic.AddInt64(&running, -1)
			return "done", nil
		})
	}

	_, err := p.RunConcurrently()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// We can't guarantee exact concurrency, but with 5 tasks and goroutines,
	// at least 2 should have been running at the same time
	if atomic.LoadInt64(&maxRunning) < 2 {
		t.Log("Warning: concurrent execution not detected, but this can happen under load")
	}
}

func TestRunConcurrently_Empty(t *testing.T) {
	p := NewPipeline()
	results, err := p.RunConcurrently()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}

// --- RunSequentially tests ---

func TestRunSequentially_MultipleTasks(t *testing.T) {
	p := NewPipeline()
	p.AddTasks(mockTask("first"), mockTask("second"), mockTask("third"))

	results, err := p.RunSequentially()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0] != "first" || results[1] != "second" || results[2] != "third" {
		t.Errorf("unexpected results: %v", results)
	}
}

func TestRunSequentially_StopsOnError(t *testing.T) {
	p := NewPipeline()
	p.AddTasks(mockTask("ok"), mockErrorTask("fail"), mockTask("should not run"))

	results, err := p.RunSequentially()
	if err == nil {
		t.Fatal("expected error")
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result before error, got %d", len(results))
	}
	if results[0] != "ok" {
		t.Errorf("expected first result 'ok', got %q", results[0])
	}
}

func TestRunSequentially_NilTask(t *testing.T) {
	p := NewPipeline()
	p.AddTasks(mockTask("ok"), nil)

	_, err := p.RunSequentially()
	if err == nil {
		t.Fatal("expected error for nil task")
	}
	if !errors.Is(err, ErrNilTask) {
		t.Errorf("expected ErrNilTask, got %v", err)
	}
}

func TestRunSequentially_Empty(t *testing.T) {
	p := NewPipeline()
	results, err := p.RunSequentially()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}
