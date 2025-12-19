package core

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// MockTask is a test task implementation.
type MockTask struct {
	BaseTask
	ExecuteFunc  func(*InstallContext, *EventBus) error
	RollbackFunc func(*InstallContext, *EventBus) error
	ValidateFunc func() error
	rollbackable bool
	executed     bool
	rolledBack   bool
}

func NewMockTask(id, taskType string) *MockTask {
	return &MockTask{
		BaseTask: BaseTask{
			TaskID:   id,
			TaskType: taskType,
			Config:   make(map[string]any),
		},
	}
}

func (t *MockTask) Validate() error {
	if t.ValidateFunc != nil {
		return t.ValidateFunc()
	}
	return nil
}

func (t *MockTask) Execute(ctx *InstallContext, bus *EventBus) error {
	t.executed = true
	if t.ExecuteFunc != nil {
		return t.ExecuteFunc(ctx, bus)
	}
	return nil
}

func (t *MockTask) Rollback(ctx *InstallContext, bus *EventBus) error {
	t.rolledBack = true
	if t.RollbackFunc != nil {
		return t.RollbackFunc(ctx, bus)
	}
	return nil
}

func (t *MockTask) CanRollback() bool {
	return t.rollbackable
}

// Tests for TaskState
func TestTaskStateString(t *testing.T) {
	tests := []struct {
		state    TaskState
		expected string
	}{
		{TaskPending, "pending"},
		{TaskRunning, "running"},
		{TaskCompleted, "completed"},
		{TaskFailed, "failed"},
		{TaskCancelled, "cancelled"},
		{TaskRolledBack, "rolled_back"},
		{TaskState(99), "unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			if tc.state.String() != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, tc.state.String())
			}
		})
	}
}

