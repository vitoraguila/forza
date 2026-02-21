package forza

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Helper: create a simple task function that returns a fixed string
func mockTask(result string) TaskChainFn {
	return func(ctx context.Context, params ...string) (string, error) {
		if len(params) > 0 {
			return result + " [ctx:" + params[0] + "]", nil
		}
		return result, nil
	}
}

// Helper: create a task that returns an error
func mockErrorTask(msg string) TaskChainFn {
	return func(ctx context.Context, params ...string) (string, error) {
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

	result, err := chain(context.Background())
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
	task2 := func(ctx context.Context, params ...string) (string, error) {
		if len(params) == 0 {
			return "", errors.New("expected context from previous task")
		}
		return "task2 received: " + params[0], nil
	}

	chain := p.CreateChain(task1, task2)
	result, err := chain(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "task2 received: result-from-task1" {
		t.Errorf("expected context passing, got %q", result)
	}
}

func TestCreateChain_ThreeTasksInOrder(t *testing.T) {
	p := NewPipeline()

	task1 := func(ctx context.Context, params ...string) (string, error) { return "A", nil }
	task2 := func(ctx context.Context, params ...string) (string, error) { return "B+" + params[0], nil }
	task3 := func(ctx context.Context, params ...string) (string, error) { return "C+" + params[0], nil }

	chain := p.CreateChain(task1, task2, task3)
	result, err := chain(context.Background())
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
	_, err := chain(context.Background())
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

	_, err := chain(context.Background())
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

	result, err := chain(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty result from empty chain, got %q", result)
	}
}

func TestCreateChain_CancelledContext(t *testing.T) {
	p := NewPipeline()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	task := func(ctx context.Context, params ...string) (string, error) {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "should not reach", nil
	}

	chain := p.CreateChain(task)
	_, err := chain(ctx)
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
}

// --- RunConcurrently tests ---

func TestRunConcurrently_MultipleTasks(t *testing.T) {
	p := NewPipeline()
	p.AddTasks(mockTask("first"), mockTask("second"), mockTask("third"))

	results, err := p.RunConcurrently(context.Background())
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
		p.AddTasks(func(ctx context.Context, params ...string) (string, error) {
			return strings.Repeat("x", idx+1), nil
		})
	}

	results, err := p.RunConcurrently(context.Background())
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

	results, err := p.RunConcurrently(context.Background())
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
	const numTasks = 5
	var started sync.WaitGroup
	started.Add(numTasks)
	barrier := make(chan struct{})

	p := NewPipeline()
	for i := 0; i < numTasks; i++ {
		p.AddTasks(func(ctx context.Context, params ...string) (string, error) {
			started.Done() // Signal "I'm running"
			<-barrier      // Wait for all to be running
			return "done", nil
		})
	}

	done := make(chan struct{})
	var results []string
	var runErr error

	go func() {
		results, runErr = p.RunConcurrently(context.Background())
		close(done)
	}()

	// Wait for all goroutines to signal they've started
	started.Wait()
	// Release them all at once
	close(barrier)

	<-done
	if runErr != nil {
		t.Fatalf("unexpected error: %v", runErr)
	}
	if len(results) != numTasks {
		t.Errorf("expected %d results, got %d", numTasks, len(results))
	}
}

func TestRunConcurrently_NilTask(t *testing.T) {
	p := NewPipeline()
	p.AddTasks(mockTask("ok"), nil, mockTask("also ok"))

	results, err := p.RunConcurrently(context.Background())
	if err == nil {
		t.Fatal("expected error for nil task")
	}
	if !strings.Contains(err.Error(), "nil") {
		t.Errorf("expected nil task error, got %v", err)
	}
	// Other tasks should still complete
	if results[0] != "ok" {
		t.Errorf("expected results[0]='ok', got %q", results[0])
	}
}

func TestRunConcurrently_TaskPanics(t *testing.T) {
	p := NewPipeline()
	p.AddTasks(
		mockTask("ok"),
		func(ctx context.Context, params ...string) (string, error) {
			panic("test panic")
		},
		mockTask("also ok"),
	)

	results, err := p.RunConcurrently(context.Background())
	if err == nil {
		t.Fatal("expected error from panicking task")
	}
	if !strings.Contains(err.Error(), "panicked") {
		t.Errorf("expected panic error, got %v", err)
	}
	// Non-panicking tasks should still have results
	if results[0] != "ok" {
		t.Errorf("expected results[0]='ok', got %q", results[0])
	}
	if results[2] != "also ok" {
		t.Errorf("expected results[2]='also ok', got %q", results[2])
	}
}

func TestRunConcurrently_MixedPanicsAndErrors(t *testing.T) {
	p := NewPipeline()
	p.AddTasks(
		mockTask("success"),
		func(ctx context.Context, params ...string) (string, error) {
			panic("boom")
		},
		mockErrorTask("regular error"),
		mockTask("another success"),
	)

	results, err := p.RunConcurrently(context.Background())
	if err == nil {
		t.Fatal("expected combined errors")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "panicked") {
		t.Errorf("expected panic error in combined error, got %v", err)
	}
	if !strings.Contains(errStr, "regular error") {
		t.Errorf("expected regular error in combined error, got %v", err)
	}
	if results[0] != "success" {
		t.Errorf("expected results[0]='success', got %q", results[0])
	}
	if results[3] != "another success" {
		t.Errorf("expected results[3]='another success', got %q", results[3])
	}
}

func TestRunConcurrently_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var taskStarted int64
	p := NewPipeline()
	for i := 0; i < 5; i++ {
		p.AddTasks(func(ctx context.Context, params ...string) (string, error) {
			atomic.AddInt64(&taskStarted, 1)
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(5 * time.Second):
				return "done", nil
			}
		})
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := p.RunConcurrently(ctx)
	if err == nil {
		t.Fatal("expected error from canceled context")
	}
}

func TestRunConcurrently_Empty(t *testing.T) {
	p := NewPipeline()
	results, err := p.RunConcurrently(context.Background())
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

	results, err := p.RunSequentially(context.Background())
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

	results, err := p.RunSequentially(context.Background())
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

	_, err := p.RunSequentially(context.Background())
	if err == nil {
		t.Fatal("expected error for nil task")
	}
	if !errors.Is(err, ErrNilTask) {
		t.Errorf("expected ErrNilTask, got %v", err)
	}
}

func TestRunSequentially_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := NewPipeline()
	p.AddTasks(func(ctx context.Context, params ...string) (string, error) {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "should not reach", nil
	})

	_, err := p.RunSequentially(ctx)
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
}

func TestRunSequentially_Empty(t *testing.T) {
	p := NewPipeline()
	results, err := p.RunSequentially(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}
