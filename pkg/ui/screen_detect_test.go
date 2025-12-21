package ui

import (
	"errors"
	"testing"
	"time"

	"github.com/HanHan666666/go-pkg-installer/pkg/core"
)

type testDetectTask struct {
	id     string
	typ    string
	execFn func(ctx *core.InstallContext) error
}

func (t *testDetectTask) ID() string {
	return t.id
}

func (t *testDetectTask) Type() string {
	return t.typ
}

func (t *testDetectTask) Validate() error {
	return nil
}

func (t *testDetectTask) Execute(ctx *core.InstallContext, _ *core.EventBus) error {
	if t.execFn == nil {
		return nil
	}
	return t.execFn(ctx)
}

func (t *testDetectTask) Rollback(ctx *core.InstallContext, _ *core.EventBus) error {
	return nil
}

func (t *testDetectTask) CanRollback() bool {
	return false
}

func registerDetectTask(t *testing.T, name string, execFn func(ctx *core.InstallContext) error) {
	t.Helper()
	if core.Tasks.Has(name) {
		return
	}
	core.Tasks.MustRegister(name, func(params map[string]any, ctx *core.InstallContext) (core.Task, error) {
		return &testDetectTask{
			id:     name,
			typ:    name,
			execFn: execFn,
		}, nil
	})
}

func waitForDetectComplete(t *testing.T, screen *DetectScreen) {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		screen.mu.Lock()
		done := screen.isComplete
		screen.mu.Unlock()
		if done {
			return
		}
		select {
		case <-deadline:
			t.Fatal("timeout waiting for detect completion")
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func TestDetectScreenSuccessUpdatesContent(t *testing.T) {
	taskName := "detect_test_success"
	registerDetectTask(t, taskName, func(ctx *core.InstallContext) error {
		ctx.Set("meta.result", "ok")
		return nil
	})

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()
	step := &core.StepConfig{
		ID: "detect",
		Screen: &core.ScreenConfig{
			Type:        "detect",
			Description: "Detecting environment...",
			Content:     "Detected: ${meta.result}",
		},
		Tasks: []core.TaskConfig{
			{Type: taskName},
		},
	}

	screen := NewDetectScreen(step).(*DetectScreen)
	screen.ctx = ctx
	screen.bus = bus
	screen.active = true

	screen.runTasks()

	if !screen.isComplete || screen.failed {
		t.Fatalf("expected success, complete=%v failed=%v", screen.isComplete, screen.failed)
	}
	if failed, _ := ctx.Get("step.failed"); failed == true {
		t.Fatalf("expected step.failed=false, got %v", failed)
	}
	if got := screen.contentText; got != "Detected: ok" {
		t.Fatalf("expected rendered content, got %q", got)
	}
	if err := screen.Validate(); err != nil {
		t.Fatalf("expected validate ok, got %v", err)
	}
}

func TestDetectScreenFailureBlocksContinue(t *testing.T) {
	taskName := "detect_test_failure"
	registerDetectTask(t, taskName, func(ctx *core.InstallContext) error {
		return errors.New("boom")
	})

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()
	step := &core.StepConfig{
		ID: "detect",
		Screen: &core.ScreenConfig{
			Type:        "detect",
			Description: "Detecting...",
			Content:     "Detected",
		},
		Tasks: []core.TaskConfig{
			{Type: taskName},
		},
	}

	screen := NewDetectScreen(step).(*DetectScreen)
	screen.ctx = ctx
	screen.bus = bus
	screen.active = true

	screen.runTasks()

	if !screen.isComplete || !screen.failed {
		t.Fatalf("expected failure, complete=%v failed=%v", screen.isComplete, screen.failed)
	}
	if got := screen.contentText; got == "" {
		t.Fatal("expected failure message content")
	}
	if err := screen.Validate(); err == nil || err.Error() != "boom" {
		t.Fatalf("expected task error, got %v", err)
	}
	if failed, _ := ctx.Get("step.failed"); failed != true {
		t.Fatalf("expected step.failed=true, got %v", failed)
	}
	if failedID, _ := ctx.Get("step.failed_id"); failedID != "detect" {
		t.Fatalf("expected step.failed_id=detect, got %v", failedID)
	}
}

func TestDetectScreenValidateBlocksUntilComplete(t *testing.T) {
	taskName := "detect_test_blocking"
	started := make(chan struct{})
	release := make(chan struct{})

	registerDetectTask(t, taskName, func(ctx *core.InstallContext) error {
		close(started)
		<-release
		return nil
	})

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()
	step := &core.StepConfig{
		ID: "detect",
		Screen: &core.ScreenConfig{
			Type:        "detect",
			Description: "Detecting...",
			Content:     "Detected",
		},
		Tasks: []core.TaskConfig{
			{Type: taskName},
		},
	}

	screen := NewDetectScreen(step).(*DetectScreen)
	screen.ctx = ctx
	screen.bus = bus
	screen.active = true

	go screen.runTasks()

	select {
	case <-started:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("task did not start")
	}

	if err := screen.Validate(); err == nil {
		t.Fatal("expected validation error while detecting")
	}

	close(release)
	waitForDetectComplete(t, screen)

	if err := screen.Validate(); err != nil {
		t.Fatalf("expected validate ok after completion, got %v", err)
	}
}
