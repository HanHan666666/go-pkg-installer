package core

import (
	"errors"
	"testing"
)

func createTestFlow() *Flow {
	return &Flow{
		ID:    "install",
		Entry: "welcome",
		Steps: []*Step{
			{ID: "welcome", Title: "Welcome"},
			{ID: "license", Title: "License"},
			{ID: "destination", Title: "Destination"},
			{ID: "install", Title: "Install"},
			{ID: "finish", Title: "Finish"},
		},
	}
}

func TestNewWorkflow(t *testing.T) {
	ctx := NewInstallContext()
	bus := NewEventBus()
	w := NewWorkflow(ctx, bus)

	if w == nil {
		t.Fatal("NewWorkflow returned nil")
	}
	if w.ctx != ctx {
		t.Error("Context not set correctly")
	}
	if w.bus != bus {
		t.Error("EventBus not set correctly")
	}
}

func TestWorkflowAddFlow(t *testing.T) {
	w := NewWorkflow(NewInstallContext(), NewEventBus())

	flow := createTestFlow()
	err := w.AddFlow(flow)
	if err != nil {
		t.Errorf("AddFlow should not error: %v", err)
	}

	// Duplicate flow
	err = w.AddFlow(flow)
	if err == nil {
		t.Error("AddFlow should error on duplicate flow")
	}

	// Empty flow
	err = w.AddFlow(&Flow{ID: "empty"})
	if err == nil {
		t.Error("AddFlow should error on empty steps")
	}

	// No ID
	err = w.AddFlow(&Flow{Steps: []*Step{{ID: "test"}}})
	if err == nil {
		t.Error("AddFlow should error on empty ID")
	}
}

func TestWorkflowSelectFlow(t *testing.T) {
	ctx := NewInstallContext()
	w := NewWorkflow(ctx, NewEventBus())
	w.AddFlow(createTestFlow())

	err := w.SelectFlow("install")
	if err != nil {
		t.Errorf("SelectFlow should not error: %v", err)
	}

	if w.CurrentStepID() != "welcome" {
		t.Errorf("Expected 'welcome', got %s", w.CurrentStepID())
	}

	if ctx.Runtime.FlowID != "install" {
		t.Errorf("Context FlowID should be 'install', got %s", ctx.Runtime.FlowID)
	}

	if ctx.Runtime.CurrentStep != "welcome" {
		t.Errorf("Context CurrentStep should be 'welcome', got %s", ctx.Runtime.CurrentStep)
	}

	// Non-existent flow
	err = w.SelectFlow("nonexistent")
	if err == nil {
		t.Error("SelectFlow should error on nonexistent flow")
	}
}

func TestWorkflowLinearNavigation(t *testing.T) {
	ctx := NewInstallContext()
	bus := NewEventBus()
	w := NewWorkflow(ctx, bus)
	w.AddFlow(createTestFlow())
	w.SelectFlow("install")

	// Track step changes
	changes := []StepChangePayload{}
	bus.Subscribe(EventStepChange, func(e Event) {
		changes = append(changes, e.Payload.(StepChangePayload))
	})

	// Navigate forward
	stepID, err := w.Next()
	if err != nil {
		t.Errorf("Next should not error: %v", err)
	}
	if stepID != "license" {
		t.Errorf("Expected 'license', got %s", stepID)
	}

	stepID, err = w.Next()
	if err != nil {
		t.Errorf("Next should not error: %v", err)
	}
	if stepID != "destination" {
		t.Errorf("Expected 'destination', got %s", stepID)
	}

	// Navigate backward
	stepID, err = w.Prev()
	if err != nil {
		t.Errorf("Prev should not error: %v", err)
	}
	if stepID != "license" {
		t.Errorf("Expected 'license', got %s", stepID)
	}

	// Check events
	if len(changes) != 3 {
		t.Errorf("Expected 3 step changes, got %d", len(changes))
	}
}

func TestWorkflowAtBoundaries(t *testing.T) {
	w := NewWorkflow(NewInstallContext(), NewEventBus())
	w.AddFlow(createTestFlow())
	w.SelectFlow("install")

	// At first step
	if !w.IsFirstStep() {
		t.Error("Should be at first step")
	}
	if w.IsLastStep() {
		t.Error("Should not be at last step")
	}

	// Cannot go back from first step
	_, err := w.Prev()
	if err == nil {
		t.Error("Prev should error at first step")
	}

	// Navigate to last step
	for i := 0; i < 4; i++ {
		w.Next()
	}

	if w.IsFirstStep() {
		t.Error("Should not be at first step")
	}
	if !w.IsLastStep() {
		t.Error("Should be at last step")
	}

	// Cannot go forward from last step
	_, err = w.Next()
	if err == nil {
		t.Error("Next should error at last step")
	}
}

