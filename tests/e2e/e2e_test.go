// Package e2e provides end-to-end integration tests for the installer framework.
package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/HanHan666666/go-pkg-installer/pkg/builtin"
	"github.com/HanHan666666/go-pkg-installer/pkg/core"
	"github.com/HanHan666666/go-pkg-installer/pkg/schema"
)

func init() {
	builtin.RegisterAll()
	core.RegisterBuiltinGuards()
}

func TestFullInstallFlow(t *testing.T) {
	tmpDir := t.TempDir()
	installDir := filepath.Join(tmpDir, "app")

	ctx := core.NewInstallContext()
	ctx.Set("install_dir", installDir)

	eventBus := core.NewEventBus()

	var taskEvents []string
	eventBus.Subscribe(core.EventTaskStart, func(e core.Event) {
		if p := e.TaskPayload(); p != nil {
			taskEvents = append(taskEvents, "start:"+p.TaskID)
		}
	})
	eventBus.Subscribe(core.EventTaskComplete, func(e core.Event) {
		if p := e.TaskPayload(); p != nil {
			taskEvents = append(taskEvents, "complete:"+p.TaskID)
		}
	})

	workflow := core.NewWorkflow(ctx, eventBus)

	flow := &core.Flow{
		ID:    "install",
		Entry: "welcome",
		Steps: []*core.Step{
			{ID: "welcome", Title: "Welcome"},
			{ID: "progress", Title: "Installing"},
			{ID: "finish", Title: "Complete"},
		},
	}

	workflow.AddFlow(flow)
	workflow.SelectFlow("install")

	if workflow.CurrentStepID() != "welcome" {
		t.Errorf("Expected initial step 'welcome', got '%s'", workflow.CurrentStepID())
	}

	workflow.Next()

	if workflow.CurrentStepID() != "progress" {
		t.Errorf("Expected step 'progress', got '%s'", workflow.CurrentStepID())
	}

	runner := core.NewTaskRunner(ctx, eventBus)

	task1 := core.TaskConfig{
		Type: "shell",
		Params: map[string]any{
			"command": "mkdir",
			"args":    []any{"-p", installDir},
		},
	}
	if err := runner.QueueConfig(task1); err != nil {
		t.Fatalf("Failed to queue task1: %v", err)
	}

	task2 := core.TaskConfig{
		Type: "shell",
		Params: map[string]any{
			"command": "bash",
			"args":    []any{"-c", "echo test > " + filepath.Join(installDir, "test.txt")},
		},
	}
	if err := runner.QueueConfig(task2); err != nil {
		t.Fatalf("Failed to queue task2: %v", err)
	}

	if err := runner.Run(); err != nil {
		t.Fatalf("Task execution failed: %v", err)
	}
	t.Logf("Task events captured: %v", taskEvents)

	workflow.Next()

	testFile := filepath.Join(installDir, "test.txt")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("test.txt was not created")
	}

	if len(taskEvents) == 0 {
		t.Error("Expected task events to be emitted")
	}

	workflow.Complete()
	if !workflow.IsComplete() {
		t.Error("Expected workflow to be complete")
	}
}

