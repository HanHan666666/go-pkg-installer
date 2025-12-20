package builtin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/HanHan666666/go-pkg-installer/pkg/core"
)

func TestRegisterShellTask(t *testing.T) {
	RegisterShellTask()

	if !core.Tasks.Has("shell") {
		t.Error("expected shell task to be registered")
	}
}

func TestShellTaskValidate(t *testing.T) {
	tests := []struct {
		name    string
		task    *ShellTask
		wantErr bool
	}{
		{
			name:    "valid",
			task:    &ShellTask{Command: "echo hello"},
			wantErr: false,
		},
		{
			name:    "missing command",
			task:    &ShellTask{},
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

func TestShellTaskExecute(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "output.txt")

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &ShellTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-shell",
			TaskType: "shell",
		},
		Command: "echo 'hello world' > " + outputFile,
	}

	err := task.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify command was executed
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if string(data) != "hello world\n" {
		t.Errorf("expected 'hello world\\n', got %q", string(data))
	}
}

func TestShellTaskWithArgs(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "output.txt")

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &ShellTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-shell",
			TaskType: "shell",
		},
		Command: "touch",
		Args:    []string{outputFile},
	}

	err := task.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputFile); err != nil {
		t.Errorf("expected file to be created: %v", err)
	}
}

func TestShellTaskWithEnv(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "output.txt")

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &ShellTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-shell",
			TaskType: "shell",
		},
		Command: "echo $TEST_VAR > " + outputFile,
		Env: map[string]string{
			"TEST_VAR": "test_value",
		},
	}

	err := task.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if string(data) != "test_value\n" {
		t.Errorf("expected 'test_value\\n', got %q", string(data))
	}
}

func TestShellTaskWithWorkDir(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	os.MkdirAll(subDir, 0755)

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &ShellTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-shell",
			TaskType: "shell",
		},
		Command: "pwd > output.txt",
		WorkDir: subDir,
	}

	err := task.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(subDir, "output.txt"))
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if string(data) != subDir+"\n" {
		t.Errorf("expected %q, got %q", subDir+"\n", string(data))
	}
}

func TestShellTaskFailure(t *testing.T) {
	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &ShellTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-shell",
			TaskType: "shell",
		},
		Command: "exit 1",
	}

	err := task.Execute(ctx, bus)
	if err == nil {
		t.Error("expected error for failed command")
	}
}

func TestShellTaskRollback(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &ShellTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-shell",
			TaskType: "shell",
		},
		Command:     "touch " + testFile,
		RollbackCmd: "rm " + testFile,
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
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("file should have been removed during rollback")
	}
}

func TestShellTaskNoRollback(t *testing.T) {
	task := &ShellTask{
		Command: "echo test",
	}

	if task.CanRollback() {
		t.Error("expected CanRollback to return false when no rollback command")
	}
}

func TestShellTaskFactory(t *testing.T) {
	RegisterShellTask()

	ctx := core.NewInstallContext()
	ctx.Set("dir", "/tmp")

	factory, ok := core.Tasks.Get("shell")
	if !ok {
		t.Fatal("shell factory not registered")
	}

	task, err := factory(map[string]any{
		"id":      "shell-1",
		"command": "ls {{.dir}}",
		"env": map[string]any{
			"PATH": "/usr/bin",
		},
	}, ctx)

	if err != nil {
		t.Fatalf("factory error = %v", err)
	}

	shellTask, ok := task.(*ShellTask)
	if !ok {
		t.Fatal("expected *ShellTask")
	}

	if shellTask.Command != "ls /tmp" {
		t.Errorf("expected command to be rendered, got %q", shellTask.Command)
	}
}