func TestWorkflowStepStatus(t *testing.T) {
	w := NewWorkflow(NewInstallContext(), NewEventBus())
	w.AddFlow(createTestFlow())
	w.SelectFlow("install")

	// Initial state
	if w.StepStatus("welcome") != StepCurrent {
		t.Error("welcome should be Current")
	}
	if w.StepStatus("license") != StepNotStarted {
		t.Error("license should be NotStarted")
	}

	w.Next()

	// After navigation
	if w.StepStatus("welcome") != StepCompleted {
		t.Error("welcome should be Completed")
	}
	if w.StepStatus("license") != StepCurrent {
		t.Error("license should be Current")
	}
}

func TestWorkflowDisableStep(t *testing.T) {
	w := NewWorkflow(NewInstallContext(), NewEventBus())
	w.AddFlow(createTestFlow())
	w.SelectFlow("install")

	// Disable license step
	err := w.DisableStep("license")
	if err != nil {
		t.Errorf("DisableStep should not error: %v", err)
	}

	if w.StepStatus("license") != StepDisabled {
		t.Error("license should be Disabled")
	}

	// Next should skip disabled step
	stepID, _ := w.Next()
	if stepID != "destination" {
		t.Errorf("Expected 'destination' (skipping license), got %s", stepID)
	}

	// Cannot disable current step
	err = w.DisableStep("destination")
	if err == nil {
		t.Error("Should not be able to disable current step")
	}
}

func TestWorkflowEnableStep(t *testing.T) {
	w := NewWorkflow(NewInstallContext(), NewEventBus())
	w.AddFlow(createTestFlow())
	w.SelectFlow("install")

	w.DisableStep("license")
	w.EnableStep("license")

	if w.StepStatus("license") != StepNotStarted {
		t.Error("license should be NotStarted after enable")
	}
}

func TestWorkflowJumpTo(t *testing.T) {
	w := NewWorkflow(NewInstallContext(), NewEventBus())
	flow := createTestFlow()
	flow.Steps[3].AllowJump = true // install step allows jump
	w.AddFlow(flow)
	w.SelectFlow("install")

	// Cannot jump to unvisited step (without allowJump)
	err := w.JumpTo("finish")
	if err == nil {
		t.Error("Should not jump to unvisited step")
	}

	// Can jump to step with allowJump
	err = w.JumpTo("install")
	if err != nil {
		t.Errorf("Should jump to step with allowJump: %v", err)
	}

	// Navigate and visit some steps
	w.SelectFlow("install")
	w.Next() // license
	w.Next() // destination

	// Can jump back to visited step
	err = w.JumpTo("license")
	if err != nil {
		t.Errorf("Should jump to visited step: %v", err)
	}
	if w.CurrentStepID() != "license" {
		t.Errorf("Expected 'license', got %s", w.CurrentStepID())
	}

	// Cannot jump to disabled step
	w.DisableStep("destination")
	err = w.JumpTo("destination")
	if err == nil {
		t.Error("Should not jump to disabled step")
	}
}

func TestWorkflowIsVisited(t *testing.T) {
	w := NewWorkflow(NewInstallContext(), NewEventBus())
	w.AddFlow(createTestFlow())
	w.SelectFlow("install")

	if !w.IsVisited("welcome") {
		t.Error("welcome should be visited")
	}
	if w.IsVisited("license") {
		t.Error("license should not be visited yet")
	}

	w.Next()

	if !w.IsVisited("license") {
		t.Error("license should be visited now")
	}
}

func TestWorkflowComplete(t *testing.T) {
	ctx := NewInstallContext()
	bus := NewEventBus()
	w := NewWorkflow(ctx, bus)
	w.AddFlow(createTestFlow())
	w.SelectFlow("install")

	completed := false
	bus.Subscribe(EventFlowComplete, func(e Event) {
		completed = true
	})

	// Navigate to end
	for i := 0; i < 4; i++ {
		w.Next()
	}
	w.Complete()

	if !ctx.Runtime.Completed {
		t.Error("Runtime.Completed should be true")
	}
	if !completed {
		t.Error("FlowComplete event should be fired")
	}
	if w.StepStatus("finish") != StepCompleted {
		t.Error("finish step should be Completed")
	}
}

// Mock guard for testing
type mockGuard struct {
	shouldPass bool
	message    string
}

func (g *mockGuard) Type() string    { return "mock" }
func (g *mockGuard) Message() string { return g.message }
func (g *mockGuard) Check(*InstallContext) error {
	if g.shouldPass {
		return nil
	}
	return errors.New(g.message)
}