// Tests for BaseTask
func TestBaseTask(t *testing.T) {
	task := &BaseTask{
		TaskID:   "test-task",
		TaskType: "test",
		Config: map[string]any{
			"str":   "hello",
			"bool":  true,
			"int":   42,
			"int64": int64(100),
			"float": 3.14,
		},
	}

	t.Run("ID", func(t *testing.T) {
		if task.ID() != "test-task" {
			t.Errorf("expected 'test-task', got %q", task.ID())
		}
	})

	t.Run("Type", func(t *testing.T) {
		if task.Type() != "test" {
			t.Errorf("expected 'test', got %q", task.Type())
		}
	})

	t.Run("CanRollback", func(t *testing.T) {
		if task.CanRollback() {
			t.Error("expected CanRollback to return false by default")
		}
	})

	t.Run("Rollback", func(t *testing.T) {
		ctx := NewInstallContext()
		bus := NewEventBus()
		if err := task.Rollback(ctx, bus); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("GetConfigString", func(t *testing.T) {
		if task.GetConfigString("str") != "hello" {
			t.Errorf("expected 'hello', got %q", task.GetConfigString("str"))
		}
		if task.GetConfigString("missing") != "" {
			t.Error("expected empty string for missing key")
		}
	})

	t.Run("GetConfigBool", func(t *testing.T) {
		if !task.GetConfigBool("bool") {
			t.Error("expected true")
		}
		if task.GetConfigBool("missing") {
			t.Error("expected false for missing key")
		}
	})

	t.Run("GetConfigInt", func(t *testing.T) {
		if task.GetConfigInt("int") != 42 {
			t.Errorf("expected 42, got %d", task.GetConfigInt("int"))
		}
		if task.GetConfigInt("int64") != 100 {
			t.Errorf("expected 100, got %d", task.GetConfigInt("int64"))
		}
		if task.GetConfigInt("float") != 3 {
			t.Errorf("expected 3, got %d", task.GetConfigInt("float"))
		}
		if task.GetConfigInt("missing") != 0 {
			t.Error("expected 0 for missing key")
		}
	})
}

// Tests for TaskRunner
func TestNewTaskRunner(t *testing.T) {
	ctx := NewInstallContext()
	bus := NewEventBus()

	runner := NewTaskRunner(ctx, bus)

	if runner == nil {
		t.Fatal("expected non-nil runner")
	}
	if runner.IsRunning() {
		t.Error("expected runner to not be running initially")
	}
	if runner.Progress() != 1.0 {
		t.Errorf("expected progress 1.0 for empty runner, got %f", runner.Progress())
	}
}

func TestTaskRunnerAddTask(t *testing.T) {
	ctx := NewInstallContext()
	bus := NewEventBus()
	runner := NewTaskRunner(ctx, bus)

	task1 := NewMockTask("task-1", "mock")
	task2 := NewMockTask("task-2", "mock")

	runner.AddTask(task1)
	runner.AddTasks([]Task{task2})

	if len(runner.tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(runner.tasks))
	}
}

func TestTaskRunnerRun(t *testing.T) {
	ctx := NewInstallContext()
	bus := NewEventBus()
	runner := NewTaskRunner(ctx, bus)

	executed := make([]string, 0)
	var mu sync.Mutex

	for i := 0; i < 3; i++ {
		task := NewMockTask("task-"+string(rune('a'+i)), "mock")
		taskID := task.TaskID
		task.ExecuteFunc = func(ctx *InstallContext, bus *EventBus) error {
			mu.Lock()
			executed = append(executed, taskID)
			mu.Unlock()
			return nil
		}
		runner.AddTask(task)
	}

	err := runner.Run()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(executed) != 3 {
		t.Errorf("expected 3 tasks executed, got %d", len(executed))
	}

	results := runner.Results()
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	for _, result := range results {
		if result.State != TaskCompleted {
			t.Errorf("expected task %s to be completed, got %s", result.TaskID, result.State)
		}
	}
}

func TestTaskRunnerRunEmpty(t *testing.T) {
	ctx := NewInstallContext()
	bus := NewEventBus()
	runner := NewTaskRunner(ctx, bus)

	err := runner.Run()
	if err != nil {
		t.Fatalf("expected no error for empty runner, got %v", err)
	}
}

func TestTaskRunnerFailureAbort(t *testing.T) {
	ctx := NewInstallContext()
	bus := NewEventBus()
	runner := NewTaskRunner(ctx, bus)
	runner.SetFailurePolicy(FailureAbort)

	task1 := NewMockTask("task-1", "mock")
	task2 := NewMockTask("task-2", "mock")
	task2.ExecuteFunc = func(ctx *InstallContext, bus *EventBus) error {
		return errors.New("task failed")
	}
	task3 := NewMockTask("task-3", "mock")

	runner.AddTasks([]Task{task1, task2, task3})

	err := runner.Run()
	if err == nil {
		t.Fatal("expected error")
	}

	if task3.executed {
		t.Error("task-3 should not have been executed after failure")
	}
}

func TestTaskRunnerFailureSkip(t *testing.T) {
	ctx := NewInstallContext()
	bus := NewEventBus()
	runner := NewTaskRunner(ctx, bus)
	runner.SetFailurePolicy(FailureSkip)

	task1 := NewMockTask("task-1", "mock")
	task2 := NewMockTask("task-2", "mock")
	task2.ExecuteFunc = func(ctx *InstallContext, bus *EventBus) error {
		return errors.New("task failed")
	}
	task3 := NewMockTask("task-3", "mock")

	runner.AddTasks([]Task{task1, task2, task3})

	err := runner.Run()
	if err != nil {
		t.Fatalf("expected no error with skip policy, got %v", err)
	}

	if !task3.executed {
		t.Error("task-3 should have been executed with skip policy")
	}
}

func TestTaskRunnerFailureRollback(t *testing.T) {
	ctx := NewInstallContext()
	bus := NewEventBus()
	runner := NewTaskRunner(ctx, bus)
	runner.SetFailurePolicy(FailureRollback)

	rollbackOrder := make([]string, 0)
	var mu sync.Mutex

	task1 := NewMockTask("task-1", "mock")
	task1.rollbackable = true
	task1.RollbackFunc = func(ctx *InstallContext, bus *EventBus) error {
		mu.Lock()
		rollbackOrder = append(rollbackOrder, "task-1")
		mu.Unlock()
		return nil
	}

	task2 := NewMockTask("task-2", "mock")
	task2.rollbackable = true
	task2.RollbackFunc = func(ctx *InstallContext, bus *EventBus) error {
		mu.Lock()
		rollbackOrder = append(rollbackOrder, "task-2")
		mu.Unlock()
		return nil
	}

	task3 := NewMockTask("task-3", "mock")
	task3.ExecuteFunc = func(ctx *InstallContext, bus *EventBus) error {
		return errors.New("task failed")
	}

	runner.AddTasks([]Task{task1, task2, task3})

	err := runner.Run()
	if err == nil {
		t.Fatal("expected error")
	}

	// Rollback should happen in reverse order
	if len(rollbackOrder) != 2 {
		t.Fatalf("expected 2 rollbacks, got %d", len(rollbackOrder))
	}
	if rollbackOrder[0] != "task-2" || rollbackOrder[1] != "task-1" {
		t.Errorf("expected rollback order [task-2, task-1], got %v", rollbackOrder)
	}
}

func TestTaskRunnerFailureRetry(t *testing.T) {
	ctx := NewInstallContext()
	bus := NewEventBus()
	runner := NewTaskRunner(ctx, bus)
	runner.SetFailurePolicy(FailureRetry)
	runner.SetMaxRetries(3)

	var attempts int32

	task := NewMockTask("retry-task", "mock")
	task.ExecuteFunc = func(ctx *InstallContext, bus *EventBus) error {
		count := atomic.AddInt32(&attempts, 1)
		if count < 3 {
			return errors.New("transient failure")
		}
		return nil // Succeed on 3rd attempt
	}

	runner.AddTask(task)

	err := runner.Run()
	if err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}

	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("expected 3 attempts, got %d", atomic.LoadInt32(&attempts))
	}
}

