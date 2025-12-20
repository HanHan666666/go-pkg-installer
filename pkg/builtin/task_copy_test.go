package builtin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/HanHan666666/go-pkg-installer/pkg/core"
)

func TestRegisterCopyTask(t *testing.T) {
	RegisterCopyTask()

	if !core.Tasks.Has("copy") {
		t.Error("expected copy task to be registered")
	}
}

func TestCopyTaskValidate(t *testing.T) {
	tests := []struct {
		name    string
		task    *CopyTask
		wantErr bool
	}{
		{
			name:    "valid",
			task:    &CopyTask{Source: "/tmp/src", Destination: "/tmp/dst"},
			wantErr: false,
		},
		{
			name:    "missing source",
			task:    &CopyTask{Destination: "/tmp/dst"},
			wantErr: true,
		},
		{
			name:    "missing destination",
			task:    &CopyTask{Source: "/tmp/src"},
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

func TestCopyTaskExecuteFile(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "source.txt")
	dstFile := filepath.Join(tmpDir, "dest.txt")

	content := "test content"
	os.WriteFile(srcFile, []byte(content), 0644)

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &CopyTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-copy",
			TaskType: "copy",
		},
		Source:      srcFile,
		Destination: dstFile,
		Mode:        0644,
	}

	err := task.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify file was copied
	data, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}
	if string(data) != content {
		t.Errorf("expected %q, got %q", content, string(data))
	}
}

func TestCopyTaskExecuteDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "source")
	dstDir := filepath.Join(tmpDir, "dest")

	// Create source directory structure
	os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content1"), 0644)
	os.WriteFile(filepath.Join(srcDir, "subdir", "file2.txt"), []byte("content2"), 0644)

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &CopyTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-copy",
			TaskType: "copy",
		},
		Source:      srcDir,
		Destination: dstDir,
		Mode:        0644,
	}

	err := task.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify files were copied
	data1, err := os.ReadFile(filepath.Join(dstDir, "file1.txt"))
	if err != nil {
		t.Fatalf("failed to read copied file1: %v", err)
	}
	if string(data1) != "content1" {
		t.Errorf("expected 'content1', got %q", string(data1))
	}

	data2, err := os.ReadFile(filepath.Join(dstDir, "subdir", "file2.txt"))
	if err != nil {
		t.Fatalf("failed to read copied file2: %v", err)
	}
	if string(data2) != "content2" {
		t.Errorf("expected 'content2', got %q", string(data2))
	}
}

func TestCopyTaskOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "source.txt")
	dstFile := filepath.Join(tmpDir, "dest.txt")

	os.WriteFile(srcFile, []byte("new content"), 0644)
	os.WriteFile(dstFile, []byte("old content"), 0644)

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	t.Run("without overwrite", func(t *testing.T) {
		task := &CopyTask{
			Source:      srcFile,
			Destination: dstFile,
			Overwrite:   false,
		}

		err := task.Execute(ctx, bus)
		if err == nil {
			t.Error("expected error when destination exists and overwrite is false")
		}
	})

	t.Run("with overwrite", func(t *testing.T) {
		task := &CopyTask{
			Source:      srcFile,
			Destination: dstFile,
			Overwrite:   true,
			Mode:        0644,
		}

		err := task.Execute(ctx, bus)
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		data, _ := os.ReadFile(dstFile)
		if string(data) != "new content" {
			t.Errorf("expected 'new content', got %q", string(data))
		}
	})
}

func TestCopyTaskRollback(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "source.txt")
	dstFile := filepath.Join(tmpDir, "dest.txt")

	os.WriteFile(srcFile, []byte("content"), 0644)

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &CopyTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-copy",
			TaskType: "copy",
		},
		Source:      srcFile,
		Destination: dstFile,
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
	if _, err := os.Stat(dstFile); !os.IsNotExist(err) {
		t.Error("copied file should have been removed during rollback")
	}
}

func TestCopyTaskSourceNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &CopyTask{
		Source:      filepath.Join(tmpDir, "nonexistent"),
		Destination: filepath.Join(tmpDir, "dest"),
	}

	err := task.Execute(ctx, bus)
	if err == nil {
		t.Error("expected error for non-existent source")
	}
}
