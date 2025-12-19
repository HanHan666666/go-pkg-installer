package builtin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/anthropics/go-pkg-installer/pkg/core"
)

func TestRegisterRemovePathTask(t *testing.T) {
	RegisterRemovePathTask()

	if !core.Tasks.Has("removePath") {
		t.Error("expected removePath task to be registered")
	}
}

func TestRemovePathTaskValidate(t *testing.T) {
	tests := []struct {
		name    string
		task    *RemovePathTask
		wantErr bool
	}{
		{
			name:    "valid",
			task:    &RemovePathTask{Path: "/tmp/file.txt"},
			wantErr: false,
		},
		{
			name:    "missing path",
			task:    &RemovePathTask{},
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

func TestRemovePathTaskExecuteFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("content"), 0644)

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &RemovePathTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-remove",
			TaskType: "removePath",
		},
		Path: testFile,
	}

	err := task.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify file was removed
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("file should have been removed")
	}
}

func TestRemovePathTaskExecuteEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "empty")
	os.MkdirAll(testDir, 0755)

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &RemovePathTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-remove",
			TaskType: "removePath",
		},
		Path:      testDir,
		Recursive: false,
	}

	err := task.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify directory was removed
	if _, err := os.Stat(testDir); !os.IsNotExist(err) {
		t.Error("directory should have been removed")
	}
}

func TestRemovePathTaskNonEmptyDirNoRecursive(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "notempty")
	os.MkdirAll(testDir, 0755)
	os.WriteFile(filepath.Join(testDir, "file.txt"), []byte("content"), 0644)

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &RemovePathTask{
		Path:      testDir,
		Recursive: false,
	}

	err := task.Execute(ctx, bus)
	if err == nil {
		t.Error("expected error for non-empty directory without recursive")
	}
}

func TestRemovePathTaskRecursive(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "recursive")
	os.MkdirAll(filepath.Join(testDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(testDir, "file1.txt"), []byte("content1"), 0644)
	os.WriteFile(filepath.Join(testDir, "subdir", "file2.txt"), []byte("content2"), 0644)

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &RemovePathTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-remove",
			TaskType: "removePath",
		},
		Path:      testDir,
		Recursive: true,
	}

	err := task.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify directory was removed
	if _, err := os.Stat(testDir); !os.IsNotExist(err) {
		t.Error("directory should have been removed recursively")
	}
}

func TestRemovePathTaskNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistent := filepath.Join(tmpDir, "nonexistent")

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	t.Run("without force", func(t *testing.T) {
		task := &RemovePathTask{
			Path:  nonExistent,
			Force: false,
		}

		err := task.Execute(ctx, bus)
		if err == nil {
			t.Error("expected error for non-existent path without force")
		}
	})

	t.Run("with force", func(t *testing.T) {
		task := &RemovePathTask{
			Path:  nonExistent,
			Force: true,
		}

		err := task.Execute(ctx, bus)
		if err != nil {
			t.Errorf("expected no error with force flag, got %v", err)
		}
	})
}

func TestRemovePathTaskNoRollback(t *testing.T) {
	task := &RemovePathTask{
		Path: "/tmp/file",
	}

	if task.CanRollback() {
		t.Error("expected CanRollback to return false for remove operations")
	}
}

func TestRemovePathTaskRemoveSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "target.txt")
	link := filepath.Join(tmpDir, "link.txt")

	os.WriteFile(target, []byte("content"), 0644)
	os.Symlink(target, link)

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &RemovePathTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-remove",
			TaskType: "removePath",
		},
		Path: link,
	}

	err := task.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Symlink should be removed but target should remain
	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Error("symlink should have been removed")
	}
	if _, err := os.Stat(target); err != nil {
		t.Error("target should still exist")
	}
}

func TestRemovePathTaskFactory(t *testing.T) {
	RegisterRemovePathTask()

	ctx := core.NewInstallContext()
	ctx.Set("installDir", "/opt/myapp")

	factory, ok := core.Tasks.Get("removePath")
	if !ok {
		t.Fatal("removePath factory not registered")
	}

	task, err := factory(map[string]any{
		"id":        "rm-1",
		"path":      "{{.installDir}}/cache",
		"recursive": true,
		"force":     true,
	}, ctx)

	if err != nil {
		t.Fatalf("factory error = %v", err)
	}

	rmTask, ok := task.(*RemovePathTask)
	if !ok {
		t.Fatal("expected *RemovePathTask")
	}

	if rmTask.Path != "/opt/myapp/cache" {
		t.Errorf("expected path to be rendered, got %q", rmTask.Path)
	}
	if !rmTask.Recursive {
		t.Error("expected recursive to be true")
	}
	if !rmTask.Force {
		t.Error("expected force to be true")
	}
}