func TestTaskRunnerCancel(t *testing.T) {
	ctx := NewInstallContext()
	bus := NewEventBus()
	runner := NewTaskRunner(ctx, bus)

	// Add multiple tasks - cancellation happens between tasks
	task1 := NewMockTask("task-1", "mock")
	task1.ExecuteFunc = func(ctx *InstallContext, bus *EventBus) error {
		// Complete quickly
		return nil
	}

	task2 := NewMockTask("task-2", "mock")
	task2Executed := false
	task2.ExecuteFunc = func(ctx *InstallContext, bus *EventBus) error {
		task2Executed = true
		return nil
	}

	runner.AddTasks([]Task{task1, task2})

	// Cancel before running
	runner.Cancel()

	err := runner.Run()
	if err == nil {
		t.Error("expected error from cancellation")
	}

	if !runner.IsCancelled() {
		t.Error("expected runner to be cancelled")
	}

	// Second task should not have been executed
	if task2Executed {
		t.Error("task-2 should not have been executed after cancellation")
	}
}

func TestTaskRunnerValidationFailure(t *testing.T) {
	ctx := NewInstallContext()
	bus := NewEventBus()
	runner := NewTaskRunner(ctx, bus)

	task := NewMockTask("invalid-task", "mock")
	task.ValidateFunc = func() error {
		return errors.New("validation error")
	}

	runner.AddTask(task)

	err := runner.Run()
	if err == nil {
		t.Fatal("expected validation error")
	}

	results := runner.Results()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].State != TaskFailed {
		t.Errorf("expected TaskFailed, got %s", results[0].State)
	}
}

func TestTaskRunnerProgress(t *testing.T) {
	ctx := NewInstallContext()
	bus := NewEventBus()
	runner := NewTaskRunner(ctx, bus)

	for i := 0; i < 4; i++ {
		runner.AddTask(NewMockTask("task", "mock"))
	}

	err := runner.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After completion, progress should be 1.0
	if runner.Progress() != 1.0 {
		t.Errorf("expected progress 1.0, got %f", runner.Progress())
	}
}

func TestTaskRunnerManualRollback(t *testing.T) {
	ctx := NewInstallContext()
	bus := NewEventBus()
	runner := NewTaskRunner(ctx, bus)

	rollbackCalled := false
	task := NewMockTask("task-1", "mock")
	task.rollbackable = true
	task.RollbackFunc = func(ctx *InstallContext, bus *EventBus) error {
		rollbackCalled = true
		return nil
	}

	runner.AddTask(task)

	err := runner.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = runner.Rollback()
	if err != nil {
		t.Fatalf("rollback failed: %v", err)
	}

	if !rollbackCalled {
		t.Error("expected rollback to be called")
	}
}

func TestTaskRunnerConcurrentRunPrevention(t *testing.T) {
	ctx := NewInstallContext()
	bus := NewEventBus()
	runner := NewTaskRunner(ctx, bus)

	started := make(chan struct{})
	done := make(chan struct{})

	task := NewMockTask("slow-task", "mock")
	task.ExecuteFunc = func(ctx *InstallContext, bus *EventBus) error {
		close(started)
		<-done
		return nil
	}

	runner.AddTask(task)

	go func() {
		_ = runner.Run()
	}()

	<-started

	// Try to run again while already running
	err := runner.Run()
	if err == nil {
		t.Error("expected error when running concurrently")
	}

	close(done)
}

func TestTaskRunnerEventPublishing(t *testing.T) {
	ctx := NewInstallContext()
	bus := NewEventBus()
	runner := NewTaskRunner(ctx, bus)

	events := make([]Event, 0)
	var mu sync.Mutex

	bus.Subscribe(EventTaskStart, func(e Event) {
		mu.Lock()
		events = append(events, e)
		mu.Unlock()
	})
	bus.Subscribe(EventTaskComplete, func(e Event) {
		mu.Lock()
		events = append(events, e)
		mu.Unlock()
	})

	task := NewMockTask("test-task", "mock")
	runner.AddTask(task)

	err := runner.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Allow events to propagate
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(events) < 2 {
		t.Fatalf("expected at least 2 events, got %d", len(events))
	}

	hasStart := false
	hasComplete := false
	for _, e := range events {
		if e.Type == EventTaskStart {
			hasStart = true
		}
		if e.Type == EventTaskComplete {
			hasComplete = true
		}
	}

	if !hasStart {
		t.Error("expected EventTaskStart")
	}
	if !hasComplete {
		t.Error("expected EventTaskComplete")
	}
}