func TestUninstallFlow(t *testing.T) {
	tmpDir := t.TempDir()
	installDir := filepath.Join(tmpDir, "app")

	if err := os.MkdirAll(installDir, 0755); err != nil {
		t.Fatalf("Failed to create install dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(installDir, "config.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	ctx := core.NewInstallContext()
	ctx.Set("install_dir", installDir)

	eventBus := core.NewEventBus()
	workflow := core.NewWorkflow(ctx, eventBus)

	flow := &core.Flow{
		ID:    "uninstall",
		Entry: "confirm",
		Steps: []*core.Step{
			{ID: "confirm", Title: "Confirm"},
			{ID: "progress", Title: "Uninstalling"},
			{ID: "finish", Title: "Complete"},
		},
	}

	workflow.AddFlow(flow)
	workflow.SelectFlow("uninstall")
	workflow.Next()

	runner := core.NewTaskRunner(ctx, eventBus)
	task := core.TaskConfig{
		Type: "removePath",
		Params: map[string]any{
			"path":      installDir,
			"recursive": true,
		},
	}
	runner.QueueConfig(task)

	if err := runner.Run(); err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	if _, err := os.Stat(installDir); !os.IsNotExist(err) {
		t.Error("Install directory should be removed after uninstall")
	}
}

func TestWorkflowWithGuards(t *testing.T) {
	ctx := core.NewInstallContext()
	eventBus := core.NewEventBus()
	workflow := core.NewWorkflow(ctx, eventBus)

	flow := &core.Flow{
		ID:    "install",
		Entry: "license",
		Steps: []*core.Step{
			{
				ID:    "license",
				Title: "License",
				GuardsCfg: []map[string]any{
					{
						"type":  "mustAccept",
						"field": "license_accepted",
					},
				},
			},
			{ID: "finish", Title: "Complete"},
		},
	}

	workflow.AddFlow(flow)
	workflow.SelectFlow("install")

	_, err := workflow.Next()
	if err == nil {
		t.Error("Expected guard to prevent navigation")
	}

	ctx.Set("license_accepted", true)

	_, err = workflow.Next()
	if err != nil {
		t.Errorf("Navigation should succeed after accepting: %v", err)
	}
}

func TestMultipleFlows(t *testing.T) {
	configContent := `
product:
  name: "Multi-Flow App"

flows:
  install:
    entry: welcome
    steps:
      - id: welcome
        title: "Welcome"
        screen:
          type: welcome
          content: "Install flow"
      - id: finish
        title: "Done"
        screen:
          type: finish
  uninstall:
    entry: confirm
    steps:
      - id: confirm
        title: "Confirm"
        screen:
          type: summary
      - id: done
        title: "Done"
        screen:
          type: finish
`

	cfg, err := schema.LoadConfig([]byte(configContent))
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	ctx := core.NewInstallContext()
	eventBus := core.NewEventBus()
	workflow := core.NewWorkflow(ctx, eventBus)

	for flowName, flowCfg := range cfg.Flows {
		flow := &core.Flow{
			ID:    flowName,
			Entry: flowCfg.Entry,
		}
		for _, stepCfg := range flowCfg.Steps {
			step := &core.Step{
				ID:     stepCfg.ID,
				Title:  stepCfg.Title,
				Config: stepCfg,
			}
			flow.Steps = append(flow.Steps, step)
		}
		workflow.AddFlow(flow)
	}

	workflow.SelectFlow("install")
	if workflow.CurrentStepID() != "welcome" {
		t.Errorf("Expected 'welcome', got '%s'", workflow.CurrentStepID())
	}

	workflow.SelectFlow("uninstall")
	if workflow.CurrentStepID() != "confirm" {
		t.Errorf("Expected 'confirm', got '%s'", workflow.CurrentStepID())
	}
}

func TestContextVariableSubstitution(t *testing.T) {
	tmpDir := t.TempDir()

	ctx := core.NewInstallContext()
	ctx.Set("install_dir", tmpDir)
	ctx.Set("app_name", "TestApp")
	ctx.Set("version", "2.0.0")

	eventBus := core.NewEventBus()
	runner := core.NewTaskRunner(ctx, eventBus)

	taskCfg := core.TaskConfig{
		Type: "writeConfig",
		Params: map[string]any{
			"destination": filepath.Join(tmpDir, "info.txt"),
			"content":     "App: ${app_name}\nVersion: ${version}",
			"format":      "text",
		},
	}

	if err := runner.QueueConfig(taskCfg); err != nil {
		t.Fatalf("Failed to queue task: %v", err)
	}

	if err := runner.Run(); err != nil {
		t.Fatalf("Task failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "info.txt"))
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	expected := "App: TestApp\nVersion: 2.0.0"
	if string(content) != expected {
		t.Errorf("Expected content:\n%s\nGot:\n%s", expected, string(content))
	}
}

func TestEventBusIntegration(t *testing.T) {
	ctx := core.NewInstallContext()
	eventBus := core.NewEventBus()

	var events []core.EventType
	eventBus.SubscribeAll(func(e core.Event) {
		events = append(events, e.Type)
	})

	workflow := core.NewWorkflow(ctx, eventBus)

	flow := &core.Flow{
		ID:    "test",
		Entry: "step1",
		Steps: []*core.Step{
			{ID: "step1", Title: "Step 1"},
			{ID: "step2", Title: "Step 2"},
		},
	}

	workflow.AddFlow(flow)
	workflow.SelectFlow("test")
	workflow.Next()
	workflow.Complete()

	hasStepChange := false
	hasFlowComplete := false
	for _, e := range events {
		if e == core.EventStepChange {
			hasStepChange = true
		}
		if e == core.EventFlowComplete {
			hasFlowComplete = true
		}
	}

	if !hasStepChange {
		t.Error("Expected step change event")
	}
	if !hasFlowComplete {
		t.Error("Expected flow complete event")
	}
}

func TestHeadlessInstallation(t *testing.T) {
	tmpDir := t.TempDir()
	installDir := filepath.Join(tmpDir, "myapp")

	ctx := core.NewInstallContext()
	ctx.Set("install_dir", installDir)
	ctx.Set("app_name", "HeadlessApp")

	eventBus := core.NewEventBus()
	workflow := core.NewWorkflow(ctx, eventBus)

	flow := &core.Flow{
		ID:    "install",
		Entry: "start",
		Steps: []*core.Step{
			{ID: "start", Title: "Start"},
			{ID: "install", Title: "Install"},
			{ID: "done", Title: "Done"},
		},
	}

	workflow.AddFlow(flow)
	workflow.SelectFlow("install")

	runner := core.NewTaskRunner(ctx, eventBus)

	for !workflow.IsComplete() {
		step := workflow.CurrentStep()
		t.Logf("Step: %s", step.Title)

		if step.ID == "install" {
			task1 := core.TaskConfig{
				Type: "shell",
				Params: map[string]any{
					"command": "mkdir",
					"args":    []any{"-p", installDir},
				},
			}
			runner.QueueConfig(task1)

			task2 := core.TaskConfig{
				Type: "writeConfig",
				Params: map[string]any{
					"destination": filepath.Join(installDir, "manifest.json"),
					"format":      "json",
					"content": map[string]any{
						"name":    "${app_name}",
						"version": "1.0.0",
					},
				},
			}
			runner.QueueConfig(task2)

			if err := runner.Run(); err != nil {
				t.Fatalf("Task execution failed: %v", err)
			}
		}

		if !workflow.IsLastStep() {
			workflow.Next()
		} else {
			workflow.Complete()
		}
	}

	manifestPath := filepath.Join(installDir, "manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("manifest.json was not created")
	}
}