func TestWorkflowGuards(t *testing.T) {
	// Register mock guard
	Guards.Clear()
	Guards.Register("mockGuard", func(config map[string]any) (Guard, error) {
		pass := true
		if v, ok := config["pass"].(bool); ok {
			pass = v
		}
		msg := "guard failed"
		if v, ok := config["message"].(string); ok {
			msg = v
		}
		return &mockGuard{shouldPass: pass, message: msg}, nil
	})

	ctx := NewInstallContext()
	w := NewWorkflow(ctx, NewEventBus())

	flow := &Flow{
		ID:    "test",
		Entry: "step1",
		Steps: []*Step{
			{
				ID:    "step1",
				Title: "Step 1",
				GuardsCfg: []map[string]any{
					{"type": "mockGuard", "pass": false, "message": "must accept"},
				},
			},
			{ID: "step2", Title: "Step 2"},
		},
	}
	w.AddFlow(flow)
	w.SelectFlow("test")

	// Guard should block
	err := w.CanGoNext()
	if err == nil {
		t.Error("Guard should block navigation")
	}
	if err.Error() != "must accept" {
		t.Errorf("Expected 'must accept', got %s", err.Error())
	}

	_, err = w.Next()
	if err == nil {
		t.Error("Next should fail when guard blocks")
	}

	// Update guard to pass
	flow.Steps[0].GuardsCfg[0]["pass"] = true

	err = w.CanGoNext()
	if err != nil {
		t.Errorf("Guard should pass now: %v", err)
	}
}

func TestWorkflowBranching(t *testing.T) {
	ctx := NewInstallContext()
	w := NewWorkflow(ctx, NewEventBus())

	flow := &Flow{
		ID:    "test",
		Entry: "start",
		Steps: []*Step{
			{
				ID:    "start",
				Title: "Start",
				Branch: &BranchConfig{
					Condition: "install.type",
					Branches: map[string]string{
						"full":    "full_install",
						"minimal": "minimal_install",
					},
					Default: "full_install",
				},
			},
			{ID: "full_install", Title: "Full Install"},
			{ID: "minimal_install", Title: "Minimal Install"},
			{ID: "finish", Title: "Finish"},
		},
	}
	w.AddFlow(flow)
	w.SelectFlow("test")

	// Set branch condition
	ctx.Set("install.type", "minimal")

	stepID, _ := w.Next()
	if stepID != "minimal_install" {
		t.Errorf("Expected 'minimal_install', got %s", stepID)
	}

	// Test default branch
	w.SelectFlow("test")
	ctx.Set("install.type", "unknown")

	stepID, _ = w.Next()
	if stepID != "full_install" {
		t.Errorf("Expected 'full_install' (default), got %s", stepID)
	}
}

func TestWorkflowExplicitNextPrev(t *testing.T) {
	w := NewWorkflow(NewInstallContext(), NewEventBus())

	flow := &Flow{
		ID:    "test",
		Entry: "a",
		Steps: []*Step{
			{ID: "a", Title: "A", Next: "c"}, // Skip B
			{ID: "b", Title: "B"},
			{ID: "c", Title: "C", Prev: "a"}, // Go back to A
		},
	}
	w.AddFlow(flow)
	w.SelectFlow("test")

	// Next should go to C (explicit next)
	stepID, _ := w.Next()
	if stepID != "c" {
		t.Errorf("Expected 'c' (explicit next), got %s", stepID)
	}

	// Prev should go to A (explicit prev)
	stepID, _ = w.Prev()
	if stepID != "a" {
		t.Errorf("Expected 'a' (explicit prev), got %s", stepID)
	}
}

func TestWorkflowNoFlowSelected(t *testing.T) {
	w := NewWorkflow(NewInstallContext(), NewEventBus())

	step := w.CurrentStep()
	if step != nil {
		t.Error("CurrentStep should return nil when no flow selected")
	}

	stepID := w.CurrentStepID()
	if stepID != "" {
		t.Error("CurrentStepID should return empty when no flow selected")
	}

	_, err := w.Next()
	if err == nil {
		t.Error("Next should error when no flow selected")
	}

	_, err = w.Prev()
	if err == nil {
		t.Error("Prev should error when no flow selected")
	}
}

func TestWorkflowSteps(t *testing.T) {
	w := NewWorkflow(NewInstallContext(), NewEventBus())

	// No flow
	if w.Steps() != nil {
		t.Error("Steps should return nil when no flow selected")
	}

	w.AddFlow(createTestFlow())
	w.SelectFlow("install")

	steps := w.Steps()
	if len(steps) != 5 {
		t.Errorf("Expected 5 steps, got %d", len(steps))
	}
}
