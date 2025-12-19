package builtin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/anthropics/go-pkg-installer/pkg/core"
	"gopkg.in/yaml.v3"
)

func TestRegisterWriteConfigTask(t *testing.T) {
	RegisterWriteConfigTask()

	if !core.Tasks.Has("writeConfig") {
		t.Error("expected writeConfig task to be registered")
	}
}

func TestWriteConfigTaskValidate(t *testing.T) {
	tests := []struct {
		name    string
		task    *WriteConfigTask
		wantErr bool
	}{
		{
			name:    "valid",
			task:    &WriteConfigTask{Destination: "/tmp/config.json", Content: map[string]any{"key": "value"}},
			wantErr: false,
		},
		{
			name:    "missing destination",
			task:    &WriteConfigTask{Content: map[string]any{"key": "value"}},
			wantErr: true,
		},
		{
			name:    "missing content",
			task:    &WriteConfigTask{Destination: "/tmp/config.json"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.task.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestWriteConfigTaskJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &WriteConfigTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-writeconfig",
			TaskType: "writeConfig",
		},
		Destination: configFile,
		Format:      "json",
		Content: map[string]any{
			"key1": "value1",
			"key2": 42,
		},
		Mode: 0644,
	}

	err := task.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify file was written
	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if result["key1"] != "value1" {
		t.Errorf("expected key1=value1, got %v", result["key1"])
	}
}

func TestWriteConfigTaskYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &WriteConfigTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-writeconfig",
			TaskType: "writeConfig",
		},
		Destination: configFile,
		Format:      "yaml",
		Content: map[string]any{
			"key1": "value1",
			"nested": map[string]any{
				"key2": "value2",
			},
		},
		Mode: 0644,
	}

	err := task.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify file was written
	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	var result map[string]any
	if err := yaml.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse YAML: %v", err)
	}

	if result["key1"] != "value1" {
		t.Errorf("expected key1=value1, got %v", result["key1"])
	}
}

func TestWriteConfigTaskText(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.txt")

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &WriteConfigTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-writeconfig",
			TaskType: "writeConfig",
		},
		Destination: configFile,
		Format:      "text",
		Content:     "Hello, World!",
		Mode:        0644,
	}

	err := task.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	if string(data) != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got %q", string(data))
	}
}

func TestWriteConfigTaskAutoDetectFormat(t *testing.T) {
	RegisterWriteConfigTask()

	ctx := core.NewInstallContext()

	factory, _ := core.Tasks.Get("writeConfig")

	tests := []struct {
		path           string
		expectedFormat string
	}{
		{"config.json", "json"},
		{"config.yaml", "yaml"},
		{"config.yml", "yaml"},
		{"config.txt", "text"},
		{"config", "text"},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			task, err := factory(map[string]any{
				"destination": tc.path,
				"content":     "test",
			}, ctx)
			if err != nil {
				t.Fatalf("factory error = %v", err)
			}

			wcTask := task.(*WriteConfigTask)
			if wcTask.Format != tc.expectedFormat {
				t.Errorf("expected format %q, got %q", tc.expectedFormat, wcTask.Format)
			}
		})
	}
}

func TestWriteConfigTaskWithTemplates(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	ctx := core.NewInstallContext()
	ctx.Set("appName", "MyApp")
	ctx.Set("version", "1.0.0")
	bus := core.NewEventBus()

	task := &WriteConfigTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-writeconfig",
			TaskType: "writeConfig",
		},
		Destination: configFile,
		Format:      "json",
		Content: map[string]any{
			"name":    "{{.appName}}",
			"version": "{{.version}}",
		},
		Mode: 0644,
	}

	err := task.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	var result map[string]any
	json.Unmarshal(data, &result)

	if result["name"] != "MyApp" {
		t.Errorf("expected name=MyApp, got %v", result["name"])
	}
	if result["version"] != "1.0.0" {
		t.Errorf("expected version=1.0.0, got %v", result["version"])
	}
}

func TestWriteConfigTaskRollback(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &WriteConfigTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-writeconfig",
			TaskType: "writeConfig",
		},
		Destination: configFile,
		Format:      "json",
		Content:     map[string]any{"key": "value"},
		Mode:        0644,
	}

	// Execute
	if err := task.Execute(ctx, bus); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify can rollback
	if !task.CanRollback() {
		t.Error("expected CanRollback to return true")
	}

	// Rollback
	if err := task.Rollback(ctx, bus); err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}

	// File should be removed
	if _, err := os.Stat(configFile); !os.IsNotExist(err) {
		t.Error("config file should have been removed during rollback")
	}
}

func TestWriteConfigTaskRollbackRestoresOld(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	// Create original file
	os.WriteFile(configFile, []byte(`{"original": true}`), 0644)

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &WriteConfigTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-writeconfig",
			TaskType: "writeConfig",
		},
		Destination: configFile,
		Format:      "json",
		Content:     map[string]any{"new": true},
		Mode:        0644,
	}

	// Execute (overwrites original)
	if err := task.Execute(ctx, bus); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Rollback
	if err := task.Rollback(ctx, bus); err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}

	// Original content should be restored
	data, _ := os.ReadFile(configFile)
	if string(data) != `{"original": true}` {
		t.Errorf("expected original content to be restored, got %q", string(data))
	}
}
